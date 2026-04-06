-- 1. Remove the Administrator role that was created in the UP migration
DELETE FROM position WHERE id = 4;

-- 2. Revert IDs 2 and 3 back to their original titles
INSERT INTO position (id, position_name) 
VALUES 
    (1, 'Teller'),
    (2, 'Branch Manager'),
    (3, 'Loan Officer')
ON CONFLICT (id) DO UPDATE 
SET position_name = EXCLUDED.position_name;