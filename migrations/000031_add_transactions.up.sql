DO $$
DECLARE
    target_account_id INT := 51; -- The ID for cust51
    counter INT := 0;
    new_journal_id INT;
BEGIN
    FOR counter IN 1..200 LOOP
        -- 1. Create the Journal Entry first (the audit trail)
        INSERT INTO journal_entries (reference_type_id, reference_id, description, created_at)
        VALUES (
            1, -- Assuming 1 is 'Standard Transaction'
            target_account_id,
            'Automated payment #' || counter,
            NOW() - (random() * interval '30 days') -- Random date in the last month
        ) RETURNING id INTO new_journal_id;

        -- 2. Create the Transaction linked to that Journal
        INSERT INTO account_transactions (
            account_id, 
            counterparty_account_id, 
            journal_entry_id, 
            amount, 
            created_at
        )
        VALUES (
            target_account_id,
            (SELECT id FROM accounts WHERE id != target_account_id ORDER BY random() LIMIT 1), -- Random recipient
            new_journal_id,
            (random() * 490 + 10)::numeric(10,2), -- Random amount between 10 and 500
            NOW() - (random() * interval '30 days')
        );
    END LOOP;
END $$;