package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/validator"
)

// 1. Updated to use AccountNumber (string)
type withdrawalRequest struct {
	AccountNumber string  `json:"account_number"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
}

func (h *HandlerDependencies) HandleWithdrawal(w http.ResponseWriter, r *http.Request) {
	var req withdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// 2. Updated validation
	v := validator.New()
	v.Check(req.AccountNumber != "", "account_number", "must be provided")
	v.Check(req.Amount > 0, "amount", "must be greater than zero")
	if !v.IsEmpty() {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(v.Errors)
		return
	}

	// 3. Fetch by Account Number, keeping our real error handling
	account, err := h.Models.Accounts.GetByAccountNumber(req.AccountNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Create Journal Entry
	je := &data.JournalEntry{
		ReferenceTypeID: 3, // 3 = withdrawal
		ReferenceID:     account.ID,
		Description:     req.Description,
		CreatedAt:       time.Now(),
	}
	if err := h.Models.JournalEntries.Insert(je); err != nil {
		http.Error(w, "could not create journal entry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Ledger Entry 1: Debit the customer's GL account (Liability decreases)
	customerDebit := &data.LedgerEntry{
		GLAccountID:    account.GLAccountID,
		JournalEntryID: je.ID,
		Debit:          req.Amount, // <-- Debit here!
		Credit:         0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := h.Models.LedgerEntries.Insert(customerDebit); err != nil {
		http.Error(w, "could not create customer ledger entry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Ledger Entry 2: Credit the Bank's Vault Cash GL account (Asset decreases)
	bankVaultGLAccountID := int64(1) // Using the exact same GL ID we used for deposits!

	bankCashCredit := &data.LedgerEntry{
		GLAccountID:    bankVaultGLAccountID,
		JournalEntryID: je.ID,
		Debit:          0,
		Credit:         req.Amount, // <-- Credit here!
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := h.Models.LedgerEntries.Insert(bankCashCredit); err != nil {
		http.Error(w, "could not create bank ledger entry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// -----------------------------------------------------------------------------------

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Withdrawal successful",
		"journal_entry": je,
		"ledger_entries": []interface{}{
			customerDebit,
			bankCashCredit,
		},
	})
}