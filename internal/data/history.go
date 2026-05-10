package data

import (
	"context"
	"database/sql"
	"time"
)

// TransactionHistory represents the result of your joined history query
type TransactionHistory struct {
	Date               time.Time     `json:"date"`
	TransactionDetails string        `json:"transaction_details"`
	Amount             float64       `json:"amount"` // Or use decimal.Decimal for precision
	RecipientID        sql.NullInt64 `json:"recipient_id"`
	AccountNumber      string        `json:"account_number"`
}

type TransactionModel struct {
	DB *sql.DB
}

// GetHistory retrieves the transaction list for a specific account number
func (m TransactionModel) GetHistory(accountNumber string, filters Filters) ([]*TransactionHistory, Metadata, error) {
	// Query includes window function for total records to support your pagination pattern
	query := `
		SELECT 
			count(*) OVER(),
			t.created_at AS date,
			j.description AS transaction_details,
			t.amount,
			t.counterparty_account_id AS recipient_id,
			a.account_number
		FROM account_transactions t
		JOIN journal_entries j ON t.journal_entry_id = j.id
		JOIN accounts a ON t.account_id = a.id
		WHERE a.account_number = $1
		ORDER BY t.created_at DESC
		LIMIT $2 OFFSET $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, accountNumber, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	history := []*TransactionHistory{}

	for rows.Next() {
		var h TransactionHistory
		err := rows.Scan(
			&totalRecords,
			&h.Date,
			&h.TransactionDetails,
			&h.Amount,
			&h.RecipientID,
			&h.AccountNumber,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		history = append(history, &h)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return history, metadata, nil
}
