DELETE FROM permissions 
WHERE code IN ('comments:read', 'comments:write', 'posts:read', 'posts:write', 'users:read', 'users:write', 
                 'accounts:read', 'accounts:write', 'customers:read', 'customers:write', 
                 'transactions:read', 'transactions:write', 'withdrawals:read', 
                 'withdrawals:write', 'loans:read', 'loans:write');
