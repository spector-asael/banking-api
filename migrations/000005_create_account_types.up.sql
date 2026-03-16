CREATE TABLE IF NOT EXISTS account_types (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    interest_rate NUMERIC(5,2) DEFAULT 0.00,
    minimum_balance NUMERIC(18,2) DEFAULT 0.00,
    monthly_fee NUMERIC(18,2) DEFAULT 0.00,
    overdraft_allowed BOOLEAN NOT NULL DEFAULT FALSE,
    withdrawal_limit INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);