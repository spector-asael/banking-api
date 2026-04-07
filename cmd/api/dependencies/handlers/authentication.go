// Filename: cmd/api/tokens.go
package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/spector-asael/banking-api/cmd/api/dependencies/helpers"
	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/validator"
)

func (a *HandlerDependencies) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Get the body from the request and store in a temporary struct
	// The client will give us their email and password. We will will give them
	// a Bearer token
	var incomingData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := a.Helper.ReadJSON(w, r, &incomingData)
	if err != nil {
		a.Helper.BadRequestResponse(w, r, err)
		return
	}
	// Validate the email and password provided by the client.
	v := validator.New()

	data.ValidateEmail(v, incomingData.Email)
	data.ValidatePasswordPlaintext(v, incomingData.Password)

	if !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}
	// Is there an associated user for the provided email?
	user, err := a.Models.Users.GetByEmail(incomingData.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.Helper.InvalidCredentialsResponse(w, r)
		default:
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}
	// The user is found. Does their password match?
	match, err := user.Password.Matches(incomingData.Password)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}
	// Wrong password
	// We will define invalidCredentialsResponse() later
	if !match {
		a.Helper.InvalidCredentialsResponse(w, r)
		return
	}
	token, err := a.Models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	data := helpers.Envelope{
		"authentication_token": token,
	}

	// Return the bearer token
	err = a.Helper.WriteJSON(w, http.StatusCreated, data, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}
