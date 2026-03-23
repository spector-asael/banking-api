package data


import (
	"database/sql"
	"time"
	"context"
	"errors"
	"github.com/spector-asael/banking-api/internal/validator"
)



type JournalEntry struct {
	ID              int64     `json:"id"`
	ReferenceTypeID int64     `json:"reference_type_id"`
	ReferenceID     int64     `json:"reference_id"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

type JournalEntryModel struct {
	DB *sql.DB
}

var ErrDuplicateJournalEntry = errors.New("journal entry with this reference already exists")

func (m JournalEntryModel) Insert(entry *JournalEntry) error {
	query := `INSERT INTO journal_entries (reference_type_id, reference_id, description, created_at)
			  VALUES ($1, $2, $3, NOW()) RETURNING id, created_at`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query,
		entry.ReferenceTypeID,
		entry.ReferenceID,
		entry.Description,
	).Scan(&entry.ID, &entry.CreatedAt)
	if err != nil {
		// Add unique constraint error check if needed
		return err
	}
	return nil
}

func (m JournalEntryModel) GetByID(id int64) (*JournalEntry, error) {
	query := `SELECT id, reference_type_id, reference_id, description, created_at FROM journal_entries WHERE id = $1`
	var je JournalEntry
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&je.ID, &je.ReferenceTypeID, &je.ReferenceID, &je.Description, &je.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &je, nil
}

func (m JournalEntryModel) GetAll(filters Filters) ([]*JournalEntry, Metadata, error) {
	query := `SELECT count(*) OVER(), id, reference_type_id, reference_id, description, created_at
			  FROM journal_entries
			  ORDER BY id LIMIT $1 OFFSET $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	var totalRecords int
	entries := []*JournalEntry{}
	for rows.Next() {
		var je JournalEntry
		if err := rows.Scan(&totalRecords, &je.ID, &je.ReferenceTypeID, &je.ReferenceID, &je.Description, &je.CreatedAt); err != nil {
			return nil, Metadata{}, err
		}
		entries = append(entries, &je)
	}
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return entries, metadata, nil
}

func (m JournalEntryModel) Delete(id int64) error {
	query := `DELETE FROM journal_entries WHERE id = $1`
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

func ValidateJournalEntry(v *validator.Validator, je *JournalEntry) {
	v.Check(je.ReferenceTypeID > 0, "reference_type_id", "must be provided and valid")
	v.Check(je.ReferenceID > 0, "reference_id", "must be provided and valid")
	v.Check(je.Description != "", "description", "must be provided")
}