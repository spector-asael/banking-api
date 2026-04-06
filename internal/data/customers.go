package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/spector-asael/banking-api/internal/validator"
)

type Customer struct {
	ID          int64     `json:"id"`
	PersonID    int64     `json:"person_id"`
	FirstName   string    `json:"first_name"` // Added
	LastName    string    `json:"last_name"`  // Added
	KYCStatusID int64     `json:"kyc_status_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CustomerModel struct {
	DB *sql.DB
}

var ErrDuplicateCustomer = errors.New("customer for this person already exists")

func (m CustomerModel) InsertTx(tx *sql.Tx, customer *Customer) error {
	query := `
		INSERT INTO customers (person_id, kyc_status_id, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := tx.QueryRow(
		query,
		customer.PersonID,
		customer.KYCStatusID,
	).Scan(&customer.ID, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" && pqErr.Constraint == "unique_person_customer" {
				return ErrDuplicateCustomer
			}
		}
		return err
	}

	return nil
}

func (m CustomerModel) Insert(customer *Customer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = m.InsertTx(tx, customer)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m CustomerModel) GetByID(id int64) (*Customer, error) {
	// 1. Updated query to JOIN the persons table and SELECT first_name and last_name
	query := `
        SELECT c.id, c.person_id, p.first_name, p.last_name, c.kyc_status_id, c.created_at, c.updated_at 
        FROM customers c
        INNER JOIN persons p ON c.person_id = p.id
        WHERE c.id = $1`

	var c Customer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 2. Updated Scan to grab all 7 fields in the exact order of the SELECT statement
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		&c.PersonID,
		&c.FirstName, // Added
		&c.LastName,  // Added
		&c.KYCStatusID,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &c, nil
}

func (m CustomerModel) GetAll(filters Filters) ([]*Customer, Metadata, error) {
	// Note: Added c.updated_at to the end of the SELECT
	query := `SELECT count(*) OVER(), c.id, c.person_id, p.first_name, p.last_name, c.kyc_status_id, c.created_at, c.updated_at
              FROM customers c
              INNER JOIN persons p ON c.person_id = p.id
              ORDER BY c.id LIMIT $1 OFFSET $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var totalRecords int
	customers := []*Customer{}

	for rows.Next() {
		var c Customer
		// Here is the fix: Scan now maps perfectly to the 8 columns from the query above
		if err := rows.Scan(
			&totalRecords,
			&c.ID,
			&c.PersonID,
			&c.FirstName,
			&c.LastName,
			&c.KYCStatusID,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, Metadata{}, err
		}
		customers = append(customers, &c)
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return customers, metadata, nil
}

func (m CustomerModel) Update(customer *Customer) error {
	query := `UPDATE customers SET kyc_status_id = $1, updated_at = NOW() WHERE id = $2 RETURNING updated_at`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return m.DB.QueryRowContext(ctx, query, customer.KYCStatusID, customer.ID).Scan(&customer.UpdatedAt)
}

func (m CustomerModel) Delete(id int64) error {
	query := `DELETE FROM customers WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// Validation for Customer
func ValidateCustomer(v *validator.Validator, c *Customer) {
	v.Check(c.PersonID > 0, "person_id", "must be provided and valid")
	v.Check(c.KYCStatusID > 0, "kyc_status_id", "must be provided and valid")
}

// For updating KYC status
type UpdateKYCStatusInput struct {
	KYCStatusID int64 `json:"kyc_status_id"`
}

func ValidateUpdateKYCStatus(v *validator.Validator, input *UpdateKYCStatusInput) {
	v.Check(input.KYCStatusID > 0, "kyc_status_id", "must be provided and valid")
}
