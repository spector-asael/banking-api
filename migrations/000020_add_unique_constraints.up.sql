-- Add unique constraint to person_id in customers table to prevent duplicate customers for the same person
ALTER TABLE customers
ADD CONSTRAINT unique_person_customer UNIQUE (person_id);