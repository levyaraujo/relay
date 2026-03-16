ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS account_id UUID REFERENCES accounts (id);