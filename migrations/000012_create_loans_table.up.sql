CREATE TABLE IF NOT EXISTS loans (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES customers(id),
    loan_type_id INTEGER NOT NULL REFERENCES loan_types(id),
    principal_amount NUMERIC(18,2) NOT NULL,
    interest_rate NUMERIC(5,2) NOT NULL, -- e.g., 3.5%
    term_months INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    maturity_date TIMESTAMPTZ NOT NULL,
    gl_account_id INTEGER NOT NULL REFERENCES gl_accounts(id), -- asset account
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);