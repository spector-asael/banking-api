CREATE TABLE IF NOT EXISTS journal_entries (
    id SERIAL PRIMARY KEY,
    reference_type_id INTEGER NOT NULL REFERENCES journal_reference_types(id),
    reference_id INTEGER NOT NULL, -- e.g., deposit id, loan_payment id
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);