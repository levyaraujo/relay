#!/usr/bin/env bash

# Requires a running Relay API with migrations applied, plus curl and jq.
# Usage: ./scripts/create-basic-organization.sh
# Override the default API URL or dataset size with environment variables:
# BASE_URL=http://localhost:8080/api/v1 SALE_COUNT=5000 ./scripts/create-basic-organization.sh

set -Eeuo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
BASE_URL="${BASE_URL%/}"

CUSTOMER_COUNT="${CUSTOMER_COUNT:-24}"
SUPPLIER_COUNT="${SUPPLIER_COUNT:-24}"
SERVICE_COUNT="${SERVICE_COUNT:-16}"
PRODUCT_COUNT="${PRODUCT_COUNT:-16}"
SALE_COUNT="${SALE_COUNT:-2000}"
PURCHASE_COUNT="${PURCHASE_COUNT:-2000}"
EXPENSE_COUNT="${EXPENSE_COUNT:-500}"

for command in curl jq; do
	if ! command -v "$command" >/dev/null 2>&1; then
		echo "Missing required command: $command" >&2
		exit 1
	fi
done

RUN_ID="$(date +%s)-${RANDOM}"
RUN_SEED="$(date +%s)${RANDOM}"
OWNER_EMAIL="owner-${RUN_ID}@relay.local"

declare -a CUSTOMER_IDS=()
declare -a SUPPLIER_IDS=()
declare -a SERVICE_IDS=()
declare -a PRODUCT_IDS=()

cnpj_digit() {
	local digits="$1"
	local weights="$2"
	local sum=0
	local i
	local value
	local remainder

	for ((i = 0; i < ${#digits}; i++)); do
		value="${digits:i:1}"
		sum=$((sum + value * ${weights:i:1}))
	done

	remainder=$((sum % 11))
	if ((remainder < 2)); then
		echo 0
	else
		echo $((11 - remainder))
	fi
}

new_cnpj() {
	local offset="$1"
	local base
	local digit_one
	local digit_two

	base="$(printf '%012d' $(((RUN_SEED + offset * 7919) % 1000000000000)))"
	digit_one="$(cnpj_digit "$base" 543298765432)"
	digit_two="$(cnpj_digit "${base}${digit_one}" 6543298765432)"
	echo "${base}${digit_one}${digit_two}"
}

cpf_digit() {
	local digits="$1"
	local start_weight="$2"
	local sum=0
	local i
	local value
	local remainder

	for ((i = 0; i < ${#digits}; i++)); do
		value="${digits:i:1}"
		sum=$((sum + value * (start_weight - i)))
	done

	remainder=$((sum % 11))
	if ((remainder < 2)); then
		echo 0
	else
		echo $((11 - remainder))
	fi
}

new_cpf() {
	local offset="$1"
	local base
	local digit_one
	local digit_two

	base="$(printf '%09d' $(((RUN_SEED + offset * 6151) % 1000000000)))"
	digit_one="$(cpf_digit "$base" 10)"
	digit_two="$(cpf_digit "${base}${digit_one}" 11)"
	echo "${base}${digit_one}${digit_two}"
}

api_post() {
	local path="$1"
	local body="$2"

	curl --fail-with-body --silent --show-error \
		-X POST \
		-H 'Content-Type: application/json' \
		"${AUTH_HEADER[@]}" \
		--data "$body" \
		"${BASE_URL}${path}"
}

api_get() {
	local path="$1"

	curl --fail-with-body --silent --show-error \
		-H 'Accept: application/json' \
		"${AUTH_HEADER[@]}" \
		"${BASE_URL}${path}"
}

create_party() {
	local name="$1"
	local document="$2"
	local role="$3"
	local payload

	payload="$(jq -nc \
		--arg name "$name" \
		--arg document "$document" \
		--arg role "$role" \
		'{name: $name, document: $document, roles: [$role]}')"

	api_post /parties "$payload" | jq -er '.id'
}

create_item() {
	local name="$1"
	local kind="$2"
	local price="$3"
	local payload

	payload="$(jq -nc \
		--arg name "$name" \
		--arg kind "$kind" \
		--argjson price "$price" \
		'{name: $name, kind: $kind, defaultPriceCents: $price, currency: "BRL"}')"

	api_post /items "$payload" | jq -er '.id'
}

create_transaction() {
	local type="$1"
	local party_id="$2"
	local description="$3"
	local lines="$4"
	local initial_payment="$5"
	local payload
	local response

	payload="$(jq -nc \
		--arg type "$type" \
		--arg party_id "$party_id" \
		--arg description "$description" \
		--arg due_date 2030-10-10T00:00:00Z \
		--argjson lines "$lines" \
		--argjson initial_payment "$initial_payment" \
		'{type: $type, description: $description, currency: "BRL", dueDate: $due_date, lines: $lines}
		| if $party_id == "" then . else . + {partyId: $party_id} end
		| if $initial_payment == null then . else . + {initialPayment: $initial_payment} end')"

	response="$(api_post /transactions "$payload")"
	jq -cer '{transactionId: .transaction.id, obligationId: .obligation.id, amountCents: .obligation.amountCents, status: .obligation.status}' <<<"$response"
}

create_payment() {
	local obligation_id="$1"
	local amount="$2"
	local method="$3"
	local payload

	payload="$(jq -nc \
		--argjson amount "$amount" \
		--arg method "$method" \
		'{amountCents: $amount, method: $method}')"

	api_post "/obligations/${obligation_id}/payments" "$payload" \
		| jq -cer '{obligationId: .obligation.id, paidAmountCents: .obligation.paidAmountCents, status: .obligation.status, payments: (.payments | length)}'
}

echo "Creating Relay sample dataset at ${BASE_URL}"
echo "Run ID: ${RUN_ID}"

organization_payload="$(jq -nc \
	--arg name "Relay Demo ${RUN_ID}" \
	--arg email "org-${RUN_ID}@relay.local" \
	--arg tax_id "$(new_cnpj 1)" \
	--arg owner_email "$OWNER_EMAIL" \
	--arg owner_document "$(new_cpf 2)" \
	'{
		organization: {
			name: $name,
			email: $email,
			taxId: $tax_id,
			type: "SERVICES",
			currency: "BRL",
			description: "Local integration-test organization"
		},
		user: {
			name: "Relay Demo Owner",
			email: $owner_email,
			password: "RelayDemo123!",
			document: $owner_document,
			phone: "5511999990000"
		}
	}')"

AUTH_HEADER=()
api_post /organizations "$organization_payload" >/dev/null
echo "✓ organization and owner created (${OWNER_EMAIL})"

login_payload="$(jq -nc \
	--arg email "$OWNER_EMAIL" \
	'{email: $email, password: "RelayDemo123!"}')"
TOKEN="$(api_post /login "$login_payload" | jq -er '.accessToken')"
AUTH_HEADER=(-H "Authorization: Bearer ${TOKEN}")
echo "✓ authenticated"

for ((i = 1; i <= CUSTOMER_COUNT; i++)); do
	CUSTOMER_IDS+=("$(create_party "Demo Customer ${i} ${RUN_ID}" "$(new_cpf $((100 + i)))" CUSTOMER)")
done
echo "✓ created ${#CUSTOMER_IDS[@]} customers"

for ((i = 1; i <= SUPPLIER_COUNT; i++)); do
	SUPPLIER_IDS+=("$(create_party "Demo Supplier ${i} ${RUN_ID}" "$(new_cnpj $((200 + i)))" SUPPLIER)")
done
echo "✓ created ${#SUPPLIER_IDS[@]} suppliers"

for ((i = 1; i <= SERVICE_COUNT; i++)); do
	SERVICE_IDS+=("$(create_item "Demo Service ${i} ${RUN_ID}" SERVICE $((150000 + i * 10000)))")
done

for ((i = 1; i <= PRODUCT_COUNT; i++)); do
	PRODUCT_IDS+=("$(create_item "Demo Product ${i} ${RUN_ID}" PRODUCT $((75000 + i * 5000)))")
done
echo "✓ created ${#SERVICE_IDS[@]} services and ${#PRODUCT_IDS[@]} products"

echo "Creating sales and exercising partial/full receivable payments"
for ((i = 0; i < SALE_COUNT; i++)); do
	service_id="${SERVICE_IDS[$((i % SERVICE_COUNT))]}"
	product_id="${PRODUCT_IDS[$((i % PRODUCT_COUNT))]}"
	customer_id="${CUSTOMER_IDS[$((i % CUSTOMER_COUNT))]}"
	service_price=$((450000 + i * 12500))
	product_price=$((25000 + i * 1500))
	lines="$(jq -nc \
		--arg service_id "$service_id" \
		--arg product_id "$product_id" \
		--argjson service_price "$service_price" \
		--argjson product_price "$product_price" \
		'[
			{itemId: $service_id, quantity: 1, unitPriceCents: $service_price},
			{itemId: $product_id, quantity: 2, unitPriceCents: $product_price}
		]')"
	initial_payment='null'
	if ((i % 2 == 0)); then
		initial_payment="$(jq -nc '{amountCents: 100000, method: "PIX"}')"
	fi
	transaction="$(create_transaction SALE "$customer_id" "Demo sale $((i + 1)) ${RUN_ID}" "$lines" "$initial_payment")"
	obligation_id="$(jq -er '.obligationId' <<<"$transaction")"
	amount="$(jq -er '.amountCents' <<<"$transaction")"
	paid_before=$((i % 2 == 0 ? 100000 : 0))
	partial=$((150000 + i * 5000))
	create_payment "$obligation_id" "$partial" PIX >/dev/null
	remaining=$((amount - paid_before - partial))
	create_payment "$obligation_id" "$remaining" BANK_TRANSFER >/dev/null
	echo "  sale $((i + 1)): obligation ${obligation_id} settled in multiple payments"
done

echo "Creating purchases and payables"
for ((i = 0; i < PURCHASE_COUNT; i++)); do
	product_id="${PRODUCT_IDS[$((i % PRODUCT_COUNT))]}"
	supplier_id="${SUPPLIER_IDS[$((i % SUPPLIER_COUNT))]}"
	price=$((180000 + i * 17500))
	lines="$(jq -nc \
		--arg product_id "$product_id" \
		--argjson price "$price" \
		'[{itemId: $product_id, quantity: 1, unitPriceCents: $price}]')"
	transaction="$(create_transaction PURCHASE "$supplier_id" "Demo purchase $((i + 1)) ${RUN_ID}" "$lines" null)"
	obligation_id="$(jq -er '.obligationId' <<<"$transaction")"
	amount="$(jq -er '.amountCents' <<<"$transaction")"
	create_payment "$obligation_id" "$amount" BANK_TRANSFER >/dev/null
	echo "  purchase $((i + 1)): obligation ${obligation_id} paid"
done

echo "Creating expenses and payables"
for ((i = 0; i < EXPENSE_COUNT; i++)); do
	price=$((35000 + i * 7500))
	lines="$(jq -nc \
		--arg description "Demo expense ${i} ${RUN_ID}" \
		--argjson price "$price" \
		'[{description: $description, quantity: 1, unitPriceCents: $price}]')"
	transaction="$(create_transaction EXPENSE "" "Demo expense $((i + 1)) ${RUN_ID}" "$lines" null)"
	obligation_id="$(jq -er '.obligationId' <<<"$transaction")"
	amount="$(jq -er '.amountCents' <<<"$transaction")"
	partial=$((amount / 2))
	create_payment "$obligation_id" "$partial" CASH >/dev/null
	create_payment "$obligation_id" "$((amount - partial))" CASH >/dev/null
	echo "  expense $((i + 1)): obligation ${obligation_id} settled in two payments"
done

party_total="$(api_get /parties | jq -er 'length')"
item_total="$(api_get /items | jq -er 'length')"
transaction_total="$(api_get /transactions | jq -er 'length')"
obligation_total="$(api_get /obligations | jq -er 'length')"
receivable_total="$(api_get '/obligations?direction=RECEIVABLE' | jq -er 'length')"
payable_total="$(api_get '/obligations?direction=PAYABLE' | jq -er 'length')"

cat <<SUMMARY

Sample dataset created successfully.
  Organization owner: ${OWNER_EMAIL}
  Parties:            ${party_total}
  Items:              ${item_total}
  Transactions:      ${transaction_total}
  Obligations:       ${obligation_total}
    Receivables:     ${receivable_total}
    Payables:        ${payable_total}

The JWT is intentionally not printed. To inspect the data, log in again with:
  email:    ${OWNER_EMAIL}
  password: RelayDemo123!
SUMMARY
