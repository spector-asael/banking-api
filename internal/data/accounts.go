package data

import (
	"database/sql"
	"time"
	"context"
	"errors"
	"github.com/spector-asael/banking-api/internal/validator"
)

type Account struct {
	ID              int64     `json:"id"`
	AccountNumber   string    `json:"account_number"`
	BranchID        int64     `json:"branch_id_opened_at"`
	AccountTypeID   int64     `json:"account_type_id"`
	GLAccountID     int64     `json:"gl_account_id"`
	Status          string    `json:"status"`
	OpenedAt        time.Time `json:"opened_at"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AccountModel struct {
	DB *sql.DB
}


// Validation for Account
func ValidateAccount(v *validator.Validator, a *Account) {
	v.Check(a.AccountNumber != "", "account_number", "must be provided")
	v.Check(a.BranchID > 0, "branch_id_opened_at", "must be provided and valid")
	v.Check(a.AccountTypeID > 0, "account_type_id", "must be provided and valid")
	v.Check(a.GLAccountID > 0, "gl_account_id", "must be provided and valid")
	v.Check(a.Status != "", "status", "must be provided")
}


var ErrDuplicateAccount = errors.New("account with this account number already exists")
var ErrDuplicateAccountOwnership = errors.New("ownership for this customer and account already exists")

// Account CRUD
func (m AccountModel) Insert(account *Account) error {
	query := `INSERT INTO accounts (account_number, branch_id_opened_at, account_type_id, gl_account_id, status, opened_at, closed_at, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW()) RETURNING id, created_at, updated_at`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query,
		account.AccountNumber,
		account.BranchID,
		account.AccountTypeID,
		account.GLAccountID,
		account.Status,
		account.OpenedAt,
		account.ClosedAt,
	).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if err.Error() != "" && (err.Error() == "UNIQUE constraint failed: accounts.account_number" || err.Error() == "pq: duplicate key value violates unique constraint \"accounts_account_number_key\"") {
			return ErrDuplicateAccount
		}
		return err
	}
	return nil
}

func (m AccountModel) GetByID(id int64) (*Account, error) {
	query := `SELECT id, account_number, branch_id_opened_at, account_type_id, gl_account_id, status, opened_at, closed_at, created_at, updated_at FROM accounts WHERE id = $1`
	var a Account
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.AccountNumber, &a.BranchID, &a.AccountTypeID, &a.GLAccountID, &a.Status, &a.OpenedAt, &a.ClosedAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (m AccountModel) GetAll(filters Filters) ([]*Account, Metadata, error) {
	query := `SELECT count(*) OVER(), id, account_number, branch_id_opened_at, account_type_id, gl_account_id, status, opened_at, closed_at, created_at, updated_at
			  FROM accounts
			  ORDER BY id LIMIT $1 OFFSET $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	var totalRecords int
	accounts := []*Account{}
	for rows.Next() {
		var a Account
		if err := rows.Scan(&totalRecords, &a.ID, &a.AccountNumber, &a.BranchID, &a.AccountTypeID, &a.GLAccountID, &a.Status, &a.OpenedAt, &a.ClosedAt, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, Metadata{}, err
		}
		accounts = append(accounts, &a)
	}
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return accounts, metadata, nil
}

func (m AccountModel) Delete(id int64) error {
	query := `DELETE FROM accounts WHERE id = $1`
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

func (m AccountModel) Update(account *Account) error {
	query := `UPDATE accounts SET status = $1, closed_at = $2, updated_at = NOW() WHERE id = $3 RETURNING updated_at`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return m.DB.QueryRowContext(ctx, query, account.Status, account.ClosedAt, account.ID).Scan(&account.UpdatedAt)
}