CREATE TABLE IF NOT EXISTS branches (
    id SERIAL PRIMARY KEY,
    branch_name TEXT NOT NULL,
    branch_code TEXT NOT NULL UNIQUE,
    address TEXT NOT NULL,
    phone TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);