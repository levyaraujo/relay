CREATE TYPE transaction_type AS ENUM ('debit', 'credit');

ALTER TABLE transactions
    ALTER COLUMN type SET DATA TYPE transaction_type
    USING CASE type
        WHEN 1 THEN 'debit'::transaction_type
        WHEN 2 THEN 'credit'::transaction_type
    END;
