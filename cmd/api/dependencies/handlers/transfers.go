package handlers

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/spector-asael/banking-api/internal/data"
    "github.com/spector-asael/banking-api/internal/validator"
)

// 1. Switched to strings for Account Numbers
type transferRequest struct {
    SourceAccountNumber      string  `json:"source_account_number"`
    DestinationAccountNumber string  `json:"destination_account_number"`
    Amount                   float64 `json:"amount"`
    Description              string  `json:"description"`
}

func (h *HandlerDependencies) HandleTransfer(w http.ResponseWriter, r *http.Request) {
    var req transferRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    // 2. Updated validation for strings
    v := validator.New()
    v.Check(req.SourceAccountNumber != "", "source_account_number", "must be provided")
    v.Check(req.DestinationAccountNumber != "", "destination_account_number", "must be provided")
    v.Check(req.Amount > 0, "amount", "must be greater than zero")
    v.Check(req.SourceAccountNumber != req.DestinationAccountNumber, "accounts", "source and destination must be different")
    if !v.IsEmpty() {
        w.WriteHeader(http.StatusUnprocessableEntity)
        json.NewEncoder(w).Encode(v.Errors)
        return
    }

    // 3. Fetch by Account Number and expose REAL errors
    sourceAccount, err := h.Models.Accounts.GetByAccountNumber(req.SourceAccountNumber)
    if err != nil {
        http.Error(w, "source account error: " + err.Error(), http.StatusInternalServerError)
        return
    }
    
    destAccount, err := h.Models.Accounts.GetByAccountNumber(req.DestinationAccountNumber)
    if err != nil {
        http.Error(w, "destination account error: " + err.Error(), http.StatusInternalServerError)
        return
    }

    // Create JournalEntry for the transfer
    je := &data.JournalEntry{
        ReferenceTypeID: 2, // 2 = transfer
        ReferenceID:     sourceAccount.ID, 
        Description:     req.Description,
        CreatedAt:       time.Now(),
    }
    if err := h.Models.JournalEntries.Insert(je); err != nil {
        http.Error(w, "could not create journal entry: " + err.Error(), http.StatusInternalServerError)
        return
    }

    // Create LedgerEntry: Debit source (decrease balance), Credit destination (increase balance)
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
        http.Error(w, "could not create debit ledger entry: " + err.Error(), http.StatusInternalServerError)
        return
    }
    if err := h.Models.LedgerEntries.Insert(creditEntry); err != nil {
        http.Error(w, "could not create credit ledger entry: " + err.Error(), http.StatusInternalServerError)
        return
    }

    // (Assuming you have AccountTransactions setup in your data models!)
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
        http.Error(w, "could not create source account transaction: " + err.Error(), http.StatusInternalServerError)
        return
    }
    if err := h.Models.AccountTransactions.Insert(dstTx); err != nil {
        http.Error(w, "could not create dest account transaction: " + err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message":             "Transfer successful",
        "journal_entry":       je,
        "debit_ledger_entry":  debitEntry,
        "credit_ledger_entry": creditEntry,
        "source_transaction":  srcTx,
        "destination_transaction": dstTx,
    })
}