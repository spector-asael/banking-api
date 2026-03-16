CREATE TABLE IF NOT EXISTS employees (
    id SERIAL PRIMARY KEY,
    person_id INTEGER NOT NULL REFERENCES persons(id),
    branch_id INTEGER NOT NULL REFERENCES branches(id),
    position_id INTEGER NOT NULL REFERENCES position(id),
    hire_date DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);