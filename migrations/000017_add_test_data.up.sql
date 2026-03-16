INSERT INTO gl_account_categories (name, description) VALUES
('Asset', 'Bank owned resources'),
('Liability', 'Customer deposits'),
('Income', 'Revenue accounts'),
('Expense', 'Operational costs');

INSERT INTO gl_accounts (account_number, name, category_id) VALUES
('1000', 'Cash on Hand', 1),
('2000', 'Customer Deposits Liability', 2),
('4000', 'Interest Income', 3),
('5000', 'Operating Expenses', 4);

INSERT INTO branches (branch_name, branch_code, address, phone) VALUES
('Main Branch', 'MB001', '123 Main St, New York, NY', '2125551111'),
('Downtown Branch', 'DT002', '456 Wall St, New York, NY', '2125552222');

INSERT INTO position (position_name) VALUES
('Teller'),
('Branch Manager'),
('Loan Officer');

INSERT INTO kyc_statuses (name, description) VALUES
('Pending', 'KYC under review'),
('Approved', 'KYC verified'),
('Rejected', 'KYC rejected');

INSERT INTO account_types 
(name, description, interest_rate, minimum_balance, monthly_fee, overdraft_allowed, withdrawal_limit)
VALUES
('Checking', 'Standard checking account', 0.00, 0.00, 10.00, TRUE, NULL),
('Savings', 'Interest bearing savings', 1.50, 100.00, 0.00, FALSE, 6);

INSERT INTO loan_types (name, category, description) VALUES
('Personal Loan', 'Consumer', 'Unsecured personal loan'),
('Mortgage', 'Real Estate', 'Home mortgage loan'),
('Auto Loan', 'Vehicle', 'Car financing');

INSERT INTO journal_reference_types (name, description) VALUES
('Deposit', 'Customer deposit transaction'),
('Withdrawal', 'Customer withdrawal transaction'),
('Loan Disbursement', 'Loan issued to customer');

INSERT INTO persons 
(first_name, last_name, social_security_number, email, date_of_birth, phone_number, living_address)
VALUES
('John', 'Doe', '123-45-6789', 'john.doe@email.com', '1990-05-12', '9175551234', '12 Broadway, NY'),
('Jane', 'Smith', '987-65-4321', 'jane.smith@email.com', '1985-09-21', '9175555678', '99 Madison Ave, NY');

INSERT INTO customers (person_id, kyc_status_id)
VALUES
(1, 2), -- John Approved
(2, 1); -- Jane Pending

INSERT INTO accounts
(account_number, branch_id_opened_at, account_type_id, gl_account_id)
VALUES
('CHK10001', 1, 1, 2),
('SAV20001', 1, 2, 2);

INSERT INTO account_ownerships (customer_id, account_id, is_joint_account)
VALUES
(1, 1, FALSE),
(2, 2, FALSE);