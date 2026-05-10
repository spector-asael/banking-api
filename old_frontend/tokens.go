package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
)

var activateUserTemplate = template.Must(template.New("activateUser").Parse(`
<!DOCTYPE html>
<html>
<head><title>Activate Account | Craboo Bank</title>` + layoutCSS + `</head>
<body>
    <div class="container" style="max-width: 450px; text-align: center; margin-top: 50px;">
        <h1 style="color: #1e293b;">Activate Your Account</h1>
        <p style="color: #64748b; margin-bottom: 30px;">Please enter the 26-character activation token sent to your email.</p>
        
        {{if .Error}}<div style="background-color: #fef2f2; color: #ef4444; padding: 10px; border-radius: 6px; margin-bottom:15px;">{{.Error}}</div>{{end}}
        {{if .Message}}
            <div style="background-color: #f0fdf4; color: #22c55e; padding: 15px; border-radius: 6px; margin-bottom:15px;">
                <strong>Success!</strong> {{.Message}}
            </div>
            <a href="/login" class="btn btn-blue" style="display:inline-block; margin-top:15px;">Proceed to Login</a>
        {{else}}
            <form method="POST">
                <div class="form-group">
                    <input type="text" name="token" value="{{.Token}}" required 
                           placeholder="e.g. 6FGIZBSQFVCAPHCQKUYGLVJQSI" 
                           style="text-align: center; font-family: monospace; font-size: 18px; letter-spacing: 2px; text-transform: uppercase;">
                </div>
                
                <button type="submit" class="btn btn-blue" style="width:100%; margin-top: 15px;">Activate Account</button>
            </form>
        {{end}}
    </div>
	<a class="back" href="/">&larr; Back to Home</a>
</body>
</html>
`))

func activateAccountHandler(w http.ResponseWriter, r *http.Request) {
	// 1. If it's a GET request, check the URL for a token and render the form
	if r.Method == http.MethodGet {
		// This allows links like: /activate?token=6FGIZBSQFVCAPHCQKUYGLVJQSI
		urlToken := r.URL.Query().Get("token")
		activateUserTemplate.Execute(w, map[string]any{"Token": urlToken})
		return
	}

	// 2. If it's a POST request, process the form submission
	if r.Method == http.MethodPost {
		token := r.FormValue("token")

		// Build the JSON payload expected by your API
		payload := map[string]string{
			"token": token,
		}
		jsonBody, _ := json.Marshal(payload)

		// Create the PUT request to your internal API
		// Make sure this URL exactly matches your routes!
		req, err := http.NewRequest(http.MethodPut, "http://localhost:4000/api/users/activated", bytes.NewBuffer(jsonBody))
		if err != nil {
			activateUserTemplate.Execute(w, map[string]any{"Error": "Internal server error. Please try again later.", "Token": token})
			return
		}
		req.Header.Set("Content-Type", "application/json")

		// Execute the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			activateUserTemplate.Execute(w, map[string]any{"Error": "Failed to connect to the activation service.", "Token": token})
			return
		}
		defer resp.Body.Close()

		// Handle the API's response
		if resp.StatusCode == http.StatusOK {
			// Success! Don't pass the token back, just the success message.
			activateUserTemplate.Execute(w, map[string]any{"Message": "Your account has been securely activated. Welcome to Craboo Bank!"})
			return
		}

		// If not 200 OK, the token was likely invalid, expired, or already used.
		activateUserTemplate.Execute(w, map[string]any{
			"Error": "Invalid or expired activation token. Please check your email and try again.",
			"Token": token,
		})
		return
	}
}
