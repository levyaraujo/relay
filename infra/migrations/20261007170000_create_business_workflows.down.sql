DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS obligations;
DROP TABLE IF EXISTS transaction_lines;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS party_roles;
DROP TABLE IF EXISTS parties;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS outbox_events;

DROP TYPE IF EXISTS payment_method;
DROP TYPE IF EXISTS obligation_status;
DROP TYPE IF EXISTS obligation_direction;
DROP TYPE IF EXISTS transaction_status;
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS item_kind;
DROP TYPE IF EXISTS party_role;
