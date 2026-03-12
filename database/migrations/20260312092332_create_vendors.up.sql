CREATE TABLE IF NOT EXISTS vendors
(
    id            UUID PRIMARY KEY,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted       BOOLEAN      NOT NULL DEFAULT FALSE,
    company_id    UUID         NOT NULL,
    name          VARCHAR(255) NOT NULL,
    cnpj          VARCHAR(14),
    email         VARCHAR(255),
    phone         VARCHAR(20),
    payment_terms INTEGER      NOT NULL DEFAULT 0,
    CONSTRAINT fk_vendor_company FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_vendors_company_id ON vendors (company_id);
