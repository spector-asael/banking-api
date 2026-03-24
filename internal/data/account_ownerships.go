package data 

import (
	"time"
	"database/sql"
	"github.com/spector-asael/banking-api/internal/validator"
	"context"
)

type AccountOwnership struct {
	ID             int64     `json:"id"`
	CustomerID     int64     `json:"customer_id"`
	AccountID      int64     `json:"account_id"`
	IsJointAccount bool      `json:"is_joint_account"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}


type AccountOwnershipModel struct {
	DB *sql.DB
}

// Validation for AccountOwnership
func ValidateAccountOwnership(v *validator.Validator, ao *AccountOwnership) {
	v.Check(ao.CustomerID > 0, "customer_id", "must be provided and valid")
	v.Check(ao.AccountID > 0, "account_id", "must be provided and valid")
}


// AccountOwnership CRUD
func (m AccountOwnershipModel) Insert(ao *AccountOwnership) error {
	query := `INSERT INTO account_ownerships (customer_id, account_id, is_joint_account, created_at, updated_at)
			  VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id, created_at, updated_at`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, ao.CustomerID, ao.AccountID, ao.IsJointAccount).Scan(&ao.ID, &ao.CreatedAt, &ao.UpdatedAt)
	if err != nil {
		if err.Error() != "" && (err.Error() == "UNIQUE constraint failed: account_ownerships.customer_id, account_ownerships.account_id" || err.Error() == "pq: duplicate key value violates unique constraint \"account_ownerships_customer_id_account_id_key\"") {
			return ErrDuplicateAccountOwnership
		}
		return err
	}
	return nil
}

func (m AccountOwnershipModel) GetAllByAccount(accountID int64) ([]*AccountOwnership, error) {
	query := `SELECT id, customer_id, account_id, is_joint_account, created_at, updated_at FROM account_ownerships WHERE account_id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ownerships []*AccountOwnership
	for rows.Next() {
		var ao AccountOwnership
		if err := rows.Scan(&ao.ID, &ao.CustomerID, &ao.AccountID, &ao.IsJointAccount, &ao.CreatedAt, &ao.UpdatedAt); err != nil {
			return nil, err
		}
		ownerships = append(ownerships, &ao)
	}
	return ownerships, nil
}