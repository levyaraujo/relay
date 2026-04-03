ALTER TABLE transactions
    ALTER COLUMN type SET DATA TYPE INTEGER
    USING CASE type::text
        WHEN 'debit' THEN 1
        WHEN 'credit' THEN 2
    END;

DROP TYPE transaction_type;
