# Basic ERP Workflows Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add organization-scoped Party, catalog Item, Transaction, Obligation, and Payment workflows without inventory tracking.

**Architecture:** Adapt the existing customer skeleton into a `party` module, add a `catalog` module, and put transaction/obligation/payment orchestration behind a deep `finance` module. HTTP DTOs stay separate from domain entities; workflow repositories own atomic database writes.

**Tech Stack:** Go 1.27, `net/http`, `database/sql` with PostgreSQL/pgx, validator/v10, existing UUID and password helpers, SQL migrations.

**Spec:** `docs/superpowers/specs/2026-10-07-basic-erp-workflows-design.md`

## Global Constraints

- Customer and Supplier are roles on one Party identity; they are not separate entities.
- Product/Service has no inventory quantity or stock state.
- Money is represented by integer `amountCents` values and a currency code.
- Request/response DTOs remain separate from domain entities.
- All new reads and writes are scoped through `organization_id`.
- Transaction creation and payment creation are atomic database workflows.
- Existing organization onboarding and standalone `/api/v1/users` creation remain available.
- Do not add tax, invoice, recurring transaction, approval, accounting-ledger, reconciliation, refund, deletion, or reminder workflows.

## Review Focus

- A resource ID from another organization must return `404` and never leak data; cover in Party, Catalog, and Finance HTTP tests.
- A transaction with the wrong Party role must return `409` and create no financial records; cover in Finance service/repository tests.
- Client-supplied totals, statuses, balances, or organization IDs must be ignored because services calculate/control them; cover in DTO/handler tests.
- A partial payment must update the obligation to `PARTIALLY_PAID`, while zero, negative, and overpayments must fail; cover in Finance payment tests.
- Two payments racing against the same balance must not overpay the obligation; cover at the repository boundary with a locked-obligation integration test.

---

### Task 1: Add authenticated organization context and workflow migration

**Files:**
- Create: `infra/migrations/20261007170000_create_business_workflows.up.sql`
- Create: `infra/migrations/20261007170000_create_business_workflows.down.sql`
- Modify: `auth/http.go`
- Modify: `user/repo.go`
- Modify: `main.go`
- Test: `auth/http_test.go`, `user/repo_test.go` as needed for the new seam

**Interfaces:**
- Produces `auth.Middleware(next http.Handler, repo user.Repo) http.Handler`.
- Produces an organization ID in request context for protected workflow handlers.
- Produces `user.Repo.FindByID(id uuid.UUID) (User, error)` or an equivalent organization lookup.

- [ ] **Step 1: Write failing tests** for middleware context population and rejection of missing/invalid bearer tokens. Assert a valid user request carries the user’s `OrganizationID`, while a user without an organization is unauthorized for organization-scoped workflows.
- [ ] **Step 2: Run `go test ./auth ./user`** and confirm the tests fail because the exported middleware/context lookup and repository lookup do not exist.
- [ ] **Step 3: Implement the auth seam** without changing login behavior. Preserve the existing JWT validation, expose only the middleware/context access needed by protected handlers, and add the repository lookup with organization scoping.
- [ ] **Step 4: Write the migration** with organization-scoped tables and constraints for `parties`, `party_roles`, `items`, `transactions`, `transaction_lines`, `obligations`, and `payments`; add enum/check values for party roles, item kinds, transaction types/statuses, obligation directions/statuses, and payment methods. Add indexes for organization, role, due date, and obligation status.
- [ ] **Step 5: Run `go test ./auth ./user ./...`** and verify all existing tests pass.
- [ ] **Step 6: Commit** the auth context and migration separately from domain modules.

### Task 2: Implement the Party module

**Files:**
- Create: `party/entity.go`
- Create: `party/http.go`
- Create: `party/repo.go`
- Create: `party/services.go`
- Create: `party/party_test.go`
- Delete after Party parity tests pass: `customer/entity.go`, `customer/services.go`
- Modify: `main.go`

**Interfaces:**
- `type Role string` with `CustomerRole` and `SupplierRole`.
- `type Party struct { ID, OrganizationID uuid.UUID; Name, Email, Phone, Document string; Roles []Role }`.
- `type CreatePartyRequest struct { Name, Email, Phone, Document string; Roles []Role }`.
- `type repository interface { Create(uuid.UUID, CreatePartyRequest) (Party, error); List(uuid.UUID, *Role) ([]Party, error); FindByID(uuid.UUID, uuid.UUID) (Party, error) }`.
- `type Service` methods: `Create`, `List`, and `FindByID` with organization ID supplied by the handler context.
- Routes: `POST /api/v1/parties`, `GET /api/v1/parties`, `GET /api/v1/parties/{id}`.

- [ ] **Step 1: Write failing service tests** for document normalization/validation, duplicate role idempotency, invalid roles, and creation of one Party with both roles. Assert organization ID comes from the service argument, not the request DTO.
- [ ] **Step 2: Run `go test ./party`** and verify failure because the Party module does not exist.
- [ ] **Step 3: Implement the Party entity/service/repository**. Keep request DTOs separate; normalize documents before persistence; use one insert transaction for the Party and role rows; map duplicate constraints to a domain conflict error. Preserve the current customer behavior in the new package before deleting the old `customer/` files.
- [ ] **Step 4: Write failing handler tests** for create/list/detail, validation errors, missing resources, and cross-organization IDs returning `404`.
- [ ] **Step 5: Implement handlers** using authenticated organization context and shared JSON/validation responses.
- [ ] **Step 6: Register protected Party routes** behind `auth.Middleware` in `main.go`.
- [ ] **Step 7: Run `go test ./party ./...`** and commit the Party module.

### Task 3: Implement the catalog Item module

**Files:**
- Create: `catalog/entity.go`
- Create: `catalog/http.go`
- Create: `catalog/repo.go`
- Create: `catalog/services.go`
- Create: `catalog/catalog_test.go`
- Modify: `main.go`

**Interfaces:**
- `type Kind string` with `ProductKind` and `ServiceKind`.
- `type Item struct { ID, OrganizationID uuid.UUID; Name, Description string; Kind Kind; DefaultPriceCents int64; Currency string }`.
- `type CreateItemRequest struct { Name, Description string; Kind Kind; DefaultPriceCents int64; Currency string }`.
- `type repository interface { Create(uuid.UUID, CreateItemRequest) (Item, error); List(uuid.UUID) ([]Item, error); FindByID(uuid.UUID, uuid.UUID) (Item, error) }`.
- Routes: `POST /api/v1/items`, `GET /api/v1/items`, `GET /api/v1/items/{id}`.

- [ ] **Step 1: Write failing service tests** for valid PRODUCT/SERVICE kinds, rejection of invalid kinds and negative prices, and defaulting an omitted currency to the organization currency (`BRL` for the current organization model).
- [ ] **Step 2: Run `go test ./catalog`** and verify failure.
- [ ] **Step 3: Implement entity, DTOs, service, and repository** with organization scoping and uniqueness constraints.
- [ ] **Step 4: Write failing handler tests** for create/list/detail, invalid payloads, and cross-organization item IDs.
- [ ] **Step 5: Implement handlers and register protected routes.**
- [ ] **Step 6: Run `go test ./catalog ./...`** and commit the catalog module.

### Task 4: Implement Finance transaction creation

**Files:**
- Create: `finance/entity.go`
- Create: `finance/http.go`
- Create: `finance/repo.go`
- Create: `finance/services.go`
- Create: `finance/finance_test.go`
- Modify: `main.go`

**Interfaces:**
- `type TransactionType string` with `Sale`, `Purchase`, and `Expense`.
- `type TransactionStatus string` with `Confirmed` and `Cancelled`.
- `type ObligationDirection string` with `Receivable` and `Payable`.
- `type ObligationStatus string` with `Open`, `PartiallyPaid`, `Paid`, and `Cancelled`.
- `type CreateTransactionRequest struct { Type TransactionType; PartyID *uuid.UUID; OccurredAt time.Time; DueDate *time.Time; Description, Currency string; Lines []CreateTransactionLineRequest; InitialPayment *CreatePaymentRequest }`.
- `type CreateTransactionLineRequest struct { ItemID *uuid.UUID; Description string; Quantity int64; UnitPriceCents int64 }`.
- `type CreatePaymentRequest struct { AmountCents int64; PaidAt time.Time; Method PaymentMethod; Reference string }`.
- `type repository interface { CreateTransaction(uuid.UUID, TransactionDraft) (TransactionDetail, error); ListTransactions(uuid.UUID, TransactionFilter) ([]TransactionDetail, error); FindTransaction(uuid.UUID, uuid.UUID) (TransactionDetail, error) }`.
- Service method: `CreateTransaction(organizationID uuid.UUID, request CreateTransactionRequest) (TransactionDetail, error)`.
- Routes: `POST /api/v1/transactions`, `GET /api/v1/transactions`, `GET /api/v1/transactions/{id}`.

- [ ] **Step 1: Write failing service tests** for line total calculation, header total calculation, required Customer/Supplier roles, optional Party for expenses, required items for sales/purchases, descriptive expense lines, and obligation direction.
- [ ] **Step 2: Run `go test ./finance`** and verify failure.
- [ ] **Step 3: Implement domain calculations and validation**. Ignore any client-provided totals/statuses/organization IDs. Reject empty lines, non-positive quantities, negative unit prices, and invalid currencies/types.
- [ ] **Step 4: Write failing repository tests** proving transaction header, lines, obligation, and optional initial payment are committed together and rolled back together. Add cross-organization lookup tests.
- [ ] **Step 5: Implement the atomic repository workflow** using one SQL transaction. Generate `RECEIVABLE` for sales and `PAYABLE` for purchases/expenses; derive initial obligation status from the optional initial payment.
- [ ] **Step 6: Write failing handler tests** for request validation, `201 Created`, `400` validation, `404` resource lookup, `409` role conflict, and ignoring client-controlled computed fields.
- [ ] **Step 7: Implement handlers, register protected routes, run `go test ./finance ./...`, and commit the transaction workflow.**

### Task 5: Implement obligation reads and payment settlement

**Files:**
- Modify: `finance/entity.go`
- Modify: `finance/http.go`
- Modify: `finance/repo.go`
- Modify: `finance/services.go`
- Modify: `finance/finance_test.go`

**Interfaces:**
- `type PaymentMethod string` with `PIX`, `CASH`, `BANK_TRANSFER`, `CARD`, and `OTHER`.
- `type Payment struct { ID, ObligationID, OrganizationID uuid.UUID; AmountCents int64; PaidAt time.Time; Method PaymentMethod; Reference string }`.
- `type repository interface { ListObligations(uuid.UUID, ObligationFilter) ([]ObligationDetail, error); FindObligation(uuid.UUID, uuid.UUID) (ObligationDetail, error); CreatePayment(uuid.UUID, uuid.UUID, PaymentDraft) (ObligationDetail, error) }`.
- Service method: `CreatePayment(organizationID, obligationID uuid.UUID, request CreatePaymentRequest) (ObligationDetail, error)`.
- Routes: `GET /api/v1/obligations`, `GET /api/v1/obligations/{id}`, `POST /api/v1/obligations/{id}/payments`.

- [ ] **Step 1: Write failing service tests** for full payment → `PAID`, partial payment → `PARTIALLY_PAID`, zero/negative amount rejection, overpayment rejection, invalid method rejection, and payment organization scoping.
- [ ] **Step 2: Run `go test ./finance`** and verify failure.
- [ ] **Step 3: Implement service validation and status calculation** from original amount minus applied payments; never accept a client-provided balance or status.
- [ ] **Step 4: Write failing repository tests** for `SELECT ... FOR UPDATE`, payment insert, obligation status update, rollback on insert/update failure, and two concurrent payment attempts against one remaining balance.
- [ ] **Step 5: Implement the atomic payment repository workflow** with row locking and organization-scoped lookup.
- [ ] **Step 6: Implement obligation/payment handlers and list/detail filters** for direction, status, party, and due date.
- [ ] **Step 7: Run `go test ./finance ./...` and commit the payment workflow.**

### Task 6: Integrate, migrate, and verify the complete slice

**Files:**
- Modify: `main.go`
- Modify: existing organization/user tests only where regression coverage needs shared helpers
- Test: end-to-end HTTP tests under `party/`, `catalog/`, and `finance/`

- [ ] **Step 1: Add an end-to-end workflow test** that creates one Party with Customer/Supplier roles, creates a Service item, creates a sale with two transaction lines, verifies the Receivable, applies a partial payment, and verifies the remaining balance/status.
- [ ] **Step 2: Add a purchase/expense workflow test** verifying Supplier role enforcement and Payable direction.
- [ ] **Step 3: Apply the migration to a disposable PostgreSQL database** and run repository integration tests against it.
- [ ] **Step 4: Run `go test ./...`, `go vet ./...`, and `git diff --check`.
- [ ] **Step 5: Verify existing organization onboarding and `/api/v1/users` creation still work.
- [ ] **Step 6: Commit the integration and verification changes.**

## Final Verification

Run from the repository root:

```bash
go test ./...
go vet ./...
git diff --check
```

Expected result: all Go packages pass, vet reports no diagnostics, and the diff has no whitespace errors.
