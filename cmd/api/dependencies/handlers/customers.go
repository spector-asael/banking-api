package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"context"

	"github.com/julienschmidt/httprouter"
	"github.com/spector-asael/banking-api/cmd/api/dependencies/helpers"
	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/validator"
)

// GET /customers
func (a *HandlerDependencies) getAllCustomersHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		filters data.Filters
	}
	qs := r.URL.Query()
	input.filters.Page = a.Helper.ReadInt(qs, "page", 1)
	input.filters.PageSize = a.Helper.ReadInt(qs, "page_size", 5)
	input.filters.Sort = a.Helper.ReadString(qs, "sort", "id")
	input.filters.SortSafelist = []string{"id", "-id", "person_id", "-person_id", "kyc_status_id", "-kyc_status_id", "created_at", "-created_at"}

	v := validator.New()
	data.ValidateFilters(v, input.filters)
	if !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}

	customers, metadata, err := a.Models.Customers.GetAll(input.filters)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}
	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"customers": customers, "@metadata": metadata}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

// GET /customers/:id
func (a *HandlerDependencies) getCustomerByIDHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	idStr := params.ByName("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"id": "must be a valid integer"})
		return
	}
	customer, err := a.Models.Customers.GetByID(id)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			a.Helper.WriteJSON(w, http.StatusNotFound, helpers.Envelope{"error": "customer not found"}, nil)
		default:
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}
	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"customer": customer}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

// POST /customers
func (a *HandlerDependencies) createCustomerHandler(w http.ResponseWriter, r *http.Request) {
	// DEBUG: Log the incoming request body
	defer r.Body.Close()
	bodyBytes, _ := io.ReadAll(r.Body)
	fmt.Printf("[DEBUG] Raw request body: %s\n", string(bodyBytes))
	// Re-create the body for JSON decoding
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Input from JSON
	var input struct {
		SSID     string `json:"ssid"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := a.Helper.ReadJSON(w, r, &input)
	if err != nil {
		a.Helper.BadRequestResponse(w, r, err)
		return
	}
	fmt.Printf("[DEBUG] Parsed input: %+v\n", input)

	// Validate input
	v := validator.New()
	v.Check(input.SSID != "", "ssid", "must be provided")
	v.Check(input.Username != "", "username", "must be provided")
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.IsEmpty() {
		fmt.Printf("[DEBUG] Validation errors: %+v\n", v.Errors)
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// Lookup person by SSID
	person, err := a.Models.Persons.GetBySSID(input.SSID)
	if err != nil {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"ssid": "person not found"})
		return
	}

	// Build customer
	customer := data.Customer{
		PersonID:    person.ID,
		KYCStatusID: 1, // Pending
	}

	// Start transaction
	tx, err := a.Models.Customers.DB.Begin()
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}
	defer tx.Rollback()

	// Insert customer
	err = a.Models.Customers.InsertTx(tx, &customer)
	if err != nil {
		if err == data.ErrDuplicateCustomer {
			a.Helper.FailedValidationResponse(w, r, map[string]string{"ssid": "customer already exists for this person"})
			return
		}
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	// Create user
	customerID := int(customer.ID)
	user := data.User{
		Username:     input.Username,
		Email:        input.Email,
		Activated:    true,
		Customer_id:  &customerID,
		Employee_id:  nil, // ensure NULL in DB for customer
		Account_type: 0,   // CUSTOMER
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	err = a.Models.Users.InsertTx(tx, &user)
	if err != nil {
		if err == data.ErrDuplicateEmail {
			a.Helper.FailedValidationResponse(w, r, map[string]string{"email": "email already in use"})
			return
		}
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	// Commit
	err = tx.Commit()
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}
	// Send the email as a Goroutine. We do this because it might take a long time
	// and we don't want our handler to wait for that to finish. We will implement
	// the background() function later
	a.Helper.Background(func() {
		err = a.Mailer.Send(user.Email, "user_welcome.tmpl", user)
		if err != nil {
			a.Logger.Error(err.Error())
		}
	})

	// Response
	err = a.Helper.WriteJSON(w, http.StatusCreated, helpers.Envelope{
		"customer": customer,
		"user":     user,
	}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

// PATCH /customers/:id/kyc-status
func (a *HandlerDependencies) updateCustomerKYCStatusHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	idStr := params.ByName("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"id": "must be a valid integer"})
		return
	}
	customer, err := a.Models.Customers.GetByID(id)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			a.Helper.WriteJSON(w, http.StatusNotFound, helpers.Envelope{"error": "customer not found"}, nil)
		default:
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}
	var input data.UpdateKYCStatusInput
	err = a.Helper.ReadJSON(w, r, &input)
	if err != nil {
		a.Helper.BadRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	data.ValidateUpdateKYCStatus(v, &input)
	if !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}
	customer.KYCStatusID = input.KYCStatusID
	err = a.Models.Customers.Update(customer)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}
	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"customer": customer}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

// DELETE /customers/:id
func (a *HandlerDependencies) deleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	idStr := params.ByName("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"id": "must be a valid integer"})
		return
	}

	// Create a context with a 3-second timeout
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// 1. Start a transaction
	tx, err := a.Models.Customers.DB.BeginTx(ctx, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}
	defer tx.Rollback()

	// 2. Delete the associated User account first
	_, err = tx.ExecContext(ctx, "DELETE FROM users WHERE customer_id = $1", id)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	// 3. Delete the Customer
	result, err := tx.ExecContext(ctx, "DELETE FROM customers WHERE id = $1", id)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	if rowsAffected == 0 {
		a.Helper.WriteJSON(w, http.StatusNotFound, helpers.Envelope{"error": "customer not found"}, nil)
		return
	}

	// 4. Commit the transaction
	if err = tx.Commit(); err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "customer and associated user account deleted successfully"}, nil)
}
