-- 👤 1. PERSONS (idempotent)
INSERT INTO persons (first_name, last_name, social_security_number, email, date_of_birth, phone_number, living_address)
VALUES
('Alice', 'Admin', 'SSN1001', 'alice.admin@example.com', '1985-05-10', '5016000001', 'San Ignacio'),
('Bob', 'Admin', 'SSN1002', 'bob.admin@example.com', '1980-07-15', '5016000002', 'San Ignacio')
ON CONFLICT (social_security_number) DO NOTHING;


INSERT INTO persons (first_name, last_name, social_security_number, email, date_of_birth, phone_number, living_address)
SELECT 
    'Customer' || i,
    'User' || i,
    'SSN_CUST_' || i,
    'customer' || i || '@mail.com',
    DATE '1990-01-01' + (i || ' days')::interval,
    '50161' || LPAD(i::text, 5, '0'),
    'Belize City'
FROM generate_series(1, 50) i
ON CONFLICT (social_security_number) DO NOTHING;


-- 👨‍💼 2. EMPLOYEES
INSERT INTO employees (person_id, branch_id, position_id, hire_date)
SELECT id, 1, 4, '2020-01-01'
FROM persons
WHERE social_security_number IN ('SSN1001','SSN1002')
ON CONFLICT DO NOTHING;


INSERT INTO employees (person_id, branch_id, position_id, hire_date)
SELECT 
    id,
    1,
    (RANDOM()*2 + 1)::int,
    CURRENT_DATE - (RANDOM()*1000)::int
FROM persons
WHERE id > 2
LIMIT 10
ON CONFLICT DO NOTHING;


-- 👥 3. CUSTOMERS
INSERT INTO customers (person_id, kyc_status_id)
SELECT id, 2
FROM persons
WHERE id > 2
ON CONFLICT (person_id) DO NOTHING;


-- 🏦 4. ACCOUNTS
INSERT INTO accounts (account_number, branch_id_opened_at, account_type_id, gl_account_id)
SELECT 
    'ACC' || c.id,
    1,
    1,
    2
FROM customers c
ON CONFLICT (account_number) DO NOTHING;


-- 🔗 5. ACCOUNT OWNERSHIP (FIXED JOIN)
INSERT INTO account_ownerships (customer_id, account_id)
SELECT 
    c.id,
    a.id
FROM customers c
JOIN accounts a 
    ON a.account_number = 'ACC' || c.id
ON CONFLICT DO NOTHING;


-- 🔐 6. USERS (FIXED PASSWORD)

-- Admins
INSERT INTO users (username, email, password_hash, activated, employee_id, account_type)
SELECT 
    'admin' || e.id,
    'admin' || e.id || '@bank.com',
    convert_to('password123', 'UTF8'),
    true,
    e.id,
    4
FROM employees e
WHERE e.position_id = 4
ON CONFLICT (email) DO NOTHING;


-- Employees
INSERT INTO users (username, email, password_hash, activated, employee_id, account_type)
SELECT
    'emp' || id,
    'emp' || id || '@bank.com',
    convert_to('password123', 'UTF8'),
    true,
    id,
    (RANDOM()*2 + 1)::int
FROM employees
WHERE id NOT IN (
    SELECT employee_id FROM users WHERE employee_id IS NOT NULL
)
ON CONFLICT (email) DO NOTHING;


-- Customers
INSERT INTO users (username, email, password_hash, activated, customer_id, account_type)
SELECT
    'cust' || id,
    'cust' || id || '@bank.com',
    convert_to('password123', 'UTF8'),
    true,
    id,
    0
FROM customers
WHERE id NOT IN (
    SELECT customer_id FROM users WHERE customer_id IS NOT NULL
)
ON CONFLICT (email) DO NOTHING;


-- 📒 7. JOURNAL ENTRIES
INSERT INTO journal_entries (reference_type_id, reference_id, description)
SELECT 1, id, 'Initial deposit'
FROM accounts
ON CONFLICT DO NOTHING;


-- 💰 8. LEDGER ENTRIES
INSERT INTO ledger_entries (gl_account_id, journal_entry_id, debit, credit)
SELECT 
    1,
    je.id,
    1000,
    0
FROM journal_entries je
LEFT JOIN ledger_entries le ON le.journal_entry_id = je.id
WHERE le.id IS NULL;