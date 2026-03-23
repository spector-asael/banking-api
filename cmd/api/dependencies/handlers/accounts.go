package handlers

import (
	"net/http"
	"strconv"
	"github.com/julienschmidt/httprouter"
	"github.com/spector-asael/banking-api/cmd/api/dependencies/helpers"
	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/validator"
	"time"
)

// PATCH /accounts/:id
func (a *HandlerDependencies) updateAccountHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	idStr := params.ByName("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"id": "must be a valid integer"})
		return
	}
	account, err := a.Models.Accounts.GetByID(id)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			a.Helper.WriteJSON(w, http.StatusNotFound, helpers.Envelope{"error": "account not found"}, nil)
		default:
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}
	var input struct {
		Status    *string    `json:"status"`
		ClosedAt  *string    `json:"closed_at"`
	}
	err = a.Helper.ReadJSON(w, r, &input)
	if err != nil {
		a.Helper.BadRequestResponse(w, r, err)
		return
	}
	// Only update provided fields
	if input.Status != nil {
		account.Status = *input.Status
	}
	if input.ClosedAt != nil {
		if *input.ClosedAt == "" {
			account.ClosedAt = nil
		} else {
			t, err := time.Parse(time.RFC3339, *input.ClosedAt)
			if err != nil {
				a.Helper.FailedValidationResponse(w, r, map[string]string{"closed_at": "must be a valid RFC3339 timestamp or empty string"})
				return
			}
			account.ClosedAt = &t
		}
	}
	// Validate updated account
	v := validator.New()
	data.ValidateAccount(v, account)
	if !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}
	// Save changes
	err = a.Models.Accounts.Update(account)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}
	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"account": account}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

// GET /accounts
func (a *HandlerDependencies) getAllAccountsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		filters data.Filters
	}
	qs := r.URL.Query()
	input.filters.Page = a.Helper.ReadInt(qs, "page", 1)
	input.filters.PageSize = a.Helper.ReadInt(qs, "page_size", 5)
	input.filters.Sort = a.Helper.ReadString(qs, "sort", "id")
	input.filters.SortSafelist = []string{"id", "-id", "account_number", "-account_number", "created_at", "-created_at"}

	v := validator.New()
	data.ValidateFilters(v, input.filters)
	if !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}

	accounts, metadata, err := a.Models.Accounts.GetAll(input.filters)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}
	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"accounts": accounts, "@metadata": metadata}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

// POST /accounts
// Input struct for account creation with SSID
type createAccountInput struct {
	AccountNumber   string    `json:"account_number"`
	BranchID        int64     `json:"branch_id_opened_at"`
	AccountTypeID   int64     `json:"account_type_id"`
	GLAccountID     int64     `json:"gl_account_id"`
	Status          string    `json:"status"`
	OpenedAt        string    `json:"opened_at"`
	ClosedAt        *string   `json:"closed_at,omitempty"`
	SSID            string    `json:"social_security_number"`
	IsJointAccount  bool      `json:"is_joint_account"`
}

func (a *HandlerDependencies) createAccountHandler(w http.ResponseWriter, r *http.Request) {
	var input createAccountInput
	err := a.Helper.ReadJSON(w, r, &input)
	if err != nil {
		a.Helper.BadRequestResponse(w, r, err)
		return
	}
	// Validate account fields
	v := validator.New()
	v.Check(input.AccountNumber != "", "account_number", "must be provided")
	v.Check(input.BranchID > 0, "branch_id_opened_at", "must be provided and valid")
	v.Check(input.AccountTypeID > 0, "account_type_id", "must be provided and valid")
	v.Check(input.GLAccountID > 0, "gl_account_id", "must be provided and valid")
	v.Check(input.Status != "", "status", "must be provided")
	v.Check(input.SSID != "", "social_security_number", "must be provided")
	if !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}
	// Parse opened_at and closed_at
	openedAt, err := time.Parse(time.RFC3339, input.OpenedAt)
	if err != nil {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"opened_at": "must be a valid RFC3339 timestamp"})
		return
	}
	var closedAtPtr *time.Time
	if input.ClosedAt != nil {
		t, err := time.Parse(time.RFC3339, *input.ClosedAt)
		if err != nil {
			a.Helper.FailedValidationResponse(w, r, map[string]string{"closed_at": "must be a valid RFC3339 timestamp"})
			return
		}
		closedAtPtr = &t
	}
	// Lookup person by SSID
	person, err := a.Models.Persons.GetBySSID(input.SSID)
	if err != nil {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"social_security_number": "person not found"})
		return
	}
	// Lookup customer by person ID
	customer, err := a.Models.Customers.GetByID(person.ID)
	if err != nil {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"customer": "customer for this person not found"})
		return
	}
	// Create account
	account := data.Account{
		AccountNumber: input.AccountNumber,
		BranchID:      input.BranchID,
		AccountTypeID: input.AccountTypeID,
		GLAccountID:   input.GLAccountID,
		Status:        input.Status,
		OpenedAt:      openedAt,
		ClosedAt:      closedAtPtr,
	}
	err = a.Models.Accounts.Insert(&account)
	if err != nil {
		if err == data.ErrDuplicateAccount {
			a.Helper.FailedValidationResponse(w, r, map[string]string{"account_number": "account with this account number already exists"})
			return
		}
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}
	// Assign ownership
	ownership := data.AccountOwnership{
		CustomerID:     customer.ID,
		AccountID:      account.ID,
		IsJointAccount: input.IsJointAccount,
	}
	err = a.Models.AccountOwnerships.Insert(&ownership)
	if err != nil {
		if err == data.ErrDuplicateAccountOwnership {
			a.Helper.FailedValidationResponse(w, r, map[string]string{"ownership": "ownership for this customer and account already exists"})
			return
		}
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}
	// Respond with account and ownership
	err = a.Helper.WriteJSON(w, http.StatusCreated, helpers.Envelope{"account": account, "ownership": ownership}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

// GET /accounts/:id
func (a *HandlerDependencies) getAccountByIDHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	idStr := params.ByName("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"id": "must be a valid integer"})
		return
	}
	account, err := a.Models.Accounts.GetByID(id)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			a.Helper.WriteJSON(w, http.StatusNotFound, helpers.Envelope{"error": "account not found"}, nil)
		default:
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}
	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"account": account}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

// DELETE /accounts/:id
func (a *HandlerDependencies) deleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	idStr := params.ByName("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"id": "must be a valid integer"})
		return
	}
	err = a.Models.Accounts.Delete(id)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			a.Helper.WriteJSON(w, http.StatusNotFound, helpers.Envelope{"error": "account not found"}, nil)
		default:
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}
	a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "account deleted successfully"}, nil)
}
