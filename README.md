# Banking System API

This project is a backend API for a banking system, developed as part of the Advanced Databases course (CMPS3162) at the university. It is designed to provide a robust, modular, and extensible platform for managing core banking operations and data.

###  Powerpoint slides: https://docs.google.com/presentation/d/1Sm1xGnqAPoaC1KrdQNbNRrLOmmB-NfmXFl6BphwJ1QQ/edit?slide=id.g7f9262ee2f_0_26276#slide=id.g7f9262ee2f_0_26276 

### ERD Diagram: https://docs.google.com/document/d/1X0Ej6I9bURw0uv_AQU7SAFZcl7AGkhmJKYpjNLXJz2I/edit?tab=t.0

## Features
- **Person Management:** Create, update, retrieve, and delete person records.
- **Customer Management:** Manage customers, including KYC status and customer data.
- **Account Management:** Create and manage bank accounts, including ownership and account types.
- **Transactions:** Support for deposits, withdrawals, transfers, and loan payments.
- **Loans:** Manage loan creation and payments.
- **Observability:** Metrics endpoint for monitoring and observability.
- **Middleware:** Includes logging, metrics, rate limiting, CORS, gzip compression, and panic recovery.

## Technologies Used
- **Go (Golang):** Main programming language.
- **PostgreSQL:** Database backend (configured via environment variables).
- **httprouter:** High-performance HTTP request router.
- **expvar:** For exposing metrics.

## Project Structure
- `cmd/api/` - Main API server and dependencies.
- `frontend/` - CLI or web frontend for interacting with the API.
- `internal/` - Internal packages for data models, validation, and business logic.
- `migrations/` - SQL migration scripts for database schema management.

## Usage
- Run the API server: `make run`
- Test endpoints using provided Makefile targets (e.g., `make getpersons`, `make getcustomers`)
- Database migrations: `make migrations/up` and `make migrations/down`

## Purpose
This project is intended for educational purposes, demonstrating advanced database concepts, API design, and backend engineering best practices in a banking context. It can be extended for further research or as a foundation for more complex financial applications.

---

**Author:** Asael Tobar
**Course:** Advanced Databases (CMPS3162)
**Date:** March 2026
