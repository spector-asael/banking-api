package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/spector-asael/banking-api/internal/validator"
)

type GLAccount struct {
	ID            int64     `json:"id"`
	AccountNumber string    `json:"account_number"`
	Name          string    `json:"name"`
	CategoryID    int64     `json:"category_id"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type GLAccountModel struct {
	DB *sql.DB
}

var ErrDuplicateGLAccount = errors.New("general ledger account with this account number already exists")

// Insert creates a new General Ledger account in the database.
func (m GLAccountModel) Insert(glAccount *GLAccount) error {
	query := `INSERT INTO gl_accounts (account_number, name, category_id, is_active, created_at)
              VALUES ($1, $2, $3, $4, NOW()) RETURNING id, created_at`
              
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, 
		glAccount.AccountNumber, 
		glAccount.Name, 
		glAccount.CategoryID, 
		glAccount.IsActive,
	).Scan(&glAccount.ID, &glAccount.CreatedAt)

	if err != nil {
		if err.Error() != "" && (err.Error() == "UNIQUE constraint failed: gl_accounts.account_number" || err.Error() == "pq: duplicate key value violates unique constraint \"gl_accounts_account_number_key\"") {
			return ErrDuplicateGLAccount
		}
		return err
	}
	return nil
}

// Validation for GL Account
func ValidateGLAccount(v *validator.Validator, gl *GLAccount) {
	v.Check(gl.AccountNumber != "", "account_number", "must be provided")
	v.Check(gl.Name != "", "name", "must be provided")
	v.Check(gl.CategoryID > 0, "category_id", "must be provided and valid")
}

