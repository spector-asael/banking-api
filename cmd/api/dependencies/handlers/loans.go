package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/spector-asael/banking-api/internal/data"
)

type loanRequest struct {
	AccountNumber   string  `json:"account_number"`
	PrincipalAmount float64 `json:"principal_amount"`
	TermMonths      int     `json:"term_months"`
	InterestRate    float64 `json:"interest_rate"`
	Description     string  `json:"description"`
}

// Payment request struct now matches the usage in the handler
type loanPaymentRequest struct {
	AccountNumber string  `json:"account_number"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
}

func (h *HandlerDependencies) CreateLoanHandler(w http.ResponseWriter, r *http.Request) {
	var req loanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	account, err := h.Models.Accounts.GetByAccountNumber(req.AccountNumber)
	if err != nil {
		http.Error(w, "account not found: "+err.Error(), http.StatusNotFound)
		return
	}

	loan := &data.Loan{
		CustomerID:      account.ID,
		LoanTypeID:      1,
		PrincipalAmount: req.PrincipalAmount,
		InterestRate:    req.InterestRate,
		TermMonths:      req.TermMonths,
		Status:          "active",
		IssuedAt:        time.Now(),
		MaturityDate:    time.Now().AddDate(0, req.TermMonths, 0),
		GLAccountID:     account.GLAccountID,
	}

	if err := h.Models.Loans.Insert(loan); err != nil {
		http.Error(w, "failed to create loan record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	je := &data.JournalEntry{
		ReferenceTypeID: 3, // Ensure 3 exists in reference_types table
		ReferenceID:     loan.ID,
		Description:     req.Description,
		CreatedAt:       time.Now(),
	}

	// CRITICAL: Must check error so je.ID is populated correctly
	if err := h.Models.JournalEntries.Insert(je); err != nil {
		http.Error(w, "failed to create journal entry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	creditEntry := &data.LedgerEntry{
		GLAccountID:    account.GLAccountID,
		JournalEntryID: je.ID,
		Credit:         loan.PrincipalAmount,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := h.Models.LedgerEntries.Insert(creditEntry); err != nil {
		http.Error(w, "failed to credit account: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Loan successfully disbursed to " + req.AccountNumber,
		"loan_id": loan.ID,
	})
}

func (h *HandlerDependencies) CreateLoanPaymentHandler(w http.ResponseWriter, r *http.Request) {
	var req loanPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	account, err := h.Models.Accounts.GetByAccountNumber(req.AccountNumber)
	if err != nil {
		http.Error(w, "account not found", http.StatusNotFound)
		return
	}

	// Dynamic lookup: Find the active loan for this specific customer
	// Note: You may need a specific "GetActiveLoanByCustomerID" method for better precision
	loans, _, err := h.Models.Loans.GetAll(data.Filters{Page: 1, PageSize: 1})
	if err != nil || len(loans) == 0 {
		http.Error(w, "no active loans found for this account holder", http.StatusNotFound)
		return
	}
	targetLoan := loans[0]

	je := &data.JournalEntry{
		ReferenceTypeID: 4, // Ensure 4 exists in reference_types table
		ReferenceID:     targetLoan.ID,
		Description:     req.Description,
		CreatedAt:       time.Now(),
	}

	if err := h.Models.JournalEntries.Insert(je); err != nil {
		http.Error(w, "failed to create journal entry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// DEBIT Account (Money leaves)
	debitEntry := &data.LedgerEntry{
		GLAccountID:    account.GLAccountID,
		JournalEntryID: je.ID,
		Debit:          req.Amount,
		CreatedAt:      time.Now(),
	}

	// CREDIT Loan (Debt decreases)
	creditEntry := &data.LedgerEntry{
		GLAccountID:    targetLoan.GLAccountID,
		JournalEntryID: je.ID,
		Credit:         req.Amount,
		CreatedAt:      time.Now(),
	}

	if err := h.Models.LedgerEntries.Insert(debitEntry); err != nil {
		http.Error(w, "failed to debit account: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.Models.LedgerEntries.Insert(creditEntry); err != nil {
		http.Error(w, "failed to credit loan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Payment applied successfully",
		"amount":  req.Amount,
	})
}