-- Remove permissions mapped to our core roles (Customer, Teller, CSR, Manager, Admin)
DELETE FROM role_permissions 
WHERE account_type_id IN (0, 1, 2, 3, 4);