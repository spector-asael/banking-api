package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// --- Models ---

type User struct {
	ID          int64  `json:"id"`
	Username    string `json:"name"`
	Email       string `json:"email"`
	Activated   bool   `json:"activated"`
	AccountType int    `json:"account_type"`
}

// AccountTypeName translates the ID to the labels you provided.
func (u User) AccountTypeName() string {
	names := map[int]string{
		0: "CUSTOMER",
		1: "TELLER",
		2: "CUSTOMER SERVICE REP",
		3: "MANAGER",
		4: "ADMIN",
	}
	if name, ok := names[u.AccountType]; ok {
		return name
	}
	return "UNKNOWN"
}

type usersPageData struct {
	Users      []User
	Page       int
	TotalPages int
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
	Sort       string
}

type userViewData struct {
	User    *User
	Message string
	Error   string
}

// --- Templates ---

var layoutCSS = `
    <style>
        body { font-family: 'Segoe UI', sans-serif; background: #f4f7f6; margin: 0; padding: 0; color: #333; }
        .container { max-width: 1000px; margin: 40px auto; background: #fff; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.08); padding: 32px; }
        h1 { color: #2c3e50; margin-bottom: 24px; text-align: center; }
        
        /* Table & Sorting */
        .controls { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 14px; text-align: left; border-bottom: 1px solid #edf2f7; }
        th { background: #f8fafc; color: #64748b; font-size: 0.8rem; text-transform: uppercase; }
        
        /* Badges & Buttons */
        .badge { padding: 4px 10px; border-radius: 20px; font-size: 0.75rem; font-weight: 600; }
        .badge-active { background: #dcfce7; color: #166534; }
        .badge-inactive { background: #fee2e2; color: #991b1b; }
        .badge-type { background: #e0f2fe; color: #0369a1; }
        
        .btn { padding: 8px 16px; border-radius: 6px; cursor: pointer; text-decoration: none; font-size: 0.9rem; border: none; transition: 0.2s; display: inline-block; }
        .btn-blue { background: #3b82f6; color: white; }
        .btn-blue:hover { background: #2563eb; }
        .btn-green { background: #10b981; color: white; }
        .btn-green:hover { background: #059669; }
        .btn-outline { background: transparent; border: 1px solid #cbd5e1; color: #64748b; }
        
        /* Pagination */
        .pagination { display: flex; justify-content: center; align-items: center; gap: 15px; margin-top: 30px; }
        .pagination button:disabled { opacity: 0.5; cursor: not-allowed; }
        
        /* Forms */
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: 600; }
        input { width: 100%; padding: 10px; border: 1px solid #cbd5e1; border-radius: 6px; box-sizing: border-box; }
        
        .back-home { display: block; text-align: center; margin-top: 20px; color: #64748b; text-decoration: none; }
		.back { margin-top: 24px; display: block; text-align: center; color: #1976d2; text-decoration: none; }
    </style>
`

var usersListTemplate = template.Must(template.New("usersList").Parse(`
<!DOCTYPE html>
<html>
<head><title>User List</title>` + layoutCSS + `</head>
<body>
    <div class="container">
        <div class="controls">
            <h1>System Users</h1>
            <form method="GET" id="sortForm">
                <select name="sort" onchange="this.form.submit()" style="padding: 8px; border-radius: 4px;">
                    <option value="id" {{if eq .Sort "id"}}selected{{end}}>Sort by ID (Asc)</option>
                    <option value="-id" {{if eq .Sort "-id"}}selected{{end}}>Sort by ID (Desc)</option>
                    <option value="name" {{if eq .Sort "name"}}selected{{end}}>Sort by Name (A-Z)</option>
                    <option value="-name" {{if eq .Sort "-name"}}selected{{end}}>Sort by Name (Z-A)</option>
                </select>
            </form>
        </div>
        <table>
            <thead>
                <tr><th>ID</th><th>Username</th><th>Email</th><th>Status</th><th>Type</th><th>Actions</th></tr>
            </thead>
            <tbody>
                {{range .Users}}
                <tr>
                    <td>#{{.ID}}</td>
                    <td><strong>{{.Username}}</strong></td>
                    <td>{{.Email}}</td>
                    <td><span class="badge {{if .Activated}}badge-active{{else}}badge-inactive{{end}}">{{if .Activated}}Active{{else}}Inactive{{end}}</span></td>
                    <td><span class="badge badge-type">{{.AccountTypeName}}</span></td>
                    <td><a href="/users/view?id={{.ID}}" class="btn btn-blue">View</a></td>
                </tr>
                {{end}}
            </tbody>
        </table>
        <div class="pagination">
            <a href="?page={{.PrevPage}}&sort={{.Sort}}" class="btn btn-outline {{if not .HasPrev}}disabled{{end}}" style="{{if not .HasPrev}}pointer-events:none;{{end}}">&larr; Prev</a>
            <span>Page <strong>{{.Page}}</strong> of {{.TotalPages}}</span>
            <a href="?page={{.NextPage}}&sort={{.Sort}}" class="btn btn-outline {{if not .HasNext}}disabled{{end}}" style="{{if not .HasNext}}pointer-events:none;{{end}}">Next &rarr;</a>
        </div>
        <a href="/" class="back-home">&larr; Back to Dashboard Home</a>
    </div>
</body>
</html>
`))

var viewUserTemplate = template.Must(template.New("viewUser").Parse(`
<!DOCTYPE html>
<html>
<head><title>User Profile</title>` + layoutCSS + `</head>
<body>
    <div class="container" style="max-width: 600px;">
        <h1>User Profile</h1>
        {{if .User}}
            <table>
                <tr><th>Username</th><td>{{.User.Username}}</td></tr>
                <tr><th>Email</th><td>{{.User.Email}}</td></tr>
                <tr><th>Account Type</th><td>{{.User.AccountTypeName}}</td></tr>
                <tr><th>Status</th><td>{{if .User.Activated}}Active{{else}}Inactive{{end}}</td></tr>
            </table>
            <div style="text-align: center; margin-top: 30px;">
                <a href="/users/edit?id={{.User.ID}}" class="btn btn-green">Edit User Details</a>
            </div>
        {{end}}
        <a href="/users" class="back-home">&larr; Back to User List</a>
    </div>
</body>
</html>
`))

var editUserTemplate = template.Must(template.New("editUser").Parse(`
<!DOCTYPE html>
<html>
<head><title>Edit User</title>` + layoutCSS + `</head>
<body>
    <div class="container" style="max-width: 500px;">
        <h1>Edit User</h1>
        {{if .Error}}<div style="color:red; margin-bottom:15px;">{{.Error}}</div>{{end}}
        {{if .Message}}<div style="color:green; margin-bottom:15px;">{{.Message}}</div>{{end}}
        
        <form method="POST">
            <div class="form-group">
                <label>Username</label>
                <input type="text" name="username" value="{{.User.Username}}" required>
            </div>
            <div class="form-group">
                <label>Email</label>
                <input type="email" name="email" value="{{.User.Email}}" required>
            </div>
            <div class="form-group">
                <label>New Password (leave blank to keep current)</label>
                <input type="password" name="password">
            </div>
            <button type="submit" class="btn btn-blue" style="width:100%;">Update Profile</button>
        </form>
        <a href="/users/view?id={{.User.ID}}" class="back-home">&larr; Cancel and Return</a>
    </div>
</body>
</html>
`))

// --- Handlers ---

func usersHandler(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		val, _ := strconv.Atoi(p)
		if val > 0 {
			page = val
		}
	}
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "id"
	}

	apiURL := fmt.Sprintf("http://localhost:4000/api/users?page=%d&page_size=10&sort=%s", page, sort)
	resp, err := http.Get(apiURL)
	if err != nil {
		http.Error(w, "Failed to connect to API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Users    []User `json:"users"`
		Metadata struct {
			LastPage int `json:"last_page"`
		} `json:"@metadata"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	data := usersPageData{
		Users:      result.Users,
		Page:       page,
		TotalPages: result.Metadata.LastPage,
		PrevPage:   page - 1,
		NextPage:   page + 1,
		HasPrev:    page > 1,
		HasNext:    page < result.Metadata.LastPage,
		Sort:       sort,
	}
	usersListTemplate.Execute(w, data)
}

func viewUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	resp, _ := http.Get("http://localhost:4000/api/users/" + id)
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	defer resp.Body.Close()

	var result struct {
		User User `json:"user"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	viewUserTemplate.Execute(w, userViewData{User: &result.User})
}

func editUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	// Fetch existing data for the form
	resp, _ := http.Get("http://localhost:4000/api/users/" + id)
	var result struct {
		User User `json:"user"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if r.Method == http.MethodPost {
		payload := map[string]string{
			"name":  r.FormValue("username"),
			"email": r.FormValue("email"),
		}
		if pwd := r.FormValue("password"); pwd != "" {
			payload["password"] = pwd
		}

		jsonBody, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPatch, "http://localhost:4000/api/users/"+id, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		respPatch, err := client.Do(req)

		if err == nil && respPatch.StatusCode == http.StatusOK {
			// Update local struct to show success on form
			result.User.Username = payload["name"]
			result.User.Email = payload["email"]
			editUserTemplate.Execute(w, userViewData{User: &result.User, Message: "User updated successfully!"})
			return
		}
		editUserTemplate.Execute(w, userViewData{User: &result.User, Error: "Update failed. Please check inputs."})
		return
	}

	editUserTemplate.Execute(w, userViewData{User: &result.User})
}
