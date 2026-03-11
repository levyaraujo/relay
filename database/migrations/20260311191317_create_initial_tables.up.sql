CREATE TABLE IF NOT EXISTS companies
(
    id         UUID PRIMARY KEY,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted    BOOLEAN      NOT NULL DEFAULT FALSE,
    name       VARCHAR(255) NOT NULL,
    cnpj       VARCHAR(14)  NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users
(
    id         UUID PRIMARY KEY,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted    BOOLEAN      NOT NULL DEFAULT FALSE,
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,
    company_id UUID         NOT NULL,
    CONSTRAINT fk_company FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS transactions
(
    id          UUID PRIMARY KEY,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted     BOOLEAN        NOT NULL DEFAULT FALSE,
    company_id  UUID           NOT NULL,
    type        INTEGER        NOT NULL,
    amount      NUMERIC(15, 2) NOT NULL,
    description TEXT           NOT NULL,
    due_date    TIMESTAMPTZ,
    paid_date   TIMESTAMPTZ,
    origin      TEXT           NOT NULL,
    creator_id  UUID           NOT NULL,
    CONSTRAINT fk_company FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE RESTRICT,
    CONSTRAINT fk_creator FOREIGN KEY (creator_id) REFERENCES users (id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_users_company_id ON users (company_id);
CREATE INDEX IF NOT EXISTS idx_transactions_company_id ON transactions (company_id);
CREATE INDEX IF NOT EXISTS idx_transactions_creator_id ON transactions (creator_id);