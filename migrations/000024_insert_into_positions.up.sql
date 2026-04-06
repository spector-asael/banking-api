-- This will overwrite 2 and 3, and create 4. (1 stays the same).
INSERT INTO position (id, position_name) 
VALUES 
    (1, 'Teller'),
    (2, 'Customer Service Representative'),
    (3, 'Manager'),
    (4, 'Administrator')
ON CONFLICT (id) DO UPDATE 
SET position_name = EXCLUDED.position_name;