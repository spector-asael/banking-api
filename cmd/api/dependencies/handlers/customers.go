package handlers

import (
	"net/http"
	"strconv"
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
	var input data.Customer
	err := a.Helper.ReadJSON(w, r, &input)
	if err != nil {
		a.Helper.BadRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	data.ValidateCustomer(v, &input)
	if !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}
	// Check if person exists
	_, err = a.Models.Persons.GetByID(input.PersonID)
	if err != nil {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"person_id": "referenced person does not exist"})
		return
	}
	// Optionally: check if KYC status exists (not strictly required if FK enforced)
	err = a.Models.Customers.Insert(&input)
	if err != nil {
		if err == data.ErrDuplicateCustomer {
			a.Helper.FailedValidationResponse(w, r, map[string]string{"person_id": "customer for this person already exists"})
			return
		}
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}
	err = a.Helper.WriteJSON(w, http.StatusCreated, helpers.Envelope{"customer": input}, nil)
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
	err = a.Models.Customers.Delete(id)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			a.Helper.WriteJSON(w, http.StatusNotFound, helpers.Envelope{"error": "customer not found"}, nil)
		default:
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}
	a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "customer deleted successfully"}, nil)
}