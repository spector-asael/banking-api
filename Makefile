# CMPS3162 Advanced Databases - Banking System
# Asael Tobar
# March 19th 2026


# Import environment variables from .envrc
include .envrc

## db/psql: Connect to the banking database using psql
.PHONY: db
db:
	psql ${BANK_DB_DSN}

run: 
	@echo 'Running application...'
	@GODEBUG=netdns=cgo go run ./cmd/api \
	-db-dsn="${BANK_DB_DSN}" \
	-env=development \
	-limiter-rps=${RPS} \
	-limiter-burst=${BURST} \
	-limiter-enabled=true \
	-cors-trusted-origins=${CORS_TRUSTED_ORIGINS} \
	-smtp-host=${SMTP_HOST} \
	-smtp-port=${SMTP_PORT} \
	-smtp-username=${SMTP_USERNAME} \
	-smtp-password=${SMTP_PASSWORD} \
.PHONY: migrations/new

run/frontend:
	@echo 'Running frontend...'
	cd frontend && go run .
migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## migrations/up: Apply all up database migrations
.PHONY: migrations/up
migrations/up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${BANK_DB_DSN} up

## db/migrations/down: Revert all migrations
.PHONY: migrations/down
migrations/down:
	@echo 'Reverting all migrations...'
	migrate -path ./migrations -database ${BANK_DB_DSN} down

## db/migrations/fix version=$1: Force schema_migrations version
.PHONY: migrations/fix
migrations/fix:
	@echo 'Forcing schema migrations version to ${version}...'
	migrate -path ./migrations -database ${BANK_DB_DSN} force ${version}

.PHONY: getpersons
# Get all persons (default page=1, page_size=5, sort=id)
getpersons:
	@echo 'Getting all persons...'
	curl -i http://localhost:4000/api/persons

.PHONY: getpersons-custom
# Get first page, 5 per page, sorted by last_name descending
getpersons-custom:
	@echo 'Getting first page of persons, 5 per page, sorted by last_name'
	curl -i "http://localhost:4000/api/persons?page=1&page_size=5&sort=-last_name"


.PHONY: getpersons-filter
# Filter by name containing "John"
getpersons-filter:
	@echo 'Getting persons with name containing "John"...'
	curl -i "http://localhost:4000/api/persons?name=John"

.PHONY: getperson-ssid
# Get person by social security number
getperson-ssid:
	@echo 'Getting person with social security number...'
	curl -i http://localhost:4000/api/persons/987-65-4321
# createperson:
# 	@echo 'Creating a new person...'
# 	curl -i -X POST http://localhost:4000/api/persons \
# 	-H "Content-Type: application/json" \
# 	-d '{"first_name":"Alice","last_name":"Smith","social_security_number":"123456789","email":"alice.smith@example.com","date_of_birth":"1990-05-15T00:00:00Z","phone_number":"1234567890","living_address":"123 Main St, Cityville"}'

# CUSTOMER ROUTE TESTS
.PHONY: getcustomers
getcustomers:
	@echo 'Getting all customers...'
	curl -i http://localhost:4000/api/customers

.PHONY: getcustomers-gzip
getcustomers-uncompressed:
	@echo 'Getting all customers with gzip encoding...'
	curl -i -H "Accept-Encoding: gzip" --compressed http://localhost:4000/api/customers

getcustomers-gzip:
	@echo 'Getting all customers with gzip encoding...'
	curl -i -H "Accept-Encoding: gzip" http://localhost:4000/api/customers --output response.gz

getcustomers-nogzip:
	@echo 'Getting all customers with gzip encoding...'
	curl -i http://localhost:4000/api/customers --output response.gz

loop-create:
	@echo "Creating multiple persons..."
	@for i in $$(seq 1 20); do \
		curl -s -i GET http://localhost:4000/api/persons \
		-H "Content-Type: application/json"; \
	done

getcustomers-compressed:
	@echo 'Getting all customers with gzip encoding...'
	curl -i -H "Accept-Encoding: gzip" http://localhost:4000/api/customers

.PHONY: getcustomers-custom
getcustomers-custom:
	@echo 'Getting first page of customers, 10 per page, sorted by created_at desc'
	curl -i "http://localhost:4000/api/customers?page=1&page_size=10&sort=-created_at"

.PHONY: createcustomer
createcustomer:
	@echo 'Creating a new customer for person with SSN 987-65-4321 (assumes person exists and has id=1, adjust if needed)...'
	curl -i -X POST http://localhost:4000/api/customers \
	-H "Content-Type: application/json" \
	-d '{"person_id":1,"kyc_status_id":1}'

.PHONY: getcustomer-id
getcustomer-id:
	@echo 'Getting customer with id=1...'
	curl -i http://localhost:4000/api/customers/1

.PHONY: updatecustomer-kyc
updatecustomer-kyc:
	@echo 'Updating KYC status for customer id=1 to Approved (kyc_status_id=2)...'
	curl -i -X PATCH http://localhost:4000/api/customers/1/kyc-status \
	-H "Content-Type: application/json" \
	-d '{"kyc_status_id":2}'

.PHONY: deletecustomer
deletecustomer:
	@echo 'Deleting customer with id=1...'
	curl -i -X DELETE http://localhost:4000/api/customers/1

# ACCOUNT ROUTE TESTS
.PHONY: getaccounts
getaccounts:
	@echo 'Getting all accounts...'
	curl -i http://localhost:4000/api/accounts

.PHONY: getaccounts-custom
getaccounts-custom:
	@echo 'Getting first page of accounts, 10 per page, sorted by created_at desc'
	curl -i "http://localhost:4000/api/accounts?page=1&page_size=10&sort=-created_at"

.PHONY: createaccount
createaccount:
	@echo 'Creating a new account (adjust fields as needed)...'
	curl -i -X POST http://localhost:4000/api/accounts \
	-H "Content-Type: application/json" \
	-d '{"account_number":"10000001","branch_id_opened_at":1,"account_type_id":1,"gl_account_id":1,"status":"active","opened_at":"2026-03-23T00:00:00Z"}'

.PHONY: getaccount-id
getaccount-id:
	@echo 'Getting account with id=1...'
	curl -i http://localhost:4000/api/accounts/1


# DEPOSIT ROUTE TESTS
.PHONY: createdeposit
createdeposit:
	@echo 'Creating a deposit for account_id=1 (adjust fields as needed)...'
	curl -i -X POST http://localhost:4000/api/deposits \
	-H "Content-Type: application/json" \
	-d '{"account_id":1,"amount":100.00,"description":"Test deposit"}'