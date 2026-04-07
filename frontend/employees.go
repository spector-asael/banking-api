package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

// --- Models ---

type Employee struct {
	ID         int64     `json:"id"`
	PersonID   int64     `json:"person_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	BranchID   int64     `json:"branch_id"`
	PositionID int64     `json:"position_id"`
	HireDate   time.Time `json:"hire_date"`
	Status     string    `json:"status"`
}

// FullName is a helper method for the templates
func (e Employee) FullName() string {
	return fmt.Sprintf("%s %s", e.FirstName, e.LastName)
}

type employeesPageData struct {
	Employees  []Employee
	Page       int
	TotalPages int
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
	Sort       string
}

type employeeViewData struct {
	Employee *Employee
	Message  string
	Error    string
}

// --- Templates ---

var employeesListTemplate = template.Must(template.New("employeesList").Parse(`
<!DOCTYPE html>
<html>
<head><title>Employee Directory</title>` + layoutCSS + `</head>
<body>
    <div class="container">
        <div class="controls" style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
            <div style="display: flex; align-items: center; gap: 20px;">
                <h1 style="margin: 0;">Employee Directory</h1>
                <a href="/employees/create" class="btn btn-green" style="text-decoration: none;">+ Add New Employee</a>
            </div>
            <form method="GET" id="sortForm" style="margin: 0;">
                <select name="sort" onchange="this.form.submit()" style="padding: 8px; border-radius: 4px; border: 1px solid #cbd5e1;">
                    <option value="id" {{if eq .Sort "id"}}selected{{end}}>Sort by ID (Asc)</option>
                    <option value="-id" {{if eq .Sort "-id"}}selected{{end}}>Sort by ID (Desc)</option>
                    <option value="first_name" {{if eq .Sort "first_name"}}selected{{end}}>Sort by Name (A-Z)</option>
                    <option value="status" {{if eq .Sort "status"}}selected{{end}}>Sort by Status</option>
                </select>
            </form>
        </div>
        <table>
            <thead>
                <tr><th>ID</th><th>Name</th><th>Branch ID</th><th>Position ID</th><th>Status</th><th>Hire Date</th><th>Actions</th></tr>
            </thead>
            <tbody>
                {{range .Employees}}
                <tr>
                    <td>#{{.ID}}</td>
                    <td><strong>{{.FullName}}</strong></td>
                    <td>{{.BranchID}}</td>
                    <td>{{.PositionID}}</td>
                    <td>
                        <span class="badge {{if eq .Status "active"}}badge-active{{else}}badge-inactive{{end}}">
                            {{.Status}}
                        </span>
                    </td>
                    <td>{{.HireDate.Format "Jan 02, 2006"}}</td>
                    <td><a href="/employees/view?id={{.ID}}" class="btn btn-blue">View</a></td>
                </tr>
                {{else}}
                <tr><td colspan="7" style="text-align:center; padding: 20px;">No employees found.</td></tr>
                {{end}}
            </tbody>
        </table>
        
        <div class="pagination" style="margin-top: 20px;">
            <a href="?page={{.PrevPage}}&sort={{.Sort}}" class="btn btn-outline {{if not .HasPrev}}disabled{{end}}" style="{{if not .HasPrev}}pointer-events:none; opacity:0.5;{{end}}">&larr; Prev</a>
            <span>Page <strong>{{.Page}}</strong> of {{.TotalPages}}</span>
            <a href="?page={{.NextPage}}&sort={{.Sort}}" class="btn btn-outline {{if not .HasNext}}disabled{{end}}" style="{{if not .HasNext}}pointer-events:none; opacity:0.5;{{end}}">Next &rarr;</a>
        </div>
        <a href="/" class="back-home" style="display: inline-block; margin-top: 20px;">&larr; Back to Dashboard Home</a>
    </div>
</body>
</html>
`))

var viewEmployeeTemplate = template.Must(template.New("viewEmployee").Parse(`
<!DOCTYPE html>
<html>
<head><title>Employee Profile</title>` + layoutCSS + `</head>
<body>
    <div class="container" style="max-width: 600px;">
        <h1>Employee Profile</h1>
        {{if .Error}}<div style="color:red; text-align:center; margin-bottom:15px;">{{.Error}}</div>{{end}}
        {{if .Employee}}
            <table>
                <tr><th>Employee ID</th><td>#{{.Employee.ID}}</td></tr>
                <tr><th>Full Name</th><td>{{.Employee.FullName}}</td></tr>
                <tr><th>Linked Person ID</th><td>{{.Employee.PersonID}}</td></tr>
                <tr><th>Branch ID</th><td>{{.Employee.BranchID}}</td></tr>
                <tr><th>Position ID</th><td>{{.Employee.PositionID}}</td></tr>
                <tr><th>Status</th><td>{{.Employee.Status}}</td></tr>
                <tr><th>Hire Date</th><td>{{.Employee.HireDate.Format "January 02, 2006"}}</td></tr>
            </table>
            <div style="text-align: center; margin-top: 30px; display: flex; justify-content: center; gap: 15px;">
                <a href="/employees/edit?id={{.Employee.ID}}" class="btn btn-green">Edit Details</a>
                <form action="/employees/delete?id={{.Employee.ID}}" method="POST" onsubmit="return confirm('Are you sure you want to delete this employee? This will also delete their user account.');">
                    <button type="submit" class="btn btn-outline" style="color: #ef4444; border-color: #ef4444;">Delete Employee</button>
                </form>
            </div>
        {{end}}
        <a href="/employees" class="back-home">&larr; Back to Employee Directory</a>
    </div>
</body>
</html>
`))

var createEmployeeTemplate = template.Must(template.New("createEmployee").Parse(`
<!DOCTYPE html>
<html>
<head><title>Onboard New Employee</title>` + layoutCSS + `</head>
<body>
    <div class="container" style="max-width: 500px;">
        <h1>Onboard New Employee</h1>
        
        {{if .Error}}<div style="color:red; margin-bottom:15px; text-align:center;">{{.Error}}</div>{{end}}
        {{if .Message}}<div style="color:green; margin-bottom:15px; text-align:center;">{{.Message}}</div>{{end}}
        
        <form method="POST">
            <div class="form-group">
                <label>Person SSID</label>
                <input type="text" name="ssid" required placeholder="Must match an existing Person record">
            </div>
            
            <hr style="border: 0; border-top: 1px solid #edf2f7; margin: 20px 0;">
            
            <div class="form-group">
                <label>Username</label>
                <input type="text" name="username" required>
            </div>
            <div class="form-group">
                <label>Email</label>
                <input type="email" name="email" required>
            </div>
            <div class="form-group">
                <label>Temporary Password</label>
                <input type="password" name="password" required>
            </div>
            
            <hr style="border: 0; border-top: 1px solid #edf2f7; margin: 20px 0;">
            
            <div class="form-group">
                <label>Branch ID</label>
                <input type="number" name="branch_id" required>
            </div>
            <div class="form-group">
                <label>System Role</label>
                <select name="account_type" style="width: 100%; padding: 10px; border: 1px solid #cbd5e1; border-radius: 6px;">
                    <option value="1">Teller</option>
                    <option value="2">Customer Service Representative</option>
                    <option value="3">Manager</option>
                    <option value="4">System Admin</option>
                </select>
            </div>
            
            <button type="submit" class="btn btn-blue" style="width:100%; margin-top: 15px;">Create Employee & User Account</button>
        </form>
        <a href="/employees" class="back-home">&larr; Back to Directory</a>
    </div>
</body>
</html>
`))

var editEmployeeTemplate = template.Must(template.New("editEmployee").Parse(`
<!DOCTYPE html>
<html>
<head><title>Edit Employee</title>` + layoutCSS + `</head>
<body>
    <div class="container" style="max-width: 500px;">
        <h1>Edit Employee Details</h1>
        <p style="text-align:center; color:#64748b; margin-bottom:20px;">Updating records for <strong>{{.Employee.FullName}}</strong></p>
        
        {{if .Error}}<div style="color:red; margin-bottom:15px; text-align:center;">{{.Error}}</div>{{end}}
        {{if .Message}}<div style="color:green; margin-bottom:15px; text-align:center;">{{.Message}}</div>{{end}}
        
        <form method="POST">
            <div class="form-group">
                <label>Branch ID</label>
                <input type="number" name="branch_id" value="{{.Employee.BranchID}}" required>
            </div>
            <div class="form-group">
                <label>System Role (Position)</label>
                <select name="position_id" style="width: 100%; padding: 10px; border: 1px solid #cbd5e1; border-radius: 6px;">
                    <option value="1" {{if eq .Employee.PositionID 1}}selected{{end}}>Teller</option>
                    <option value="2" {{if eq .Employee.PositionID 2}}selected{{end}}>Customer Service Representative</option>
                    <option value="3" {{if eq .Employee.PositionID 3}}selected{{end}}>Manager</option>
                    <option value="4" {{if eq .Employee.PositionID 4}}selected{{end}}>System Admin</option>
                </select>
            </div>
            <div class="form-group">
                <label>Status</label>
                <select name="status" style="width: 100%; padding: 10px; border: 1px solid #cbd5e1; border-radius: 6px;">
                    <option value="active" {{if eq .Employee.Status "active"}}selected{{end}}>Active</option>
                    <option value="inactive" {{if eq .Employee.Status "inactive"}}selected{{end}}>Inactive</option>
                    <option value="on leave" {{if eq .Employee.Status "on leave"}}selected{{end}}>On Leave</option>
                    <option value="terminated" {{if eq .Employee.Status "terminated"}}selected{{end}}>Terminated</option>
                </select>
            </div>
            <button type="submit" class="btn btn-blue" style="width:100%; margin-top: 15px;">Update Records</button>
        </form>
        <a href="/employees/view?id={{.Employee.ID}}" class="back-home">&larr; Cancel and Return</a>
    </div>
</body>
</html>
`))

// --- Handlers ---

func employeesHandler(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "id"
	}

	apiURL := fmt.Sprintf("http://localhost:4000/api/employees?page=%d&page_size=10&sort=%s", page, sort)

	// UPDATED: Use callAPI instead of http.Get
	resp, err := callAPI(r, http.MethodGet, apiURL, nil)
	if err != nil {
		if err.Error() == "unauthorized" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.Error(w, "Failed to connect to API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("API returned status: %d", resp.StatusCode), http.StatusInternalServerError)
		return
	}

	var result struct {
		Employees []Employee `json:"employees"`
		Metadata  struct {
			LastPage int `json:"last_page"`
		} `json:"@metadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}

	data := employeesPageData{
		Employees:  result.Employees,
		Page:       page,
		TotalPages: result.Metadata.LastPage,
		PrevPage:   page - 1,
		NextPage:   page + 1,
		HasPrev:    page > 1,
		HasNext:    page < result.Metadata.LastPage,
		Sort:       sort,
	}

	employeesListTemplate.Execute(w, data)
}

func createEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		createEmployeeTemplate.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		branchID, _ := strconv.ParseInt(r.FormValue("branch_id"), 10, 64)
		accountType, _ := strconv.Atoi(r.FormValue("account_type"))

		positionID := int64(accountType)

		payload := map[string]interface{}{
			"ssid":         r.FormValue("ssid"),
			"username":     r.FormValue("username"),
			"email":        r.FormValue("email"),
			"password":     r.FormValue("password"),
			"branch_id":    branchID,
			"position_id":  positionID,
			"account_type": accountType,
		}

		jsonBody, _ := json.Marshal(payload)

		// UPDATED: Use callAPI instead of http.Post
		resp, err := callAPI(r, http.MethodPost, "http://localhost:4000/api/employees", bytes.NewBuffer(jsonBody))

		if err != nil {
			if err.Error() == "unauthorized" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			createEmployeeTemplate.Execute(w, map[string]string{"Error": "Failed to connect to the API."})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			createEmployeeTemplate.Execute(w, map[string]string{"Message": "Success! Employee and associated user account have been created."})
			return
		}

		createEmployeeTemplate.Execute(w, map[string]string{"Error": "Failed to create employee. Ensure the SSID exists and the email is not already in use."})
		return
	}
}

func viewEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		viewEmployeeTemplate.Execute(w, employeeViewData{Error: "Employee ID is required."})
		return
	}

	// UPDATED: Use callAPI instead of http.Get
	resp, err := callAPI(r, http.MethodGet, "http://localhost:4000/api/employees/"+id, nil)
	if err != nil {
		if err.Error() == "unauthorized" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		viewEmployeeTemplate.Execute(w, employeeViewData{Error: "Failed to reach API."})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		viewEmployeeTemplate.Execute(w, employeeViewData{Error: "Employee not found."})
		return
	}

	var result struct {
		Employee Employee `json:"employee"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	viewEmployeeTemplate.Execute(w, employeeViewData{Employee: &result.Employee})
}

func editEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	// UPDATED: Use callAPI to fetch existing data
	resp, err := callAPI(r, http.MethodGet, "http://localhost:4000/api/employees/"+id, nil)
	if err != nil {
		if err.Error() == "unauthorized" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.Error(w, "Failed to reach API", http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	var result struct {
		Employee Employee `json:"employee"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if r.Method == http.MethodPost {
		branchID, _ := strconv.ParseInt(r.FormValue("branch_id"), 10, 64)
		positionID, _ := strconv.ParseInt(r.FormValue("position_id"), 10, 64)
		status := r.FormValue("status")

		payload := map[string]interface{}{
			"branch_id":   branchID,
			"position_id": positionID,
			"status":      status,
		}

		jsonBody, _ := json.Marshal(payload)

		// UPDATED: Use callAPI instead of http.NewRequest / client.Do
		respPatch, err := callAPI(r, http.MethodPatch, "http://localhost:4000/api/employees/"+id, bytes.NewBuffer(jsonBody))

		if err != nil && err.Error() == "unauthorized" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if err == nil && respPatch.StatusCode == http.StatusOK {
			respPatch.Body.Close()
			result.Employee.BranchID = branchID
			result.Employee.PositionID = positionID
			result.Employee.Status = status
			editEmployeeTemplate.Execute(w, employeeViewData{Employee: &result.Employee, Message: "Employee records updated successfully!"})
			return
		}

		if respPatch != nil {
			respPatch.Body.Close()
		}
		editEmployeeTemplate.Execute(w, employeeViewData{Employee: &result.Employee, Error: "Failed to update employee records. Check input."})
		return
	}

	editEmployeeTemplate.Execute(w, employeeViewData{Employee: &result.Employee})
}

func deleteEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")

	// UPDATED: Use callAPI instead of http.NewRequest / client.Do
	resp, err := callAPI(r, http.MethodDelete, "http://localhost:4000/api/employees/"+id, nil)

	if err != nil {
		if err.Error() == "unauthorized" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.Error(w, "Failed to connect to API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		http.Redirect(w, r, "/employees", http.StatusSeeOther)
		return
	}

	http.Error(w, "Failed to delete employee", http.StatusInternalServerError)
}
