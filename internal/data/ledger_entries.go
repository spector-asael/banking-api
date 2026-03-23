
package data

import (
	"database/sql"
	"time"
	"context"
	"errors"
	"github.com/spector-asael/banking-api/internal/validator"
)

type LedgerEntry struct {
	ID             int64     `json:"id"`
	GLAccountID    int64     `json:"gl_account_id"`
	JournalEntryID int64     `json:"journal_entry_id"`
	Debit          float64   `json:"debit"`
	Credit         float64   `json:"credit"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type LedgerEntryModel struct {
	DB *sql.DB
}

var ErrDuplicateLedgerEntry = errors.New("ledger entry for this journal and GL account already exists")

func (m LedgerEntryModel) Insert(entry *LedgerEntry) error {
	query := `INSERT INTO ledger_entries (gl_account_id, journal_entry_id, debit, credit, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id, created_at, updated_at`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query,
		entry.GLAccountID,
		entry.JournalEntryID,
		entry.Debit,
		entry.Credit,
	).Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)
	if err != nil {
		// Add unique constraint error check if needed
		return err
	}
	return nil
}

func (m LedgerEntryModel) GetByID(id int64) (*LedgerEntry, error) {
	query := `SELECT id, gl_account_id, journal_entry_id, debit, credit, created_at, updated_at FROM ledger_entries WHERE id = $1`
	var le LedgerEntry
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&le.ID, &le.GLAccountID, &le.JournalEntryID, &le.Debit, &le.Credit, &le.CreatedAt, &le.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &le, nil
}

func (m LedgerEntryModel) GetAll(filters Filters) ([]*LedgerEntry, Metadata, error) {
	query := `SELECT count(*) OVER(), id, gl_account_id, journal_entry_id, debit, credit, created_at, updated_at
			  FROM ledger_entries
			  ORDER BY id LIMIT $1 OFFSET $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	var totalRecords int
	entries := []*LedgerEntry{}
	for rows.Next() {
		var le LedgerEntry
		if err := rows.Scan(&totalRecords, &le.ID, &le.GLAccountID, &le.JournalEntryID, &le.Debit, &le.Credit, &le.CreatedAt, &le.UpdatedAt); err != nil {
			return nil, Metadata{}, err
		}
		entries = append(entries, &le)
	}
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return entries, metadata, nil
}

func (m LedgerEntryModel) Delete(id int64) error {
	query := `DELETE FROM ledger_entries WHERE id = $1`
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

func ValidateLedgerEntry(v *validator.Validator, le *LedgerEntry) {
	v.Check(le.GLAccountID > 0, "gl_account_id", "must be provided and valid")
	v.Check(le.JournalEntryID > 0, "journal_entry_id", "must be provided and valid")
	v.Check(le.Debit >= 0, "debit", "must be zero or positive")
	v.Check(le.Credit >= 0, "credit", "must be zero or positive")
	v.Check(!(le.Debit == 0 && le.Credit == 0), "debit_credit", "either debit or credit must be non-zero")
}