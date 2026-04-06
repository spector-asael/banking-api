// Filename: internal/data/employees.go
package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/spector-asael/banking-api/internal/validator"
)

type Employee struct {
	ID         int64     `json:"id"`
	PersonID   int64     `json:"person_id"`
	FirstName  string    `json:"first_name"` // Joined from persons
	LastName   string    `json:"last_name"`  // Joined from persons
	BranchID   int64     `json:"branch_id"`
	PositionID int64     `json:"position_id"`
	HireDate   time.Time `json:"hire_date"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type EmployeeModel struct {
	DB *sql.DB
}

var ErrDuplicateEmployee = errors.New("employee record for this person already exists")

// InsertTx handles the insertion within a transaction (useful for multi-step onboarding)
func (m EmployeeModel) InsertTx(tx *sql.Tx, e *Employee) error {
	query := `
        INSERT INTO employees (person_id, branch_id, position_id, hire_date, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
        RETURNING id, created_at, updated_at
    `

	args := []any{e.PersonID, e.BranchID, e.PositionID, e.HireDate, e.Status}

	err := tx.QueryRow(query, args...).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Assuming you have a unique constraint on person_id in the employees table
			if pqErr.Code == "23505" {
				return ErrDuplicateEmployee
			}
		}
		return err
	}

	return nil
}

func (m EmployeeModel) Insert(e *Employee) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := m.InsertTx(tx, e); err != nil {
		return err
	}

	return tx.Commit()
}

func (m EmployeeModel) GetByID(id int64) (*Employee, error) {
	query := `
        SELECT e.id, e.person_id, p.first_name, p.last_name, e.branch_id, e.position_id, e.hire_date, e.status, e.created_at, e.updated_at 
        FROM employees e
        INNER JOIN persons p ON e.person_id = p.id
        WHERE e.id = $1`

	var e Employee
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&e.ID,
		&e.PersonID,
		&e.FirstName,
		&e.LastName,
		&e.BranchID,
		&e.PositionID,
		&e.HireDate,
		&e.Status,
		&e.CreatedAt,
		&e.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &e, nil
}

func (m EmployeeModel) GetAll(filters Filters) ([]*Employee, Metadata, error) {
	query := `
        SELECT count(*) OVER(), e.id, e.person_id, p.first_name, p.last_name, e.branch_id, e.position_id, e.hire_date, e.status, e.created_at, e.updated_at
        FROM employees e
        INNER JOIN persons p ON e.person_id = p.id
        ORDER BY e.id LIMIT $1 OFFSET $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	employees := []*Employee{}

	for rows.Next() {
		var e Employee
		err := rows.Scan(
			&totalRecords,
			&e.ID,
			&e.PersonID,
			&e.FirstName,
			&e.LastName,
			&e.BranchID,
			&e.PositionID,
			&e.HireDate,
			&e.Status,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		employees = append(employees, &e)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return employees, metadata, nil
}

func (m EmployeeModel) Update(e *Employee) error {
	query := `
        UPDATE employees 
        SET branch_id = $1, position_id = $2, status = $3, updated_at = NOW() 
        WHERE id = $4 
        RETURNING updated_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, e.BranchID, e.PositionID, e.Status, e.ID).Scan(&e.UpdatedAt)
}

func (m EmployeeModel) Delete(id int64) error {
	query := `DELETE FROM employees WHERE id = $1`

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

// --- Validation ---

func ValidateEmployee(v *validator.Validator, e *Employee) {
	v.Check(e.PersonID > 0, "person_id", "must be provided and valid")
	v.Check(e.BranchID > 0, "branch_id", "must be provided and valid")
	v.Check(e.PositionID > 0, "position_id", "must be provided and valid")
	v.Check(!e.HireDate.IsZero(), "hire_date", "must be provided")
	v.Check(e.HireDate.Before(time.Now().AddDate(0, 0, 1)), "hire_date", "cannot be in the future")

	// Status Safelist
	v.Check(validator.PermittedValue(e.Status, "active", "inactive", "on leave", "terminated"), "status", "invalid status value")
}
