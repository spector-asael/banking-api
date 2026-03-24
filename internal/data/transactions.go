package data

import (
    "database/sql"
    "time"

)

type AccountTransaction struct {
    ID                 int64     `json:"id"`
    AccountID          int64     `json:"account_id"`
    CounterpartyAcctID *int64    `json:"counterparty_account_id,omitempty"`
    JournalEntryID     int64     `json:"journal_entry_id"`
    Amount             float64   `json:"amount"`
    CreatedAt          time.Time `json:"created_at"`
}

type AccountTransactionModel struct {
    DB *sql.DB
}

func (m AccountTransactionModel) Insert(tx *AccountTransaction) error {
    query := `INSERT INTO account_transactions (account_id, counterparty_account_id, journal_entry_id, amount, created_at)
              VALUES ($1, $2, $3, $4, NOW()) RETURNING id, created_at`
    var counterparty sql.NullInt64
    if tx.CounterpartyAcctID != nil {
        counterparty = sql.NullInt64{Int64: *tx.CounterpartyAcctID, Valid: true}
    } else {
        counterparty = sql.NullInt64{Valid: false}
    }
    err := m.DB.QueryRow(
        query,
        tx.AccountID,
        counterparty,
        tx.JournalEntryID,
        tx.Amount,
    ).Scan(&tx.ID, &tx.CreatedAt)
    if err != nil {
        return err
    }
    return nil
}

// Fixed: Added EntryType so the rows.Scan has a place to put the 'ledger'/'loan' strings
type TransactionHistory struct {
    Amount      float64   `json:"amount"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    EntryType   string    `json:"entry_type"` 
}