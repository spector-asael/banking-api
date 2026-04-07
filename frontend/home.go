package main

import (
	"html/template"
	"net/http"
)

var homeTemplate = template.Must(template.New("home").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Banking System - Home</title>
	<style>
		body { font-family: Arial, sans-serif; background: #f8f8f8; margin: 0; padding: 0; }
		.container { max-width: 600px; margin: 60px auto; background: #fff; border-radius: 8px; box-shadow: 0 2px 8px #ccc; padding: 32px; }
		h1 { text-align: center; }
		ul { list-style: none; padding: 0; }
		li { margin: 18px 0; }
		a { text-decoration: none; color: #1976d2; font-size: 1.2em; }
		a:hover { text-decoration: underline; }
	</style>
</head>
<body>
	<div class="container">
		<h1>Banking System Portal</h1>
		<ul>
			<li><a href="/persons">Manage Persons</a></li>
			<li><a href="/customers">Manage Customers</a></li>
			<li><a href="/employees">Manage Employees</a></li>
			<li><a href="/users">Manage Users</a></li>
			<li><a href="/accounts">Manage Accounts</a></li>
			<li><a href="/deposits">Deposits</a></li>
			<li><a href="/withdrawals">Withdrawals</a></li>
			<li><a href="/loans">Loans</a></li>
			<li><a href="/transfers">Transfers</a></li>
			<li><a href="/login">Login</a></li>
		</ul>
	</div>
</body>
</html>
`))

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := homeTemplate.Execute(w, nil); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
