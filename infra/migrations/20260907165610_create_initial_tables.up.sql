CREATE TYPE role as ENUM (
  'OWNER',
  'ADMIN',
);

CREATE TYPE orgtype as ENUM(
  'PRODUCTS',
  'SERVICES'
)


CREATE TABLE IF NOT EXISTS users
(
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  name VARCHAR(255) NOT NULL,
  document VARCHAR(255) NOT NULL UNIQUE,
  phone VARCHAR(255) UNIQUE,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL UNIQUE,
  organization_id UUID NOT NULL,
  role role NOT NULL DEFAULT 'OWNER',

  CONSTRAINT organization_id_fk
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS organizations
(
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  name VARCHAR(255) NOT NULL,
  website VARCHAR(255),
  phone VARCHAR UNIQUE NOT NULL,
  email VARCHAR UNIQUE NOT NULL,
  tax_id VARCHAR UNIQUE NOT NULL,
  currency VARCHAR NOT NULL DEFAULT 'BRL',
  type orgtype NOT NULL,
  description VARCHAR
);
