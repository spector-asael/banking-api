// Filename: cmd/api/tokens.go
package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/spector-asael/banking-api/cmd/api/dependencies/helpers"
	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/validator"
)

// GET /accounts/:accountNumber/history
func (a *HandlerDependencies) getAccountHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extract the account number from the URL parameters
	params := httprouter.ParamsFromContext(r.Context())
	accountNumber := params.ByName("accountNumber")

	if accountNumber == "" {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"account_number": "must be provided"})
		return
	}

	// 2. Setup Input struct for filters (Pagination)
	var input struct {
		filters data.Filters
	}

	qs := r.URL.Query()

	// Using your Helper methods to read query parameters
	input.filters.Page = a.Helper.ReadInt(qs, "page", 1)
	input.filters.PageSize = a.Helper.ReadInt(qs, "page_size", 10) // Standard banking default
	input.filters.Sort = a.Helper.ReadString(qs, "sort", "-date")  // Default to newest first
	input.filters.SortSafelist = []string{"date", "-date", "amount", "-amount"}

	// 3. Validate Filters
	v := validator.New()
	data.ValidateFilters(v, input.filters)
	if !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// 4. Fetch the data from the Model
	// Assuming you added Transactions to your Models struct: a.Models.Transactions
	history, metadata, err := a.Models.Transactions.GetHistory(accountNumber, input.filters)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	// 5. Send the JSON response with the "@metadata" envelope
	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{
		"history":   history,
		"@metadata": metadata,
	}, nil)

	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}
