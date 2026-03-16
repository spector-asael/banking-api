CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    account_number TEXT NOT NULL UNIQUE,
    branch_id_opened_at INTEGER NOT NULL REFERENCES branches(id),
    account_type_id INTEGER NOT NULL REFERENCES account_types(id),
    gl_account_id INTEGER NOT NULL REFERENCES gl_accounts(id), -- liability account
    status TEXT NOT NULL DEFAULT 'active',
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);