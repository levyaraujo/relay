ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS vendor_id UUID REFERENCES vendors (id);