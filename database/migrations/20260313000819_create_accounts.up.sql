CREATE TABLE IF NOT EXISTS accounts
(
    id         UUID PRIMARY KEY,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted    BOOLEAN      NOT NULL DEFAULT FALSE,
    company_id UUID         NOT NULL,
    code       VARCHAR(255) NOT NULL,
    name       VARCHAR(255) NOT NULL,
    type       INTEGER      NOT NULL,
    parent_id  UUID,
    CONSTRAINT fk_company FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE RESTRICT,
    CONSTRAINT fk_parent FOREIGN KEY (parent_id) REFERENCES accounts (id) ON DELETE RESTRICT
)