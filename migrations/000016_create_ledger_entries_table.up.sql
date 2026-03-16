CREATE TABLE IF NOT EXISTS ledger_entries (
    id SERIAL PRIMARY KEY,
    gl_account_id INTEGER NOT NULL REFERENCES gl_accounts(id),
    journal_entry_id INTEGER NOT NULL REFERENCES journal_entries(id),
    debit NUMERIC(18,2) NOT NULL DEFAULT 0,
    credit NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        (debit > 0 AND credit = 0)
        OR
        (credit > 0 AND debit = 0)
    )
);