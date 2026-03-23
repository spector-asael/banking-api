# CMPS3162 Advanced Databases - Banking System
# Asael Tobar
# March 19th 2026

# Pass in the .envrc file, which exports BANK_DB_DSN
include .envrc

## run: run the cmd/api application
run: 
	@echo 'Running application...'
	@go run ./cmd/api \
	-db-dsn="${BANK_DB_DSN}" \
	-env=development -limiter-rps=2 \
	-limiter-burst=5 \
	-limiter-enabled=false \
	-cors-trusted-origins="http://localhost:9000 http://localhost:9000"

## db/psql: Connect to the banking database using psql
.PHONY: db
db:
	psql ${BANK_DB_DSN}

## db/migrations/new name=$1: Create a new database migration
.PHONY: migrations/new
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


.PHONY: testroute 
testroute:
	@echo 'Testing test route...'
	curl -i http://localhost:4000/test

.PHONY: getpersons
# Get all persons (default page=1, page_size=5, sort=id)
getpersons:
	@echo 'Getting all persons...'
	curl -i http://localhost:4000/api/persons

.PHONY: getpersons-custom
# Get first page, 10 per page, sorted by last_name descending
getpersons-custom:
	@echo 'Getting first page of persons, 10 per page, sorted by last_name'
	curl -i "http://localhost:4000/api/persons?page=1&page_size=10&sort=-last_name"


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