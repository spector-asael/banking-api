package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
)

var makeTransferTemplate = template.Must(template.New("makeTransfer").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Internal Transfer</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f8f8; }
        .container { max-width: 500px; margin: 60px auto; background: #fff; padding: 32px; border-radius: 8px; box-shadow: 0 2px 8px #ccc; }
        input { width: 100%; padding: 10px; margin: 10px 0; box-sizing: border-box; border: 1px solid #ccc; border-radius: 4px; }
        /* Using blue for transfers */
        button { width: 100%; padding: 12px; background: #0277bd; color: white; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; margin-top: 10px; }
        button:hover { background: #01579b; }
        .msg { padding: 12px; border-radius: 4px; margin-bottom: 20px; font-weight: bold; text-align: center; }
        .success { background: #e8f5e9; color: #2e7d32; border: 1px solid #c8e6c9; }
        .error { background: #ffebee; color: #c62828; border: 1px solid #ffcdd2; }
        .back-link { display: block; text-align: center; margin-top: 20px; color: #1976d2; text-decoration: none; }
        .back-link:hover { text-decoration: underline; }
        label { font-weight: bold; color: #555; display: block; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Internal Transfer</h1>
        
        {{if .Success}}<div class="msg success">Transfer of ${{.Amount}} was successful!</div>{{end}}
        {{if .Error}}<div class="msg error">{{.Error}}</div>{{end}} <form method="POST">

        <form method="POST">
            <label>From Account (Source):</label>
            <input type="text" name="source_account_number" placeholder="e.g., CHK10001" required>

            <label>To Account (Destination):</label>
            <input type="text" name="destination_account_number" placeholder="e.g., SAV20001" required>

            <label>Transfer Amount ($):</label>
            <input type="number" step="0.01" min="0.01" name="amount" placeholder="100.00" required>

            <label>Description / Memo:</label>
            <input type="text" name="description" placeholder="e.g., Rent money">

            <button type="submit">Execute Transfer</button>
        </form>
        <a href="/" class="back-link">&larr; Back to Home</a>
    </div>
</body>
</html>
`))

func makeTransferHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		makeTransferTemplate.Execute(w, nil)
		return
	}

	// Parse form values
	source := r.FormValue("source_account_number")
	dest := r.FormValue("destination_account_number")
	amountStr := r.FormValue("amount")
	description := r.FormValue("description")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		makeTransferTemplate.Execute(w, map[string]interface{}{"Error": "Invalid amount format."})
		return
	}

	// Payload keys MUST match your backend transferRequest struct
	payload := map[string]interface{}{
		"source_account_number":      source,
		"destination_account_number": dest,
		"amount":                     amount,
		"description":                description,
	}

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(payload)

	// Call the backend API
	resp, err := http.Post("http://localhost:4000/api/transfers", "application/json", buf)
	if err != nil {
		makeTransferTemplate.Execute(w, map[string]interface{}{"Error": "Could not connect to the backend server."})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		errorMsg := fmt.Sprintf("Transfer failed. Backend says: %s", string(raw))
		makeTransferTemplate.Execute(w, map[string]interface{}{"Error": errorMsg})
		return
	}

	makeTransferTemplate.Execute(w, map[string]interface{}{
		"Success": true,
		"Amount":  fmt.Sprintf("%.2f", amount),
	})
}