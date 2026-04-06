package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// --- Data Structures ---

type Customer struct {
	ID                   int64  `json:"id"`
	PersonID             int64  `json:"person_id"`
	FirstName            string `json:"first_name"` // From backend JOIN
	LastName             string `json:"last_name"`  // From backend JOIN
	SocialSecurityNumber string `json:"social_security_number"`
	KYCStatusID          int    `json:"kyc_status_id"`
	CreatedAt            string `json:"created_at"`
}

type customersPageData struct {
	Customers  []Customer
	Page       int
	TotalPages int
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
	Error      string
}

// --- HTML Templates ---

var customersTemplate = template.Must(template.New("customers").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Customers List</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f8f8; margin: 0; padding: 0; }
        .container { max-width: 1000px; margin: 60px auto; background: #fff; border-radius: 8px; box-shadow: 0 2px 8px #ccc; padding: 32px; }
        h1 { text-align: center; }
        table { width: 100%; border-collapse: collapse; margin-top: 32px; }
        th, td { padding: 12px 8px; border-bottom: 1px solid #eee; text-align: left; }
        th { background: #f0f0f0; }
        tr:hover { background: #f5faff; }
        .kyc-badge { padding: 4px 8px; border-radius: 4px; font-size: 0.85em; background: #e3f2fd; color: #1976d2; font-weight: bold; }
        .pagination { display: flex; justify-content: center; align-items: center; gap: 16px; margin-top: 24px; }
        .pagination a { padding: 8px 18px; background: #1976d2; color: #fff; border-radius: 4px; text-decoration: none; font-size: 0.9em; }
        .pagination a.disabled { background: #ccc; pointer-events: none; }
        .view-btn { background: #43a047; color: #fff; border: none; border-radius: 4px; padding: 6px 16px; cursor: pointer; text-decoration: none; font-size: 0.9em; }
        .back { margin-top: 32px; display: block; text-align: center; color: #666; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <div style="display: flex; justify-content: space-between; align-items: center;">
            <h1 style="margin: 0;">Customers</h1>
            <a href="/customers/create" style="background: #1976d2; color: #fff; padding: 10px 22px; border-radius: 4px; text-decoration: none;">+ Register Customer</a>
        </div>
        <table>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>First Name</th>
                    <th>Last Name</th>
                    <th>KYC Status</th>
                    <th>Created At</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                {{range .Customers}}
                <tr>
                    <td>{{.ID}}</td>
                    <td>{{.FirstName}}</td>
                    <td>{{.LastName}}</td>
                    <td><span class="kyc-badge">Level {{.KYCStatusID}}</span></td>
                    <td>{{.CreatedAt}}</td>
                    <td><a href="/customers/view?id={{.ID}}" class="view-btn">View</a></td>
                </tr>
                {{else}}
                <tr><td colspan="6" style="text-align:center; color: #888; padding: 20px;">No customers found.</td></tr>
                {{end}}
            </tbody>
        </table>
        <div class="pagination">
            <a href="/customers?page={{.PrevPage}}" class="{{if not .HasPrev}}disabled{{end}}">&larr; Prev</a>
            <span>Page {{.Page}} of {{.TotalPages}}</span>
            <a href="/customers?page={{.NextPage}}" class="{{if not .HasNext}}disabled{{end}}">Next &rarr;</a>
        </div>
        <a class="back" href="/">&larr; Back to Home</a>
    </div>
</body>
</html>
`))

var viewCustomerTemplate = template.Must(template.New("viewCustomer").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Customer Profile</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f8f8; }
        .container { max-width: 600px; margin: 60px auto; background: #fff; padding: 32px; border-radius: 8px; box-shadow: 0 2px 8px #ccc; }
        .info-row { display: flex; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid #eee; }
        .label { font-weight: bold; color: #555; }
        .actions { margin-top: 32px; display: flex; gap: 10px; }
        .btn { padding: 10px 20px; border-radius: 4px; border: none; cursor: pointer; color: white; text-decoration: none; text-align: center; flex: 1; }
        .btn-edit { background: #1976d2; }
        .btn-delete { background: #d32f2f; }
        .back { margin-top: 24px; display: block; text-align: center; color: #1976d2; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Customer Profile</h1>
        <div class="info-row"><span class="label">Customer ID:</span> <span>{{.ID}}</span></div>
        <div class="info-row"><span class="label">Name:</span> <span>{{.FirstName}} {{.LastName}}</span></div>
        <div class="info-row"><span class="label">Person ID:</span> <span>{{.PersonID}}</span></div>
        <div class="info-row"><span class="label">KYC Status:</span> <span>Level {{.KYCStatusID}}</span></div>
        <div class="info-row"><span class="label">Registered:</span> <span>{{.CreatedAt}}</span></div>
        
        <div class="actions">
            <a href="/customers/kyc?id={{.ID}}" class="btn btn-edit">Update KYC</a>
            <button onclick="deleteCustomer({{.ID}})" class="btn btn-delete">Delete Customer</button>
        </div>
        <a href="/customers" class="back">&larr; Back to List</a>
    </div>
    <script>
    function deleteCustomer(id) {
        if(!confirm('Delete this customer record?')) return;
        fetch('http://localhost:4000/api/customers/' + id, { method: 'DELETE' })
        .then(resp => resp.ok ? window.location='/customers' : alert('Delete failed'));
    }
    </script>
</body>
</html>
`))

var createCustomerTemplate = template.Must(template.New("createCustomer").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Register Customer</title>
    <style>
        body { font-family: Arial; background: #f8f8f8; }
        .container { max-width: 500px; margin: 60px auto; background: #fff; padding: 32px; border-radius: 8px; box-shadow: 0 2px 8px #ccc; }
        input { width: 100%; padding: 10px; margin: 10px 0; border: 1px solid #ccc; border-radius: 4px; }
        button { width: 100%; padding: 12px; background: #1976d2; color: white; border: none; border-radius: 4px; cursor: pointer; }
        .err { color: #d32f2f; background: #ffebee; padding: 10px; border-radius: 4px; margin-bottom: 10px; display:none; }
    </style>
</head>
<body>
<div class="container">
    <h1>Register New Customer</h1>

    <div class="err" id="errorBox"></div>

    <form id="customerForm">
        <label>Person SSID:</label>
        <input type="text" name="ssid" required>

        <label>Username:</label>
        <input type="text" name="username" required>

        <label>Email:</label>
        <input type="email" name="email" required>

        <label>Password:</label>
        <input type="password" name="password" required>

        <button type="submit">Create Customer</button>
    </form>

    <a href="/customers" style="display:block; text-align:center; margin-top:20px;">Cancel</a>
</div>

<script>
document.getElementById("customerForm").addEventListener("submit", async function(e) {
    e.preventDefault();

    const errorBox = document.getElementById("errorBox");
    errorBox.style.display = "none";
    errorBox.innerHTML = "";

    const data = {
        ssid: document.querySelector('[name="ssid"]').value,
        username: document.querySelector('[name="username"]').value,
        email: document.querySelector('[name="email"]').value,
        password: document.querySelector('[name="password"]').value
    };

    try {
        const res = await fetch("http://localhost:4000/api/customers", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(data)
        });

        const result = await res.json();

        if (!res.ok) {
            errorBox.style.display = "block";
            if (result.errors) {
                errorBox.innerHTML = Object.values(result.errors).join("<br>");
            } else if (typeof result.error === "string") {
                errorBox.innerText = result.error;
            } else if (typeof result.error === "object" && result.error !== null) {
                errorBox.innerHTML = Object.values(result.error).join("<br>");
            } else {
                errorBox.innerText = "Something went wrong";
            }
            return;
        }

        // Success
        window.location.href = "/customers";

    } catch (err) {
        errorBox.style.display = "block";
        errorBox.innerText = "Network error";
    }
});
</script>

</body>
</html>
`))

var updateKYCTemplate = template.Must(template.New("updateKYC").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Update KYC</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f8f8; }
        .container { max-width: 500px; margin: 60px auto; background: #fff; padding: 32px; border-radius: 8px; box-shadow: 0 2px 8px #ccc; }
        input { width: 100%; padding: 10px; margin: 10px 0; box-sizing: border-box; border: 1px solid #ccc; border-radius: 4px; }
        button { width: 100%; padding: 12px; background: #43a047; color: white; border: none; border-radius: 4px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Update KYC Status</h1>
        <p>Customer: <strong>{{.FirstName}} {{.LastName}}</strong> (ID: {{.ID}})</p>
        <form method="POST">
            <label>KYC Level:</label>
            <input type="number" name="kyc_status_id" value="{{.KYCStatusID}}" required>
            <button type="submit">Save Changes</button>
        </form>
        <a href="/customers/view?id={{.ID}}" style="display:block; text-align:center; margin-top:20px; color: #666; text-decoration: none;">Back</a>
    </div>
</body>
</html>
`))

// --- Handlers ---

func customersHandler(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			page = val
		}
	}

	apiURL := fmt.Sprintf("http://localhost:4000/api/customers?page=%d&page_size=5", page)
	resp, err := http.Get(apiURL)
	if err != nil {
		customersTemplate.Execute(w, customersPageData{Error: "Backend API unreachable"})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Customers []Customer `json:"customers"`
		Metadata  struct {
			CurrentPage int `json:"current_page"`
			LastPage    int `json:"last_page"`
		} `json:"@metadata"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	data := customersPageData{
		Customers:  result.Customers,
		Page:       result.Metadata.CurrentPage,
		TotalPages: result.Metadata.LastPage,
		PrevPage:   result.Metadata.CurrentPage - 1,
		NextPage:   result.Metadata.CurrentPage + 1,
		HasPrev:    result.Metadata.CurrentPage > 1,
		HasNext:    result.Metadata.CurrentPage < result.Metadata.LastPage,
	}

	customersTemplate.Execute(w, data)
}

func viewCustomerHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	resp, err := http.Get("http://localhost:4000/api/customers/" + id)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Redirect(w, r, "/customers", http.StatusSeeOther)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Customer Customer `json:"customer"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	viewCustomerTemplate.Execute(w, result.Customer)
}

func createCustomerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	createCustomerTemplate.Execute(w, nil)
}

func updateKYCHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if r.Method == http.MethodGet {
		resp, _ := http.Get("http://localhost:4000/api/customers/" + id)
		var result struct {
			Customer Customer `json:"customer"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		updateKYCTemplate.Execute(w, result.Customer)
		return
	}

	kycID, _ := strconv.Atoi(r.FormValue("kyc_status_id"))
	payload := map[string]int{"kyc_status_id": kycID}
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(payload)

	req, _ := http.NewRequest(http.MethodPatch, "http://localhost:4000/api/customers/"+id+"/kyc-status", buf)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		http.Redirect(w, r, "/customers/view?id="+id, http.StatusSeeOther)
		return
	}
	http.Error(w, "Update failed", http.StatusInternalServerError)
}
