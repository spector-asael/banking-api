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

var loansTemplate = template.Must(template.New("loans").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Loan Management</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f4f7f6; padding: 20px; }
        .container { max-width: 900px; margin: 0 auto; }
        .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 30px; }
        .card { background: #fff; padding: 25px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h2 { color: #2c3e50; border-bottom: 2px solid #eee; padding-bottom: 10px; margin-top: 0; }
        label { display: block; margin-top: 15px; font-weight: bold; color: #555; }
        input { width: 100%; padding: 10px; margin-top: 5px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
        button { width: 100%; padding: 12px; margin-top: 20px; border: none; border-radius: 4px; color: white; font-size: 16px; cursor: pointer; }
        .btn-disburse { background: #27ae60; }
        .btn-payment { background: #2980b9; }
        .status-msg { padding: 15px; margin-bottom: 20px; border-radius: 4px; text-align: center; font-weight: bold; }
        .success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .back-link { display: block; text-align: center; margin-top: 30px; color: #3498db; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Loan Central</h1>

        {{if .SuccessMsg}}<div class="status-msg success">{{.SuccessMsg}}</div>{{end}}
        {{if .ErrorMsg}}<div class="status-msg error">{{.ErrorMsg}}</div>{{end}}

        <div class="grid">
            <div class="card">
                <h2>New Loan</h2>
                <form method="POST">
                    <input type="hidden" name="form_type" value="disburse">
                    <label>Account Number</label>
                    <input type="text" name="account_number" placeholder="CHK10001" required>
                    <label>Principal Amount ($)</label>
                    <input type="number" step="0.01" name="amount" required>
                    <label>Interest Rate (%)</label>
                    <input type="number" step="0.01" name="rate" value="5.0">
                    <label>Term (Months)</label>
                    <input type="number" name="term" value="12">
                    <label>Description</label>
                    <input type="text" name="description" placeholder="Vehicle Loan">
                    <button type="submit" class="btn-disburse">Disburse Funds</button>
                </form>
            </div>

            <div class="card">
                <h2>Make Payment</h2>
                <form method="POST">
                    <input type="hidden" name="form_type" value="payment">
                    <label>Account Number</label>
                    <input type="text" name="account_number" placeholder="CHK10001" required>
                    <label>Payment Amount ($)</label>
                    <input type="number" step="0.01" name="amount" required>
                    <label>Memo</label>
                    <input type="text" name="description" placeholder="Monthly Installment">
                    <button type="submit" class="btn-payment">Post Payment</button>
                </form>
            </div>
        </div>
        <a href="/" class="back-link">&larr; Back to Dashboard</a>
    </div>
</body>
</html>
`))

type LoanPageData struct {
	SuccessMsg string
	ErrorMsg   string
}

func loansHandler(w http.ResponseWriter, r *http.Request) {
	data := LoanPageData{}

	if r.Method == http.MethodPost {
		formType := r.FormValue("form_type")
		var apiURL string
		var payload map[string]interface{}

		switch formType {
		case "disburse":
			apiURL = "http://localhost:4000/api/loans"
			amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
			rate, _ := strconv.ParseFloat(r.FormValue("rate"), 64)
			term, _ := strconv.Atoi(r.FormValue("term"))

			payload = map[string]interface{}{
				"account_number":   r.FormValue("account_number"),
				"principal_amount": amount,
				"interest_rate":    rate,
				"term_months":      term,
				"description":      r.FormValue("description"),
			}
		case "payment":
			apiURL = "http://localhost:4000/api/loans/payments"
			amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)

			payload = map[string]interface{}{
				"account_number": r.FormValue("account_number"),
				"amount":         amount,
				"description":    r.FormValue("description"),
			}
		}

		// Send to Backend
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(payload)
		resp, err := http.Post(apiURL, "application/json", buf)

		if err != nil {
			data.ErrorMsg = "Backend server is unreachable."
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
				data.SuccessMsg = "Transaction processed successfully!"
			} else {
				body, _ := io.ReadAll(resp.Body)
				data.ErrorMsg = fmt.Sprintf("Error: %s", string(body))
			}
		}
	}

	loansTemplate.Execute(w, data)
}
