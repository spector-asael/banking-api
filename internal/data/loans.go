
package data

import (
	"database/sql"
	"time"
	"context"
	"errors"
	"github.com/spector-asael/banking-api/internal/validator"
)

type Loan struct {
	ID              int64     `json:"id"`
	CustomerID      int64     `json:"customer_id"`
	LoanTypeID      int64     `json:"loan_type_id"`
	PrincipalAmount float64   `json:"principal_amount"`
	InterestRate    float64   `json:"interest_rate"`
	TermMonths      int       `json:"term_months"`
	Status          string    `json:"status"`
	IssuedAt        time.Time `json:"issued_at"`
	MaturityDate    time.Time `json:"maturity_date"`
	GLAccountID     int64     `json:"gl_account_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type LoanModel struct {
	DB *sql.DB
}

var ErrDuplicateLoan = errors.New("loan for this customer and type already exists")

func (m LoanModel) Insert(loan *Loan) error {
	query := `INSERT INTO loans (customer_id, loan_type_id, principal_amount, interest_rate, term_months, status, issued_at, maturity_date, gl_account_id, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) RETURNING id, created_at, updated_at`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query,
		loan.CustomerID,
		loan.LoanTypeID,
		loan.PrincipalAmount,
		loan.InterestRate,
		loan.TermMonths,
		loan.Status,
		loan.IssuedAt,
		loan.MaturityDate,
		loan.GLAccountID,
	).Scan(&loan.ID, &loan.CreatedAt, &loan.UpdatedAt)
	if err != nil {
		// Add unique constraint error check if needed
		return err
	}
	return nil
}

func (m LoanModel) GetByID(id int64) (*Loan, error) {
	query := `SELECT id, customer_id, loan_type_id, principal_amount, interest_rate, term_months, status, issued_at, maturity_date, gl_account_id, created_at, updated_at FROM loans WHERE id = $1`
	var l Loan
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&l.ID, &l.CustomerID, &l.LoanTypeID, &l.PrincipalAmount, &l.InterestRate, &l.TermMonths, &l.Status, &l.IssuedAt, &l.MaturityDate, &l.GLAccountID, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &l, nil
}

func (m LoanModel) GetAll(filters Filters) ([]*Loan, Metadata, error) {
	query := `SELECT count(*) OVER(), id, customer_id, loan_type_id, principal_amount, interest_rate, term_months, status, issued_at, maturity_date, gl_account_id, created_at, updated_at
			  FROM loans
			  ORDER BY id LIMIT $1 OFFSET $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	var totalRecords int
	loans := []*Loan{}
	for rows.Next() {
		var l Loan
		if err := rows.Scan(&totalRecords, &l.ID, &l.CustomerID, &l.LoanTypeID, &l.PrincipalAmount, &l.InterestRate, &l.TermMonths, &l.Status, &l.IssuedAt, &l.MaturityDate, &l.GLAccountID, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, Metadata{}, err
		}
		loans = append(loans, &l)
	}
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return loans, metadata, nil
}

func (m LoanModel) Update(loan *Loan) error {
	query := `UPDATE loans SET principal_amount = $1, interest_rate = $2, term_months = $3, status = $4, maturity_date = $5, gl_account_id = $6, updated_at = NOW() WHERE id = $7 RETURNING updated_at`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return m.DB.QueryRowContext(ctx, query,
		loan.PrincipalAmount,
		loan.InterestRate,
		loan.TermMonths,
		loan.Status,
		loan.MaturityDate,
		loan.GLAccountID,
		loan.ID,
	).Scan(&loan.UpdatedAt)
}

func (m LoanModel) Delete(id int64) error {
	query := `DELETE FROM loans WHERE id = $1`
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

func ValidateLoan(v *validator.Validator, l *Loan) {
	v.Check(l.CustomerID > 0, "customer_id", "must be provided and valid")
	v.Check(l.LoanTypeID > 0, "loan_type_id", "must be provided and valid")
	v.Check(l.PrincipalAmount > 0, "principal_amount", "must be greater than zero")
	v.Check(l.InterestRate > 0, "interest_rate", "must be greater than zero")
	v.Check(l.TermMonths > 0, "term_months", "must be greater than zero")
	v.Check(l.Status != "", "status", "must be provided")
	v.Check(!l.IssuedAt.IsZero(), "issued_at", "must be provided")
	v.Check(!l.MaturityDate.IsZero(), "maturity_date", "must be provided")
	v.Check(l.GLAccountID > 0, "gl_account_id", "must be provided and valid")
}