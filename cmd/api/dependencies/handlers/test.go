package handlers

import (
	"github.com/spector-asael/banking-api/cmd/api/dependencies/helpers"
	"net/http"
)

func (a *HandlerDependencies) testHandler(w http.ResponseWriter, r *http.Request) {
	helper := &helpers.HelperDependencies{
		Logger: a.Logger,
	}

	err := helper.WriteJSON(nil, 200, helpers.Envelope{"message": "Test successful"}, nil)
	if err != nil {
		a.Logger.Error("error writing JSON response", "error", err)
	}
}