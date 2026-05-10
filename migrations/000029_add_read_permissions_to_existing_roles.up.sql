/*-- 0: CUSTOMER
-- Can view their own data and make transactions/loans. No access to persons/employees.
INSERT INTO role_permissions (account_type_id, permission_id)
SELECT 0, id FROM permissions WHERE code IN (
    'accounts:read', 
    'customers:read',
    'transactions:read', 
    'transactions:write', 
    'withdrawals:read', 
    'withdrawals:write', 
    'loans:read', 
    'loans:write'
);

-- 1: TELLER
-- Basically the same system-level permissions as a customer, handling cash/checks and transfers.
-- Later on I'll add a way for customers to not view data of other customers, so this is fine for now.
INSERT INTO role_permissions (account_type_id, permission_id)
SELECT 1, id FROM permissions WHERE code IN (
    'accounts:read', 
    'customers:read',
    'transactions:read', 
    'transactions:write', 
    'withdrawals:read', 
    'withdrawals:write', 
    'loans:read', 
    'loans:write'
);

-- 2: CUSTOMER SERVICE REPRESENTATIVE (CSR)
-- Focuses on account/customer creation and updates. (Added persons/users access to help with KYC/Logins)
INSERT INTO role_permissions (account_type_id, permission_id)
SELECT 2, id FROM permissions WHERE code IN (
    'accounts:read', 
    'accounts:write', 
    'customers:read', 
    'customers:write',
    'persons:read',
    'persons:write',
    'users:read',
    'users:write'
);

-- 3: MANAGER
-- Needs CSR/Teller permissions plus the ability to manage employees.
INSERT INTO role_permissions (account_type_id, permission_id)
SELECT 3, id FROM permissions WHERE code IN (
    'accounts:read', 
    'accounts:write', 
    'customers:read', 
    'customers:write',
    'persons:read',
    'persons:write',
    'users:read',
    'users:write',
    'transactions:read', 
    'transactions:write', 
    'withdrawals:read', 
    'withdrawals:write', 
    'loans:read', 
    'loans:write',
    'employees:read',
    'employees:write'
);

-- 4: ADMIN
-- Full access to everything. Select all IDs.
INSERT INTO role_permissions (account_type_id, permission_id)
SELECT 4, id FROM permissions;
*/