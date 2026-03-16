CREATE TABLE IF NOT EXISTS persons (
    id SERIAL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    social_security_number varchar(20) NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    date_of_birth DATE NOT NULL,
    phone_number varchar(10) NOT NULL,
    living_address TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);