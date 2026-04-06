ALTER TABLE users
ADD COLUMN employee_id BIGINT,
ADD COLUMN customer_id BIGINT,
ADD COLUMN account_type SMALLINT;

ALTER TABLE users
ADD CONSTRAINT fk_users_employee
FOREIGN KEY (employee_id) REFERENCES employees(id);

ALTER TABLE users
ADD CONSTRAINT fk_users_customer
FOREIGN KEY (customer_id) REFERENCES customers(id);

ALTER TABLE users
ADD CONSTRAINT fk_users_account_type
FOREIGN KEY (account_type) REFERENCES account_types(id);

ALTER TABLE users
ADD CONSTRAINT users_type_match_check
CHECK (
    (account_type = 0 AND customer_id IS NOT NULL AND employee_id IS NULL)
    OR
    (account_type <> 0 AND employee_id IS NOT NULL AND customer_id IS NULL)
);

ALTER TABLE users
ALTER COLUMN account_type SET NOT NULL;

CREATE INDEX idx_users_employee_id ON users(employee_id);
CREATE INDEX idx_users_customer_id ON users(customer_id);
CREATE INDEX idx_users_account_type ON users(account_type);