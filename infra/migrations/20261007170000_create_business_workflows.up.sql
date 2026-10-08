CREATE TYPE party_role AS ENUM (
  'CUSTOMER',
  'SUPPLIER'
);

CREATE TYPE item_kind AS ENUM (
  'PRODUCT',
  'SERVICE'
);

CREATE TYPE transaction_type AS ENUM (
  'SALE',
  'PURCHASE',
  'EXPENSE'
);

CREATE TYPE transaction_status AS ENUM (
  'CONFIRMED',
  'CANCELLED'
);

CREATE TYPE obligation_direction AS ENUM (
  'RECEIVABLE',
  'PAYABLE'
);

CREATE TYPE obligation_status AS ENUM (
  'OPEN',
  'PARTIALLY_PAID',
  'PAID',
  'CANCELLED'
);

CREATE TYPE payment_method AS ENUM (
  'PIX',
  'CASH',
  'BANK_TRANSFER',
  'CARD',
  'OTHER'
);

CREATE TABLE IF NOT EXISTS parties
(
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255),
  phone VARCHAR(255),
  document VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT parties_organization_document_unique UNIQUE (organization_id, document),
  CONSTRAINT parties_id_organization_unique UNIQUE (id, organization_id)
);

CREATE TABLE IF NOT EXISTS party_roles
(
  party_id UUID NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
  role party_role NOT NULL,

  PRIMARY KEY (party_id, role)
);

CREATE TABLE IF NOT EXISTS items
(
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  description VARCHAR(1000),
  kind item_kind NOT NULL,
  default_price_cents BIGINT NOT NULL DEFAULT 0 CHECK (default_price_cents >= 0),
  currency VARCHAR(3) NOT NULL DEFAULT 'BRL',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT items_id_organization_unique UNIQUE (id, organization_id)
);

CREATE TABLE IF NOT EXISTS transactions
(
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  party_id UUID,
  type transaction_type NOT NULL,
  status transaction_status NOT NULL DEFAULT 'CONFIRMED',
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  description VARCHAR(1000),
  currency VARCHAR(3) NOT NULL DEFAULT 'BRL',
  total_cents BIGINT NOT NULL CHECK (total_cents >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT transactions_party_organization_fk
    FOREIGN KEY (party_id, organization_id)
    REFERENCES parties(id, organization_id),
  CONSTRAINT transactions_id_organization_unique UNIQUE (id, organization_id)
);

CREATE TABLE IF NOT EXISTS transaction_lines
(
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
  item_id UUID,
  description VARCHAR(1000),
  quantity BIGINT NOT NULL CHECK (quantity > 0),
  unit_price_cents BIGINT NOT NULL CHECK (unit_price_cents >= 0),
  total_cents BIGINT NOT NULL CHECK (total_cents >= 0),

  CONSTRAINT transaction_lines_item_fk
    FOREIGN KEY (item_id)
    REFERENCES items(id)
);

CREATE TABLE IF NOT EXISTS obligations
(
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  transaction_id UUID NOT NULL UNIQUE REFERENCES transactions(id) ON DELETE RESTRICT,
  party_id UUID,
  direction obligation_direction NOT NULL,
  status obligation_status NOT NULL DEFAULT 'OPEN',
  amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
  paid_amount_cents BIGINT NOT NULL DEFAULT 0 CHECK (paid_amount_cents >= 0),
  currency VARCHAR(3) NOT NULL DEFAULT 'BRL',
  due_date DATE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT obligations_paid_amount_check CHECK (paid_amount_cents <= amount_cents),
  CONSTRAINT obligations_party_organization_fk
    FOREIGN KEY (party_id, organization_id)
    REFERENCES parties(id, organization_id),
  CONSTRAINT obligations_id_organization_unique UNIQUE (id, organization_id)
);

CREATE TABLE IF NOT EXISTS payments
(
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  obligation_id UUID NOT NULL REFERENCES obligations(id) ON DELETE RESTRICT,
  amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
  paid_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  method payment_method NOT NULL,
  reference VARCHAR(255),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT payments_obligation_organization_fk
    FOREIGN KEY (obligation_id, organization_id)
    REFERENCES obligations(id, organization_id)
);

CREATE INDEX IF NOT EXISTS parties_organization_idx ON parties (organization_id);
CREATE INDEX IF NOT EXISTS party_roles_role_idx ON party_roles (role);
CREATE INDEX IF NOT EXISTS items_organization_idx ON items (organization_id);
CREATE INDEX IF NOT EXISTS transactions_organization_idx ON transactions (organization_id);
CREATE INDEX IF NOT EXISTS obligations_organization_status_idx ON obligations (organization_id, status);
CREATE INDEX IF NOT EXISTS obligations_due_date_idx ON obligations (due_date);
CREATE INDEX IF NOT EXISTS payments_obligation_idx ON payments (obligation_id);
