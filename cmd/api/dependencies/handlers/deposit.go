
package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/validator"
)

type depositRequest struct {
	AccountID   int64   `json:"account_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

func (h *HandlerDependencies) HandleDeposit(w http.ResponseWriter, r *http.Request) {
	var req depositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	v := validator.New()
	v.Check(req.AccountID > 0, "account_id", "must be provided and valid")
	v.Check(req.Amount > 0, "amount", "must be greater than zero")
	   if !v.IsEmpty() {
		   w.WriteHeader(http.StatusUnprocessableEntity)
		   json.NewEncoder(w).Encode(v.Errors)
		   return
	   }

	account, err := h.Models.Accounts.GetByID(req.AccountID)
	if err != nil {
		http.Error(w, "account not found", http.StatusNotFound)
		return
	}

	// Create JournalEntry
	je := &data.JournalEntry{
		ReferenceTypeID: 1, // 1 = deposit (adjust as needed)
		ReferenceID:     account.ID,
		Description:     req.Description,
		CreatedAt:       time.Now(),
	}
	if err := h.Models.JournalEntries.Insert(je); err != nil {
		http.Error(w, "could not create journal entry", http.StatusInternalServerError)
		return
	}

	// Create LedgerEntry (credit customer GL account)
	le := &data.LedgerEntry{
		GLAccountID:    account.GLAccountID,
		JournalEntryID: je.ID,
		Debit:          0,
		Credit:         req.Amount,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := h.Models.LedgerEntries.Insert(le); err != nil {
		http.Error(w, "could not create ledger entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"journal_entry": je,
		"ledger_entry":  le,
	})
}