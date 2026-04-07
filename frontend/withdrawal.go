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

var makeWithdrawalTemplate = template.Must(template.New("makeWithdrawal").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Make a Withdrawal</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f8f8; }
        .container { max-width: 500px; margin: 60px auto; background: #fff; padding: 32px; border-radius: 8px; box-shadow: 0 2px 8px #ccc; }
        input { width: 100%; padding: 10px; margin: 10px 0; box-sizing: border-box; border: 1px solid #ccc; border-radius: 4px; }
        /* Changed button color to orange/red for withdrawals! */
        button { width: 100%; padding: 12px; background: #e65100; color: white; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; margin-top: 10px; }
        button:hover { background: #bf360c; }
        .msg { padding: 12px; border-radius: 4px; margin-bottom: 20px; font-weight: bold; text-align: center; }
        .success { background: #e8f5e9; color: #2e7d32; border: 1px solid #c8e6c9; }
        .error { background: #ffebee; color: #c62828; border: 1px solid #ffcdd2; }
        .back-link { display: block; text-align: center; margin-top: 20px; color: #1976d2; text-decoration: none; }
        .back-link:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Make a Withdrawal</h1>
        
        {{if .Success}}<div class="msg success">Withdrawal of ${{.Amount}} was successful!</div>{{end}}
        {{if .Error}}<div class="msg error">{{.Error}}</div>{{end}}

        <form method="POST">
            <label>Bank Account Number:</label>
            <input type="text" name="account_number" placeholder="e.g., SAV20001" required>

            <label>Withdrawal Amount ($):</label>
            <input type="number" step="0.01" min="0.01" name="amount" placeholder="50.00" required>

            <label>Description / Memo (Optional):</label>
            <input type="text" name="description" placeholder="e.g., ATM Cash Withdrawal">

            <button type="submit">Process Withdrawal</button>
        </form>
        <a href="/" class="back-link">&larr; Back to Home</a>
    </div>
</body>
</html>
`))

func makeWithdrawalHandler(w http.ResponseWriter, r *http.Request) {
	// If GET, show the empty form
	if r.Method == http.MethodGet {
		makeWithdrawalTemplate.Execute(w, nil)
		return
	}

	// If POST, parse the form values
	accountNumber := r.FormValue("account_number")
	amountStr := r.FormValue("amount")
	description := r.FormValue("description")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		makeWithdrawalTemplate.Execute(w, map[string]interface{}{"Error": "Invalid amount format."})
		return
	}

	// Prepare the JSON payload
	payload := map[string]interface{}{
		"account_number": accountNumber,
		"amount":         amount,
		"description":    description,
	}

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(payload)

	// --- UPDATED: Use callAPI to securely pass the auth token ---
	resp, err := callAPI(r, http.MethodPost, "http://localhost:4000/api/withdrawals", buf)

	if err != nil {
		// Redirect if the token is missing or invalid
		if err.Error() == "unauthorized" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		makeWithdrawalTemplate.Execute(w, map[string]interface{}{"Error": "Could not connect to the backend server."})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		errorMsg := fmt.Sprintf("Withdrawal failed. Backend says: %s", string(raw))
		makeWithdrawalTemplate.Execute(w, map[string]interface{}{"Error": errorMsg})
		return
	}

	// Success! Re-render the form with a success message
	makeWithdrawalTemplate.Execute(w, map[string]interface{}{
		"Success": true,
		"Amount":  fmt.Sprintf("%.2f", amount),
	})
}
