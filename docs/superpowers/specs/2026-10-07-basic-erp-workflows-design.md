# Basic ERP Workflows Design

## Goal

Extend Relay beyond organization onboarding with the core records and workflows needed to operate a small business: shared parties, products/services, business transactions, receivables/payables, and partial payments.

## Context

The repository currently has organization onboarding, standalone user creation, authentication, and an in-progress customer skeleton. The database currently persists organizations and users. This feature adds the first operational ERP slice without inventory tracking, tax calculation, invoicing, or accounting-ledger behavior.

## Domain model

### Party

`Party` is the shared identity for a person or organization that does business with the organization. `Customer` and `Supplier` are roles, not separate entities. A party may have both roles.

Party data is organization-scoped and includes name, email, phone, and document. Roles are stored separately so role assignment is composable and idempotent.

### Item

`Item` is a catalog entry with kind `PRODUCT` or `SERVICE`, name, description, and a default price in integer minor units. It has no inventory quantity or stock state.

### Transaction and transaction lines

`Transaction` is the operation header with type `SALE`, `PURCHASE`, or `EXPENSE`, an optional party, date, description, currency, status, and calculated total. `TransactionLine` is a separate one-to-many table containing the itemized item or descriptive expense entry, quantity, unit price, and calculated amount.

Sales require a party with the Customer role. Purchases require a party with the Supplier role. Expenses may have an optional supplier party and may contain descriptive lines without an item.

### Obligation and payment

Each confirmed transaction creates at most one obligation in this slice. A sale creates a `RECEIVABLE`; a purchase or expense creates a `PAYABLE`. An obligation stores the original amount, due date, direction, currency, and derived status: `OPEN`, `PARTIALLY_PAID`, `PAID`, or `CANCELLED`.

`Payment` is a separate record linked to an obligation. Payments may be partial and may not exceed the outstanding balance. Payment creation recalculates the obligation status atomically.

## Module boundaries

- `party/`: Party entity, role assignment, HTTP handlers, repository, and service.
- `catalog/`: Item entity, HTTP handlers, repository, and service.
- `finance/`: Transaction, TransactionLine, Obligation, and Payment entities; workflow HTTP handlers, repository, and service.
- `organization/`, `user/`, and `auth/`: existing onboarding and authentication flows remain available.
- `shared/`: reusable document normalization, validation, response, and money primitives only.

The existing `customer/` skeleton is adapted into the `party/` module so the repository has one canonical concept for business counterparties.

HTTP request and response DTOs remain separate from domain entities. Request DTOs contain only client-controlled fields, while identifiers, organization ownership, calculated totals, statuses, balances, and timestamps stay under service/repository control. Mapping helpers may reduce repetition, but entities are not decoded directly from write requests or used as unrestricted write responses.

## API contract

Party and catalog endpoints provide organization-scoped create, list, and detail operations:

- `POST /api/v1/parties`
- `GET /api/v1/parties?role=CUSTOMER|SUPPLIER`
- `GET /api/v1/parties/{id}`
- `POST /api/v1/items`
- `GET /api/v1/items`
- `GET /api/v1/items/{id}`

Finance endpoints expose workflow operations rather than independent raw inserts:

- `POST /api/v1/transactions`: creates a transaction, lines, and its obligation; an optional initial payment is supported.
- `GET /api/v1/transactions` and `GET /api/v1/transactions/{id}`: read transaction headers and lines.
- `GET /api/v1/obligations`: filters by receivable/payable direction, status, party, and due date.
- `GET /api/v1/obligations/{id}`: returns the obligation and applied payments.
- `POST /api/v1/obligations/{id}/payments`: creates a full or partial settlement.

Money is represented by `amountCents` and a currency code. Clients provide line inputs; services calculate line totals and transaction totals.

## Persistence and consistency

The migration adds:

- `parties` and `party_roles` with organization-scoped uniqueness and foreign keys.
- `items` with organization-scoped identity and item kind.
- `transactions` and `transaction_lines`.
- `obligations` and `payments`.

Transaction creation inserts the transaction header, lines, obligation, and optional initial payment in one database transaction. Payment creation locks the obligation row, validates the remaining balance, inserts the payment, and updates the derived status in the same database transaction. All repository queries scope access through `organization_id`.

## Error handling

- `400 Bad Request`: malformed or invalid payloads.
- `404 Not Found`: resource does not exist within the current organization.
- `409 Conflict`: duplicate party/item, invalid party role for a transaction, or overpayment.
- `500 Internal Server Error`: persistence failures without leaking database details.

## Testing strategy

Tests will cover:

- Party role creation, duplicate prevention, and organization isolation.
- Product/service validation and catalog reads.
- Transaction type/role rules, line-total calculation, and generated obligation direction.
- Atomic failure behavior for transaction creation.
- Full, partial, zero, negative, and over-payments.
- Derived obligation statuses and concurrent-safe payment behavior at the repository boundary.
- HTTP validation, status codes, and organization scoping.
- Regression coverage for organization onboarding and standalone user creation.

## Out of scope

Inventory quantities and stock movements, tax calculation, invoice documents, recurring transactions, approvals, accounting journal entries, reconciliation, refunds, deletion, and automated reminders are intentionally deferred.
