package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

var loginTemplate = template.Must(template.New("login").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>System Login</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            background: #f4f7f6; 
            margin: 0; 
            height: 100vh; 
            display: flex; 
            align-items: center; 
            justify-content: center; 
        }
        .card { 
            background: #fff; 
            padding: 30px; 
            border-radius: 8px; 
            box-shadow: 0 4px 15px rgba(0,0,0,0.1); 
            width: 100%; 
            max-width: 400px; 
            box-sizing: border-box;
        }
        h2 { color: #2c3e50; border-bottom: 2px solid #eee; padding-bottom: 10px; margin-top: 0; text-align: center; }
        label { display: block; margin-top: 15px; font-weight: bold; color: #555; }
        input { width: 100%; padding: 10px; margin-top: 5px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
        button { width: 100%; padding: 12px; margin-top: 25px; border: none; border-radius: 4px; color: white; font-size: 16px; cursor: pointer; background: #3498db; }
        button:hover { background: #2980b9; }
        .status-msg { padding: 15px; margin-bottom: 20px; border-radius: 4px; text-align: center; font-weight: bold; }
        .success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
    </style>
</head>
<body>
    <div class="card">
        <h2>Secure Login</h2>

        {{if .SuccessMsg}}<div class="status-msg success">{{.SuccessMsg}}</div>{{end}}
        {{if .ErrorMsg}}<div class="status-msg error">{{.ErrorMsg}}</div>{{end}}

        <form method="POST">
            <label>Email Address</label>
            <input type="email" name="email" placeholder="name@example.com" required>
            
            <label>Password</label>
            <input type="password" name="password" placeholder="••••••••" required>
            
            <button type="submit">Log In</button>
        </form>
    </div>
</body>
</html>
`))

type LoginPageData struct {
	SuccessMsg string
	ErrorMsg   string
}

type TokenResponse struct {
	AuthenticationToken struct {
		Token  string `json:"token"`
		Expiry string `json:"expiry"`
	} `json:"authentication_token"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	data := LoginPageData{}

	if r.Method == http.MethodPost {
		apiURL := "http://localhost:4000/api/authentication/token"

		payload := map[string]interface{}{
			"email":    r.FormValue("email"),
			"password": r.FormValue("password"),
		}

		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(payload)
		resp, err := http.Post(apiURL, "application/json", buf)

		if err != nil {
			data.ErrorMsg = "Authentication server is unreachable."
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {

				var tokenResp TokenResponse
				if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
					data.ErrorMsg = "Failed to parse token from server."
				} else {
					// --- UPDATED: Grab the token from the nested struct ---
					cookie := &http.Cookie{
						Name:     "auth_token",
						Value:    tokenResp.AuthenticationToken.Token,
						Path:     "/",
						HttpOnly: true,
						Secure:   false,
						SameSite: http.SameSiteLaxMode,
					}

					http.SetCookie(w, cookie)
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}

			} else {
				body, _ := io.ReadAll(resp.Body)
				data.ErrorMsg = fmt.Sprintf("Login failed: %s", string(body))
			}
		}
	}

	loginTemplate.Execute(w, data)
}
