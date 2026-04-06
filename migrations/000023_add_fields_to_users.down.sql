-- 1. Drop indexes
DROP INDEX IF EXISTS idx_users_employee_id;
DROP INDEX IF EXISTS idx_users_customer_id;
DROP INDEX IF EXISTS idx_users_account_type;

-- 2. Drop constraints
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_type_match_check;

ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_employee;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_customer;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_account_type;

-- 3. Drop columns
ALTER TABLE users
DROP COLUMN IF EXISTS employee_id,
DROP COLUMN IF EXISTS customer_id,
DROP COLUMN IF EXISTS account_type;