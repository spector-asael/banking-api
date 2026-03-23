-- Remove unique constraint from person_id in customers table (for rollback)
ALTER TABLE customers
DROP CONSTRAINT IF EXISTS unique_person_customer;