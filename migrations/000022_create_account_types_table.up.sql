CREATE TABLE IF NOT EXISTS account_types (
    id SMALLINT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- Customer accounts
INSERT INTO account_types (id, name) VALUES (0, 'CUSTOMER');

-- Employee accounts
-- Can deposit/withdraw money. Can transfer funds between accounts.
-- Handle cash and checks.
INSERT INTO account_types (id, name) VALUES (1, 'TELLER');

-- Create customer accounts, update customer info
INSERT INTO account_types (id, name) VALUES (2, 'CUSTOMER SERVICE REPRESENTATIVE');

-- Approve large transactions, manage employee accounts
INSERT INTO account_types (id, name) VALUES (3, 'MANAGER');

-- Full access to all system features and data
INSERT INTO account_types (id, name) VALUES (4, 'ADMIN');