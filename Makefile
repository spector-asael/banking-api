# CMPS3162 Advanced Databases - Banking System
# Asael Tobar
# March 19th 2026

# Pass in the .envrc file, which exports BANK_DB_DSN
include .envrc

## run: run the cmd/api application
run: 
	@echo 'Running application...'
	@go run ./cmd/api -db-dsn="${BANK_DB_DSN}" -env=development -limiter-rps=2 -limiter-burst=5

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