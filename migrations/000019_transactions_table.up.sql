CREATE TABLE account_transactions (
    id SERIAL PRIMARY KEY,

    account_id INTEGER NOT NULL
        REFERENCES accounts(id),

    counterparty_account_id INTEGER
        REFERENCES accounts(id),

    journal_entry_id INTEGER NOT NULL
        REFERENCES journal_entries(id),

    amount NUMERIC(18,2) NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL
);