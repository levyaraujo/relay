# Relay Business Context

Relay is a conversational ERP that records the parties, catalog items, business operations, obligations, and settlements of an organization.

## Parties

**Party**:
The person or organization that participates in business with the organization. A party may hold one or both of the Customer and Supplier roles.

**Customer**:
A Party that buys products or services from the organization.
_Avoid_: Client as a separate entity

**Supplier**:
A Party that sells products or services to the organization.
_Avoid_: Vendor as a separate entity

## Commercial workflow

**Product/Service**:
A catalog item the organization can buy or sell. The item is either a physical product or a service.

**Transaction**:
A recorded business operation involving the organization and, when applicable, a Party and one or more Product/Service items. A Transaction is the operation header; its itemized breakdown is held by Transaction Lines.
_Avoid_: Payment as a synonym for a transaction

**Transaction Line**:
One itemized Product/Service or descriptive expense entry within a Transaction, including its quantity, unit price, and calculated amount.

**Receivable/Payable**:
An outstanding financial obligation created by a transaction: money expected from a Customer or owed to a Supplier.
_Avoid_: Account as a synonym for a User or Party

**Payment**:
A settlement applied to a Receivable/Payable. A single obligation may have multiple Payments and may remain partially outstanding.
