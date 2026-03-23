package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/validator"
)

type transferRequest struct {
	SourceAccountID      int64   `json:"source_account_id"`
	DestinationAccountID int64   `json:"destination_account_id"`
	Amount               float64 `json:"amount"`
	Description          string  `json:"description"`
}

func (h *HandlerDependencies) HandleTransfer(w http.ResponseWriter, r *http.Request) {
	var req transferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	v := validator.New()
	v.Check(req.SourceAccountID > 0, "source_account_id", "must be provided and valid")
	v.Check(req.DestinationAccountID > 0, "destination_account_id", "must be provided and valid")
	v.Check(req.Amount > 0, "amount", "must be greater than zero")
	v.Check(req.SourceAccountID != req.DestinationAccountID, "accounts", "source and destination must be different")
	if !v.IsEmpty() {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(v.Errors)
		return
	}

	sourceAccount, err := h.Models.Accounts.GetByID(req.SourceAccountID)
	if err != nil {
		http.Error(w, "source account not found", http.StatusNotFound)
		return
	}
	destAccount, err := h.Models.Accounts.GetByID(req.DestinationAccountID)
	if err != nil {
		http.Error(w, "destination account not found", http.StatusNotFound)
		return
	}

	// Create JournalEntry for the transfer
	je := &data.JournalEntry{
		ReferenceTypeID: 2, // 2 = transfer (adjust as needed)
		ReferenceID:     sourceAccount.ID, // or a new transfer id if you have one
		Description:     req.Description,
		CreatedAt:       time.Now(),
	}
	if err := h.Models.JournalEntries.Insert(je); err != nil {
		http.Error(w, "could not create journal entry", http.StatusInternalServerError)
		return
	}

	// Create LedgerEntry: Debit source, Credit destination
	debitEntry := &data.LedgerEntry{
		GLAccountID:    sourceAccount.GLAccountID,
		JournalEntryID: je.ID,
		Debit:          req.Amount,
		Credit:         0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	creditEntry := &data.LedgerEntry{
		GLAccountID:    destAccount.GLAccountID,
		JournalEntryID: je.ID,
		Debit:          0,
		Credit:         req.Amount,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := h.Models.LedgerEntries.Insert(debitEntry); err != nil {
		http.Error(w, "could not create debit ledger entry", http.StatusInternalServerError)
		return
	}
	if err := h.Models.LedgerEntries.Insert(creditEntry); err != nil {
		http.Error(w, "could not create credit ledger entry", http.StatusInternalServerError)
		return
	}

	// Create AccountTransaction for both accounts
	srcTx := &data.AccountTransaction{
		AccountID:          sourceAccount.ID,
		CounterpartyAcctID: &destAccount.ID,
		JournalEntryID:     je.ID,
		Amount:             -req.Amount,
		CreatedAt:          time.Now(),
	}
	dstTx := &data.AccountTransaction{
		AccountID:          destAccount.ID,
		CounterpartyAcctID: &sourceAccount.ID,
		JournalEntryID:     je.ID,
		Amount:             req.Amount,
		CreatedAt:          time.Now(),
	}
	if err := h.Models.AccountTransactions.Insert(srcTx); err != nil {
		http.Error(w, "could not create source account transaction", http.StatusInternalServerError)
		return
	}
	if err := h.Models.AccountTransactions.Insert(dstTx); err != nil {
		http.Error(w, "could not create destination account transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"journal_entry": je,
		"debit_ledger_entry":  debitEntry,
		"credit_ledger_entry": creditEntry,
		"source_transaction":  srcTx,
		"destination_transaction": dstTx,
	})
}