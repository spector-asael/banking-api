// Filename: cmd/api/users.go
package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/spector-asael/banking-api/cmd/api/dependencies/helpers"
	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/validator"
)

// GET /users
func (a *HandlerDependencies) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Define input struct for filtering/pagination
	var input struct {
		Username string
		Email    string
		Filters  data.Filters
	}

	// 2. Initialize validator
	v := validator.New()

	// 3. Read query parameters
	qs := r.URL.Query()

	input.Username = a.Helper.ReadString(qs, "name", "")
	input.Email = a.Helper.ReadString(qs, "email", "")

	input.Filters.Page = a.Helper.ReadInt(qs, "page", 1)
	input.Filters.PageSize = a.Helper.ReadInt(qs, "page_size", 10)
	input.Filters.Sort = a.Helper.ReadString(qs, "sort", "id")

	// 4. Define Safelist for SQL injection protection
	input.Filters.SortSafelist = []string{
		"id", "-id",
		"name", "-name",
		"email", "-email",
		"activated", "-activated",
	}

	// 5. Execute validation
	data.ValidateFilters(v, input.Filters)
	if !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// 6. Call the Model (Pass filters and search criteria)
	users, metadata, err := a.Models.Users.GetAll(input.Username, input.Email, input.Filters)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	// 7. Send Envelope Response
	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{
		"users":     users,
		"@metadata": metadata,
	}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

// GET /users/:id
func (a *HandlerDependencies) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		a.Helper.NotFoundResponse(w, r)
		return
	}

	user, err := a.Models.Users.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.Helper.NotFoundResponse(w, r)
		default:
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"user": user}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

// PATCH /users/:id
func (a *HandlerDependencies) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		a.Helper.NotFoundResponse(w, r)
		return
	}

	// Retrieve existing record
	user, err := a.Models.Users.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.Helper.NotFoundResponse(w, r)
		default:
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}

	// Read the update JSON into an anonymous struct (Partial Updates)
	var input struct {
		Username    *string `json:"name"`
		Email       *string `json:"email"`
		Activated   *bool   `json:"activated"`
		AccountType *int    `json:"account_type"`
	}

	err = a.Helper.ReadJSON(w, r, &input)
	if err != nil {
		a.Helper.BadRequestResponse(w, r, err)
		return
	}

	// Update only the fields that were provided
	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Activated != nil {
		user.Activated = *input.Activated
	}
	if input.AccountType != nil {
		user.Account_type = *input.AccountType
	}

	// Validate the updated user record
	v := validator.New()
	data.ValidateUser(v, user) // Assuming you have this validation method
	if !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// Save the changes
	err = a.Models.Users.Update(user)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"user": user}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}
