package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/validator"
)

type depositRequest struct {
	AccountNumber string  `json:"account_number"` // Changed from AccountID int64
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
}

func (h *HandlerDependencies) HandleDeposit(w http.ResponseWriter, r *http.Request) {
    var req depositRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    // Validate against the string AccountNumber
    v := validator.New()
    v.Check(req.AccountNumber != "", "account_number", "must be provided")
    v.Check(req.Amount > 0, "amount", "must be greater than zero")
    if !v.IsEmpty() {
        w.WriteHeader(http.StatusUnprocessableEntity)
        json.NewEncoder(w).Encode(v.Errors)
        return
    }

    // 1. Get the customer's account by Account Number instead of ID
    account, err := h.Models.Accounts.GetByAccountNumber(req.AccountNumber)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 2. Create JournalEntry
    je := &data.JournalEntry{
        ReferenceTypeID: 1, // 1 = deposit
        ReferenceID:     account.ID, 
        Description:     req.Description,
        CreatedAt:       time.Now(),
    }
    // FIX: This must save the JournalEntry 'je', not 'bankCashDebit'
    if err := h.Models.JournalEntries.Insert(je); err != nil {
        http.Error(w, "could not create journal entry: " + err.Error(), http.StatusInternalServerError)
        return
    }

    // 3. Ledger Entry 1: Credit the customer's GL account (Liability increases)
    customerCredit := &data.LedgerEntry{
        GLAccountID:    account.GLAccountID,
        JournalEntryID: je.ID,
        Debit:          0,
        Credit:         req.Amount,
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }
    if err := h.Models.LedgerEntries.Insert(customerCredit); err != nil {
        http.Error(w, "could not create customer ledger entry: " + err.Error(), http.StatusInternalServerError)
        return
    }

    // 4. Ledger Entry 2: Debit the Bank's Vault Cash GL account (Asset increases)
    bankVaultGLAccountID := int64(1) // <-- FIXED: Changed from 999 to 1

    bankCashDebit := &data.LedgerEntry{
        GLAccountID:    bankVaultGLAccountID,
        JournalEntryID: je.ID,
        Debit:          req.Amount,
        Credit:         0,
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }
    // FIX: Here is the correct place to check for the bankCashDebit error
    if err := h.Models.LedgerEntries.Insert(bankCashDebit); err != nil {
        http.Error(w, "could not create bank ledger entry: " + err.Error(), http.StatusInternalServerError)
        return
    }

    // -----------------------------------------------------------------------------------

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message":       "Deposit successful",
        "journal_entry": je,
        "ledger_entries": []interface{}{
            customerCredit,
            bankCashDebit,
        },
    })
}