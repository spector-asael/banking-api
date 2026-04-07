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

var personsTemplate = template.Must(template.New("persons").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Persons List</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f8f8; margin: 0; padding: 0; }
        .container { max-width: 900px; margin: 60px auto; background: #fff; border-radius: 8px; box-shadow: 0 2px 8px #ccc; padding: 32px; }
        h1 { text-align: center; }
        table { width: 100%; border-collapse: collapse; margin-top: 32px; }
        th, td { padding: 12px 8px; border-bottom: 1px solid #eee; text-align: left; }
        th { background: #f0f0f0; }
        tr:hover { background: #f5faff; }
        .pagination { display: flex; justify-content: center; align-items: center; gap: 16px; margin-top: 24px; }
        .pagination button { padding: 8px 18px; background: #1976d2; color: #fff; border: none; border-radius: 4px; cursor: pointer; font-size: 1em; }
        .pagination button:disabled { background: #ccc; cursor: not-allowed; }
        .view-btn { background: #43a047; color: #fff; border: none; border-radius: 4px; padding: 6px 16px; cursor: pointer; }
        .view-btn:hover { background: #2e7031; }
        .back { margin-top: 32px; display: block; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div style="display: flex; justify-content: space-between; align-items: center;">
            <h1 style="margin: 0;">Persons</h1>
            <a href="/persons/create" style="background: #1976d2; color: #fff; padding: 10px 22px; border-radius: 4px; text-decoration: none; font-size: 1em;">+ Create New Entry</a>
        </div>
        <table>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>First Name</th>
                    <th>Last Name</th>
                    <th>SSID</th>
                    <th>Email</th>
                    <th>Phone</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                {{range .Persons}}
                <tr>
                    <td>{{.ID}}</td>
                    <td>{{.FirstName}}</td>
                    <td>{{.LastName}}</td>
                    <td>{{.SocialSecurityNumber}}</td>
                    <td>{{.Email}}</td>
                    <td>{{.PhoneNumber}}</td>
                    <td><form method="GET" action="/persons/view"><input type="hidden" name="ssid" value="{{.SocialSecurityNumber}}"><button class="view-btn" type="submit">View</button></form></td>
                </tr>
                {{else}}
                <tr><td colspan="7" style="text-align:center; color:#888;">No persons found.</td></tr>
                {{end}}
            </tbody>
        </table>
        <div class="pagination">
            <form method="GET" style="display:inline;">
                <input type="hidden" name="page" value="{{.PrevPage}}">
                <button type="submit" {{if not .HasPrev}}disabled{{end}}>&larr; Prev</button>
            </form>
            <span>Page {{.Page}} of {{.TotalPages}}</span>
            <form method="GET" style="display:inline;">
                <input type="hidden" name="page" value="{{.NextPage}}">
                <button type="submit" {{if not .HasNext}}disabled{{end}}>Next &rarr;</button>
            </form>
        </div>
        <a class="back" href="/">&larr; Back to Home</a>
    </div>
</body>
</html>
`))

type Person struct {
	ID                   int
	FirstName            string
	LastName             string
	SocialSecurityNumber string
	Email                string
	PhoneNumber          string
}

type personsPageData struct {
	Persons    []Person
	Page       int
	TotalPages int
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
}

func personsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	pageSize := 5

	apiURL := fmt.Sprintf("http://localhost:4000/api/persons?page=%d&page_size=%d", page, pageSize)

	// UPDATED: Use callAPI
	resp, err := callAPI(r, http.MethodGet, apiURL, nil)
	if err != nil {
		if err.Error() == "unauthorized" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.Error(w, "Failed to fetch persons", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Persons []struct {
			ID                   int    `json:"id"`
			FirstName            string `json:"first_name"`
			LastName             string `json:"last_name"`
			SocialSecurityNumber string `json:"social_security_number"`
			Email                string `json:"email"`
			PhoneNumber          string `json:"phone_number"`
		} `json:"persons"`
		Metadata struct {
			CurrentPage int `json:"current_page"`
			LastPage    int `json:"last_page"`
		} `json:"@metadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "Failed to parse API response", http.StatusInternalServerError)
		return
	}

	persons := make([]Person, len(result.Persons))
	for i, p := range result.Persons {
		persons[i] = Person{
			ID:                   p.ID,
			FirstName:            p.FirstName,
			LastName:             p.LastName,
			SocialSecurityNumber: p.SocialSecurityNumber,
			Email:                p.Email,
			PhoneNumber:          p.PhoneNumber,
		}
	}

	data := personsPageData{
		Persons:    persons,
		Page:       result.Metadata.CurrentPage,
		TotalPages: result.Metadata.LastPage,
		PrevPage:   result.Metadata.CurrentPage - 1,
		NextPage:   result.Metadata.CurrentPage + 1,
		HasPrev:    result.Metadata.CurrentPage > 1,
		HasNext:    result.Metadata.CurrentPage < result.Metadata.LastPage,
	}

	if err := personsTemplate.Execute(w, data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

var createPersonTemplate = template.Must(template.New("createPerson").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Create New Person</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f8f8; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 60px auto; background: #fff; border-radius: 8px; box-shadow: 0 2px 8px #ccc; padding: 32px; }
        h1 { text-align: center; }
        form { margin-top: 24px; }
        label { display: block; margin-top: 12px; }
        input[type=text], input[type=date], input[type=email] { width: 100%; padding: 8px; margin-top: 4px; border-radius: 4px; border: 1px solid #ccc; }
        button { margin-top: 18px; padding: 10px 24px; background: #1976d2; color: #fff; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #125ea2; }
        .back { margin-top: 24px; display: block; text-align: center; }
        .msg { text-align: center; margin-top: 18px; color: #388e3c; }
        .err { text-align: center; margin-top: 18px; color: #d32f2f; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Create New Person</h1>
        {{if .Message}}<div class="msg">{{.Message}}</div>{{end}}
        {{if .Error}}<div class="err">{{.Error}}</div>{{end}}
        <form method="POST" action="/persons/create">
            <label>First Name:<input type="text" name="first_name" required></label>
            <label>Last Name:<input type="text" name="last_name" required></label>
            <label>Social Security Number:<input type="text" name="social_security_number" required></label>
            <label>Email:<input type="email" name="email" required></label>
            <label>Date of Birth:<input type="date" name="date_of_birth" required></label>
            <label>Phone Number:<input type="text" name="phone_number" required></label>
            <label>Living Address:<input type="text" name="living_address" required></label>
            <button type="submit">Create</button>
        </form>
        <a class="back" href="/persons">&larr; Back to Persons</a>
    </div>
</body>
</html>
`))

func createPersonHandler(w http.ResponseWriter, r *http.Request) {
	type formData struct {
		FirstName            string
		LastName             string
		SocialSecurityNumber string
		Email                string
		DateOfBirth          string
		PhoneNumber          string
		LivingAddress        string
	}
	type pageData struct {
		Message string
		Error   string
	}
	if r.Method == http.MethodGet {
		createPersonTemplate.Execute(w, pageData{})
		return
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			createPersonTemplate.Execute(w, pageData{Error: "Invalid form submission."})
			return
		}
		data := formData{
			FirstName:            r.FormValue("first_name"),
			LastName:             r.FormValue("last_name"),
			SocialSecurityNumber: r.FormValue("social_security_number"),
			Email:                r.FormValue("email"),
			DateOfBirth:          r.FormValue("date_of_birth"),
			PhoneNumber:          r.FormValue("phone_number"),
			LivingAddress:        r.FormValue("living_address"),
		}

		dobRFC3339 := data.DateOfBirth
		if data.DateOfBirth != "" {
			if t, err := time.Parse("2006-01-02", data.DateOfBirth); err == nil {
				dobRFC3339 = t.Format(time.RFC3339)
			}
		}
		payload := map[string]string{
			"first_name":             data.FirstName,
			"last_name":              data.LastName,
			"social_security_number": data.SocialSecurityNumber,
			"email":                  data.Email,
			"date_of_birth":          dobRFC3339,
			"phone_number":           data.PhoneNumber,
			"living_address":         data.LivingAddress,
		}
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(payload); err != nil {
			createPersonTemplate.Execute(w, pageData{Error: "Failed to encode data."})
			return
		}

		// UPDATED: Use callAPI
		resp, err := callAPI(r, http.MethodPost, "http://localhost:4000/api/persons", buf)
		if err != nil {
			if err.Error() == "unauthorized" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			createPersonTemplate.Execute(w, pageData{Error: "Failed to contact backend API."})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			raw, _ := io.ReadAll(resp.Body)
			var errResp struct {
				Error string `json:"error"`
			}
			msg := ""
			if err := json.Unmarshal(raw, &errResp); err == nil && errResp.Error != "" {
				msg = errResp.Error
			} else {
				msg = string(raw)
			}
			if msg == "" {
				msg = "Backend error."
			}
			fmt.Println("[DEBUG] Backend error response:", msg)
			createPersonTemplate.Execute(w, pageData{Error: msg})
			return
		}
		createPersonTemplate.Execute(w, pageData{Message: "Person created successfully!"})
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// UPDATED: Replaced JS Fetch directly to the backend with a standard Form POST to our Go app
var viewPersonTemplate = template.Must(template.New("viewPerson").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>View Person</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f8f8; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 60px auto; background: #fff; border-radius: 8px; box-shadow: 0 2px 8px #ccc; padding: 32px; }
        h1 { text-align: center; }
        .info-table { width: 100%; border-collapse: collapse; margin-top: 24px; }
        .info-table th, .info-table td { padding: 10px 8px; border-bottom: 1px solid #eee; text-align: left; }
        .info-table th { background: #f0f0f0; width: 180px; }
        .actions { display: flex; gap: 18px; justify-content: center; margin-top: 32px; }
        .edit-btn { background: #1976d2; color: #fff; border: none; border-radius: 4px; padding: 10px 24px; cursor: pointer; }
        .edit-btn:hover { background: #125ea2; }
        .delete-btn { background: #d32f2f; color: #fff; border: none; border-radius: 4px; padding: 10px 24px; cursor: pointer; }
        .delete-btn:hover { background: #a31515; }
        .back { margin-top: 32px; display: block; text-align: center; }
        .msg { text-align: center; margin-top: 18px; color: #388e3c; }
        .err { text-align: center; margin-top: 18px; color: #d32f2f; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Person Details</h1>
        {{if .Error}}<div class="err">{{.Error}}</div>{{end}}
        {{if .Message}}<div class="msg">{{.Message}}</div>{{end}}
        {{if .Person}}
        <table class="info-table">
            <tr><th>ID</th><td>{{.Person.ID}}</td></tr>
            <tr><th>First Name</th><td>{{.Person.FirstName}}</td></tr>
            <tr><th>Last Name</th><td>{{.Person.LastName}}</td></tr>
            <tr><th>Social Security Number</th><td>{{.Person.SocialSecurityNumber}}</td></tr>
            <tr><th>Email</th><td>{{.Person.Email}}</td></tr>
            <tr><th>Date of Birth</th><td>{{.Person.DateOfBirth}}</td></tr>
            <tr><th>Phone Number</th><td>{{.Person.PhoneNumber}}</td></tr>
            <tr><th>Living Address</th><td>{{.Person.LivingAddress}}</td></tr>
            <tr><th>Created At</th><td>{{.Person.CreatedAt}}</td></tr>
            <tr><th>Updated At</th><td>{{.Person.UpdatedAt}}</td></tr>
        </table>
        <div class="actions">
            <form method="GET" action="/persons/edit" style="display:inline;">
                <input type="hidden" name="ssid" value="{{.Person.SocialSecurityNumber}}">
                <button class="edit-btn" type="submit">Edit</button>
            </form>
            <form action="/persons/delete?ssid={{.Person.SocialSecurityNumber}}" method="POST" onsubmit="return confirm('Are you sure you want to delete this person?');" style="display:inline;">
                <button type="submit" class="delete-btn">Delete</button>
            </form>
        </div>
        {{end}}
        <a class="back" href="/persons">&larr; Back to Persons</a>
    </div>
</body>
</html>
`))

type PersonDetails struct {
	ID                   int    `json:"id"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
	SocialSecurityNumber string `json:"social_security_number"`
	Email                string `json:"email"`
	DateOfBirth          string `json:"date_of_birth"`
	PhoneNumber          string `json:"phone_number"`
	LivingAddress        string `json:"living_address"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}

type viewPersonData struct {
	Person  *PersonDetails
	Message string
	Error   string
}

func viewPersonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	ssid := r.URL.Query().Get("ssid")
	if ssid == "" {
		viewPersonTemplate.Execute(w, viewPersonData{Error: "No SSID provided."})
		return
	}

	apiURL := fmt.Sprintf("http://localhost:4000/api/persons/%s", ssid)

	// UPDATED: Use callAPI
	resp, err := callAPI(r, http.MethodGet, apiURL, nil)
	if err != nil {
		if err.Error() == "unauthorized" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		viewPersonTemplate.Execute(w, viewPersonData{Error: "Failed to contact backend API."})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		viewPersonTemplate.Execute(w, viewPersonData{Error: "Person not found or API error."})
		return
	}

	var result struct {
		Person PersonDetails `json:"person"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		viewPersonTemplate.Execute(w, viewPersonData{Error: "Failed to parse API response."})
		return
	}
	data := viewPersonData{Person: &result.Person}
	viewPersonTemplate.Execute(w, data)
}

var editPersonTemplate = template.Must(template.New("editPerson").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Edit Person</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f8f8f8; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 60px auto; background: #fff; border-radius: 8px; box-shadow: 0 2px 8px #ccc; padding: 32px; }
        h1 { text-align: center; }
        form { margin-top: 24px; }
        label { display: block; margin-top: 12px; }
        input[type=text], input[type=date], input[type=email] { width: 100%; padding: 8px; margin-top: 4px; border-radius: 4px; border: 1px solid #ccc; }
        button { margin-top: 18px; padding: 10px 24px; background: #1976d2; color: #fff; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #125ea2; }
        .back { margin-top: 24px; display: block; text-align: center; }
        .msg { text-align: center; margin-top: 18px; color: #388e3c; }
        .err { text-align: center; margin-top: 18px; color: #d32f2f; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Edit Person</h1>
        {{if .Message}}<div class="msg">{{.Message}}</div>{{end}}
        {{if .Error}}<div class="err">{{.Error}}</div>{{end}}
        {{if .Person}}
        <form method="POST">
            <label>First Name:<input type="text" name="first_name" value="{{.Person.FirstName}}" required></label>
            <label>Last Name:<input type="text" name="last_name" value="{{.Person.LastName}}" required></label>
            <label>Social Security Number:<input type="text" name="social_security_number" value="{{.Person.SocialSecurityNumber}}" readonly></label>
            <label>Email:<input type="email" name="email" value="{{.Person.Email}}" required></label>
            <label>Date of Birth:<input type="date" name="date_of_birth" value="{{.Person.DateOfBirth}}" required></label>
            <label>Phone Number:<input type="text" name="phone_number" value="{{.Person.PhoneNumber}}" required></label>
            <label>Living Address:<input type="text" name="living_address" value="{{.Person.LivingAddress}}" required></label>
            <button type="submit">Update</button>
        </form>
        {{end}}
        <a class="back" href="/persons/view?ssid={{.Person.SocialSecurityNumber}}">&larr; Back to Person</a>
    </div>
</body>
</html>
`))

func editPersonHandler(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		Person  *PersonDetails
		Message string
		Error   string
	}
	ssid := r.URL.Query().Get("ssid")
	if ssid == "" {
		editPersonTemplate.Execute(w, pageData{Error: "No SSID provided."})
		return
	}
	if r.Method == http.MethodGet {
		apiURL := fmt.Sprintf("http://localhost:4000/api/persons/%s", ssid)

		// UPDATED: Use callAPI
		resp, err := callAPI(r, http.MethodGet, apiURL, nil)
		if err != nil {
			if err.Error() == "unauthorized" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			editPersonTemplate.Execute(w, pageData{Error: "Failed to contact backend API."})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			editPersonTemplate.Execute(w, pageData{Error: "Person not found or API error."})
			return
		}
		var result struct {
			Person PersonDetails `json:"person"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			editPersonTemplate.Execute(w, pageData{Error: "Failed to parse API response."})
			return
		}
		if len(result.Person.DateOfBirth) >= 10 {
			result.Person.DateOfBirth = result.Person.DateOfBirth[:10]
		}
		editPersonTemplate.Execute(w, pageData{Person: &result.Person})
		return
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			editPersonTemplate.Execute(w, pageData{Error: "Invalid form submission."})
			return
		}
		payload := map[string]string{
			"first_name":     r.FormValue("first_name"),
			"last_name":      r.FormValue("last_name"),
			"email":          r.FormValue("email"),
			"date_of_birth":  r.FormValue("date_of_birth"),
			"phone_number":   r.FormValue("phone_number"),
			"living_address": r.FormValue("living_address"),
		}
		if dob := payload["date_of_birth"]; dob != "" {
			if t, err := time.Parse("2006-01-02", dob); err == nil {
				payload["date_of_birth"] = t.Format(time.RFC3339)
			}
		}
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(payload); err != nil {
			editPersonTemplate.Execute(w, pageData{Error: "Failed to encode data."})
			return
		}

		// UPDATED: Use callAPI instead of client.Do
		patchURL := fmt.Sprintf("http://localhost:4000/api/persons/%s", ssid)
		resp, err := callAPI(r, http.MethodPatch, patchURL, buf)
		if err != nil {
			if err.Error() == "unauthorized" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			editPersonTemplate.Execute(w, pageData{Error: "Failed to contact backend API."})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			raw, _ := io.ReadAll(resp.Body)
			msg := string(raw)
			if msg == "" {
				msg = "Backend error."
			}
			editPersonTemplate.Execute(w, pageData{Error: msg})
			return
		}

		// Success: fetch updated data for display
		apiURL := fmt.Sprintf("http://localhost:4000/api/persons/%s", ssid)
		resp2, err := callAPI(r, http.MethodGet, apiURL, nil)
		if err != nil {
			editPersonTemplate.Execute(w, pageData{Message: "Person updated!", Error: "(But failed to reload data)"})
			return
		}
		defer resp2.Body.Close()

		var result struct {
			Person PersonDetails `json:"person"`
		}
		if err := json.NewDecoder(resp2.Body).Decode(&result); err != nil {
			editPersonTemplate.Execute(w, pageData{Message: "Person updated!", Error: "(But failed to parse updated data)"})
			return
		}
		if len(result.Person.DateOfBirth) >= 10 {
			result.Person.DateOfBirth = result.Person.DateOfBirth[:10]
		}
		editPersonTemplate.Execute(w, pageData{Person: &result.Person, Message: "Person updated!"})
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// NEW: Handler to securely pass delete requests to the backend with the token
func deletePersonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ssid := r.URL.Query().Get("ssid")
	if ssid == "" {
		http.Error(w, "SSID is required", http.StatusBadRequest)
		return
	}

	apiURL := fmt.Sprintf("http://localhost:4000/api/persons/%s", ssid)
	resp, err := callAPI(r, http.MethodDelete, apiURL, nil)

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
		http.Redirect(w, r, "/persons", http.StatusSeeOther)
		return
	}

	body, _ := io.ReadAll(resp.Body)
	http.Error(w, fmt.Sprintf("Failed to delete person: %s", string(body)), http.StatusInternalServerError)
}
