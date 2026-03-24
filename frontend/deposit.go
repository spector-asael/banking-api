package main

import (
	"net/http"
	"strconv"
	"bytes"
	"encoding/json"
	"io"
	"fmt"
	"html/template"
)
var makeDepositTemplate = template.Must(template.New("makeDeposit").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Make a Deposit</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f8f8; }
        .container { max-width: 500px; margin: 60px auto; background: #fff; padding: 32px; border-radius: 8px; box-shadow: 0 2px 8px #ccc; }
        input { width: 100%; padding: 10px; margin: 10px 0; box-sizing: border-box; border: 1px solid #ccc; border-radius: 4px; }
        button { width: 100%; padding: 12px; background: #2e7d32; color: white; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; margin-top: 10px; }
        button:hover { background: #1b5e20; }
        .msg { padding: 12px; border-radius: 4px; margin-bottom: 20px; font-weight: bold; text-align: center; }
        .success { background: #e8f5e9; color: #2e7d32; border: 1px solid #c8e6c9; }
        .error { background: #ffebee; color: #c62828; border: 1px solid #ffcdd2; }
        .back-link { display: block; text-align: center; margin-top: 20px; color: #1976d2; text-decoration: none; }
        .back-link:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Make a Deposit</h1>
        
        {{if .Success}}<div class="msg success">Deposit of ${{.Amount}} was successful!</div>{{end}}
        {{if .Error}}<div class="msg error">{{.Error}}</div>{{end}}

        <form method="POST">
            <label>Bank Account Number:</label>
            <input type="text" name="account_number" placeholder="e.g., SAV20001" required>

            <label>Deposit Amount ($):</label>
            <input type="number" step="0.01" min="0.01" name="amount" placeholder="50.00" required>

            <label>Description / Memo (Optional):</label>
            <input type="text" name="description" placeholder="e.g., Cash Deposit at Teller">

            <button type="submit">Process Deposit</button>
        </form>
        <a href="/" class="back-link">&larr; Back to Home</a>
    </div>
</body>
</html>
`))

func makeDepositHandler(w http.ResponseWriter, r *http.Request) {
	// If it's a GET request, just show the empty form
	if r.Method == http.MethodGet {
		makeDepositTemplate.Execute(w, nil)
		return
	}

	// It's a POST request, so parse the form values
	accountNumber := r.FormValue("account_number")
	amountStr := r.FormValue("amount")
	description := r.FormValue("description")

	// Convert the string amount from the form into a float64 for the API
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		makeDepositTemplate.Execute(w, map[string]interface{}{"Error": "Invalid amount format. Please enter a valid number."})
		return
	}

	// Prepare the JSON payload exactly how our backend expects it
	payload := map[string]interface{}{
		"account_number": accountNumber,
		"amount":         amount,
		"description":    description,
	}

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(payload)

	// Send the POST request to the backend API
	// Note: Make sure the URL matches where your backend deposit route is mapped!
	resp, err := http.Post("http://localhost:4000/api/deposits", "application/json", buf)
	
	if err != nil {
		makeDepositTemplate.Execute(w, map[string]interface{}{"Error": "Could not connect to the backend server."})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		errorMsg := fmt.Sprintf("Deposit failed. Backend says: %s", string(raw))
		makeDepositTemplate.Execute(w, map[string]interface{}{"Error": errorMsg})
		return
	}

	// Success! Re-render the form with a nice success message
	makeDepositTemplate.Execute(w, map[string]interface{}{
		"Success": true,
		"Amount":  fmt.Sprintf("%.2f", amount),
	})
}