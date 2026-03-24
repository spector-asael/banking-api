package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "html/template"
    "io"
    "net/http"
    "strconv"
    "time"
)

// --- Data Structures ---

type Account struct {
    ID            int64  `json:"id"`
    AccountNumber string `json:"account_number"`
    Status        string `json:"status"`
    OpenedAt      string `json:"opened_at"`
    ClosedAt      string `json:"closed_at,omitempty"`
}

type accountsPageData struct {
    Accounts   []Account
    Page       int
    TotalPages int
    PrevPage   int
    NextPage   int
    HasPrev    bool
    HasNext    bool
    Error      string
}

var accountsTemplate = template.Must(template.New("accounts").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Accounts List</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f8f8; }
        .container { max-width: 1000px; margin: 60px auto; background: #fff; padding: 32px; border-radius: 8px; box-shadow: 0 2px 8px #ccc; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { padding: 12px; border-bottom: 1px solid #eee; text-align: left; }
        th { background: #f0f0f0; }
        .status-badge { padding: 4px 8px; border-radius: 4px; font-size: 0.85em; background: #e8f5e9; color: #2e7d32; }
        .btn { background: #1976d2; color: white; padding: 8px 16px; text-decoration: none; border-radius: 4px; }
        .view-btn { background: #43a047; }
        .pagination { display: flex; justify-content: center; gap: 16px; margin-top: 24px; }
    </style>
</head>
<body>
    <div class="container">
        <div style="display: flex; justify-content: space-between; align-items: center;">
            <h1>Accounts</h1>
            <a href="/accounts/create" class="btn">+ Open New Account</a>
        </div>
        <table>
            <tr><th>ID</th><th>Account #</th><th>Status</th><th>Opened At</th><th>Actions</th></tr>
            {{range .Accounts}}
            <tr>
                <td>{{.ID}}</td>
                <td>{{.AccountNumber}}</td>
                <td><span class="status-badge">{{.Status}}</span></td>
                <td>{{.OpenedAt}}</td>
                <td><a href="/accounts/view?id={{.ID}}" class="btn view-btn">View</a></td>
            </tr>
            {{else}}
            <tr><td colspan="5" style="text-align:center;">No accounts found.</td></tr>
            {{end}}
        </table>
        <div class="pagination">
            {{if .HasPrev}}<a href="/accounts?page={{.PrevPage}}">Prev</a>{{end}}
            <span>Page {{.Page}} of {{.TotalPages}}</span>
            {{if .HasNext}}<a href="/accounts?page={{.NextPage}}">Next</a>{{end}}
        </div>
        <a href="/" style="display:block; margin-top:30px; text-align:center;">&larr; Back to Home</a>
    </div>
</body>
</html>
`))

var createAccountTemplate = template.Must(template.New("createAccount").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Open Account</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f8f8; }
        .container { max-width: 500px; margin: 60px auto; background: #fff; padding: 32px; border-radius: 8px; }
        input, select { width: 100%; padding: 10px; margin: 10px 0; box-sizing: border-box; }
        button { width: 100%; padding: 12px; background: #1976d2; color: white; border: none; cursor: pointer; }
        .err { color: red; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Open Bank Account</h1>
        {{if .Error}}<div class="err">{{.Error}}</div>{{end}}
        <form method="POST">
            <label>Customer SSID:</label>
            <input type="text" name="ssid" required>
            
            <label>Account Number (Manual generation for now):</label>
            <input type="text" name="account_number" placeholder="SAV20001" required>

            <label>Branch ID:</label>
            <input type="number" name="branch_id" value="1" required>

            <label>Account Type ID (e.g., 1 for Checking):</label>
            <input type="number" name="account_type_id" value="1" required>

            <button type="submit">Open Account</button>
        </form>
        <a href="/accounts" style="display:block; text-align:center; margin-top:20px;">Cancel</a>
    </div>
</body>
</html>
`))

var viewAccountTemplate = template.Must(template.New("viewAccount").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>View Account</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f8f8; }
        .container { max-width: 600px; margin: 60px auto; background: #fff; padding: 32px; border-radius: 8px; }
        .info-row { display: flex; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid #eee; }
        .label { font-weight: bold; }
        .delete-btn { background: #d32f2f; color: white; padding: 10px; border: none; width: 100%; margin-top: 20px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Account Details</h1>
        <div class="info-row"><span class="label">ID:</span> <span>{{.ID}}</span></div>
        <div class="info-row"><span class="label">Account Number:</span> <span>{{.AccountNumber}}</span></div>
        <div class="info-row"><span class="label">Status:</span> <span>{{.Status}}</span></div>
        <div class="info-row"><span class="label">Opened:</span> <span>{{.OpenedAt}}</span></div>
        
        <button class="delete-btn" onclick="deleteAccount({{.ID}})">Delete Account</button>
        <a href="/accounts" style="display:block; text-align:center; margin-top:20px;">&larr; Back to List</a>
    </div>
    <script>
    function deleteAccount(id) {
        if(!confirm('Delete this account?')) return;
        fetch('http://localhost:4000/api/accounts/' + id, { method: 'DELETE' })
        .then(resp => resp.ok ? window.location='/accounts' : alert('Delete failed'));
    }
    </script>
</body>
</html>
`))

func accountsHandler(w http.ResponseWriter, r *http.Request) {
    page := 1
    if p := r.URL.Query().Get("page"); p != "" {
        if val, err := strconv.Atoi(p); err == nil { page = val }
    }

    apiURL := fmt.Sprintf("http://localhost:4000/api/accounts?page=%d&page_size=5", page)
    resp, err := http.Get(apiURL)
    if err != nil {
        accountsTemplate.Execute(w, accountsPageData{Error: "Backend API unreachable"})
        return
    }
    defer resp.Body.Close()

    var result struct {
        Accounts []Account `json:"accounts"`
        Metadata struct {
            CurrentPage int `json:"current_page"`
            LastPage    int `json:"last_page"`
        } `json:"@metadata"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    data := accountsPageData{
        Accounts:   result.Accounts,
        Page:       result.Metadata.CurrentPage,
        TotalPages: result.Metadata.LastPage,
        PrevPage:   result.Metadata.CurrentPage - 1,
        NextPage:   result.Metadata.CurrentPage + 1,
        HasPrev:    result.Metadata.CurrentPage > 1,
        HasNext:    result.Metadata.CurrentPage < result.Metadata.LastPage,
    }

    accountsTemplate.Execute(w, data)
}

func createAccountHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        createAccountTemplate.Execute(w, nil)
        return
    }

    // Parse form values
    branchID, _ := strconv.ParseInt(r.FormValue("branch_id"), 10, 64)
    accountTypeID, _ := strconv.ParseInt(r.FormValue("account_type_id"), 10, 64)
    
    // Format OpenedAt to RFC3339 for the backend
    openedAt := time.Now().Format(time.RFC3339)

    // GL Account ID is no longer sent to the backend!
    payload := map[string]interface{}{
        "account_number":         r.FormValue("account_number"),
        "branch_id_opened_at":    branchID,
        "account_type_id":        accountTypeID,
        "status":                 "active",
        "opened_at":              openedAt,
        "social_security_number": r.FormValue("ssid"),
        "is_joint_account":       false,
    }

    buf := new(bytes.Buffer)
    json.NewEncoder(buf).Encode(payload)

    resp, err := http.Post("http://localhost:4000/api/accounts", "application/json", buf)
    if err != nil || resp.StatusCode != http.StatusCreated {
        raw, _ := io.ReadAll(resp.Body)
        createAccountTemplate.Execute(w, map[string]string{"Error": string(raw)})
        return
    }

    http.Redirect(w, r, "/accounts", http.StatusSeeOther)
}

func viewAccountHandler(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    resp, err := http.Get("http://localhost:4000/api/accounts/" + id)
    if err != nil || resp.StatusCode != http.StatusOK {
        http.Redirect(w, r, "/accounts", http.StatusSeeOther)
        return
    }
    defer resp.Body.Close()

    var result struct { Account Account `json:"account"` }
    json.NewDecoder(resp.Body).Decode(&result)
    viewAccountTemplate.Execute(w, result.Account)
}

