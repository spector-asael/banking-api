package handlers

import (
	"github.com/spector-asael/banking-api/cmd/api/dependencies/helpers"
	"net/http"
)

func (a *HandlerDependencies) testHandler(w http.ResponseWriter, r *http.Request) {
	helper := &helpers.HelperDependencies{
		Logger: a.Logger,
	}

	err := helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "Test successful"}, nil)
	if err != nil {
		a.Logger.Error("error writing JSON response", "error", err)
	}
}

func (a *HandlerDependencies) postTestHandler(w http.ResponseWriter, r *http.Request) {
	helper := &helpers.HelperDependencies{
		Logger: a.Logger,
	}

	// Define a struct to decode incoming JSON
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode JSON request body
	err := helper.ReadJSON(w, r, &input)
	if err != nil {
		a.Logger.Error("error reading JSON", "error", err)
		helper.WriteJSON(w, http.StatusBadRequest, helpers.Envelope{
			"error": "invalid request body",
		}, nil)
		return
	}

	// Hardcoded credentials
	const correctEmail = "test@example.com"
	const correctPassword = "password123"

	// Check credentials
	if input.Email == correctEmail && input.Password == correctPassword {
		err = helper.WriteJSON(w, http.StatusOK, helpers.Envelope{
			"message": "Login successful",
		}, nil)
	} else {
		err = helper.WriteJSON(w, http.StatusUnauthorized, helpers.Envelope{
			"error": "invalid email or password",
		}, nil)
	}

	if err != nil {
		a.Logger.Error("error writing JSON response", "error", err)
	}
}