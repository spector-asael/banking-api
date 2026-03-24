
package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/validator"
)

type loanRequest struct {
	CustomerID      int64   `json:"customer_id"`
	LoanTypeID      int64   `json:"loan_type_id"`
	PrincipalAmount float64 `json:"principal_amount"`
	InterestRate    float64 `json:"interest_rate"`
	TermMonths      int     `json:"term_months"`
	Status          string  `json:"status"`
	IssuedAt        string  `json:"issued_at"`
	MaturityDate    string  `json:"maturity_date"`
	GLAccountID     int64   `json:"gl_account_id"`
	DisbursementGLAccountID int64 `json:"disbursement_gl_account_id"`
	Description     string  `json:"description"`
}

// POST /loans
func (h *HandlerDependencies) createLoanHandler(w http.ResponseWriter, r *http.Request) {
	var req loanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	v := validator.New()
	data.ValidateLoan(v, &data.Loan{
		CustomerID:      req.CustomerID,
		LoanTypeID:      req.LoanTypeID,
		PrincipalAmount: req.PrincipalAmount,
		InterestRate:    req.InterestRate,
		TermMonths:      req.TermMonths,
		Status:          req.Status,
		GLAccountID:     req.GLAccountID,
	})
	if req.IssuedAt == "" {
		v.AddError("issued_at", "must be provided")
	}
	if req.MaturityDate == "" {
		v.AddError("maturity_date", "must be provided")
	}
	if req.DisbursementGLAccountID <= 0 {
		v.AddError("disbursement_gl_account_id", "must be provided and valid")
	}
	if !v.IsEmpty() {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(v.Errors)
		return
	}

	issuedAt, err := time.Parse(time.RFC3339, req.IssuedAt)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"issued_at": "must be a valid RFC3339 timestamp"})
		return
	}
	maturityDate, err := time.Parse(time.RFC3339, req.MaturityDate)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"maturity_date": "must be a valid RFC3339 timestamp"})
		return
	}

	loan := &data.Loan{
		CustomerID:      req.CustomerID,
		LoanTypeID:      req.LoanTypeID,
		PrincipalAmount: req.PrincipalAmount,
		InterestRate:    req.InterestRate,
		TermMonths:      req.TermMonths,
		Status:          req.Status,
		IssuedAt:        issuedAt,
		MaturityDate:    maturityDate,
		GLAccountID:     req.GLAccountID,
	}
	if err := h.Models.Loans.Insert(loan); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "could not create loan"})
		return
	}

	// Create JournalEntry for loan disbursement
	je := &data.JournalEntry{
		ReferenceTypeID: 4, // 4 = loan disbursement (adjust as needed)
		ReferenceID:     loan.ID,
		Description:     req.Description,
		CreatedAt:       time.Now(),
	}
	if err := h.Models.JournalEntries.Insert(je); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "could not create journal entry"})
		return
	}

	// Ledger: Debit Loan Receivable (loan.GLAccountID), Credit Disbursement GL (req.DisbursementGLAccountID)
	debitEntry := &data.LedgerEntry{
		GLAccountID:    loan.GLAccountID,
		JournalEntryID: je.ID,
		Debit:          loan.PrincipalAmount,
		Credit:         0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	creditEntry := &data.LedgerEntry{
		GLAccountID:    req.DisbursementGLAccountID,
		JournalEntryID: je.ID,
		Debit:          0,
		Credit:         loan.PrincipalAmount,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := h.Models.LedgerEntries.Insert(debitEntry); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "could not create debit ledger entry"})
		return
	}
	if err := h.Models.LedgerEntries.Insert(creditEntry); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "could not create credit ledger entry"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"loan": loan,
		"journal_entry": je,
		"debit_ledger_entry": debitEntry,
		"credit_ledger_entry": creditEntry,
	})
}