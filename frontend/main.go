package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/persons", personsHandler)
	http.HandleFunc("/persons/create", createPersonHandler)
	http.HandleFunc("/persons/view", viewPersonHandler)
	http.HandleFunc("/persons/edit", editPersonHandler)
	http.HandleFunc("/persons/delete", deletePersonHandler)

	http.HandleFunc("/customers", customersHandler)
	http.HandleFunc("/customers/create", createCustomerHandler)
	http.HandleFunc("/customers/view", viewCustomerHandler)
	http.HandleFunc("/customers/kyc", updateKYCHandler)

	http.HandleFunc("/employees", employeesHandler)
	http.HandleFunc("/employees/create", createEmployeeHandler)
	http.HandleFunc("/employees/view", viewEmployeeHandler)
	http.HandleFunc("/employees/edit", editEmployeeHandler)
	http.HandleFunc("/employees/delete", deleteEmployeeHandler)

	http.HandleFunc("/accounts", accountsHandler)
	http.HandleFunc("/accounts/create", createAccountHandler)
	http.HandleFunc("/accounts/view", viewAccountHandler)

	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/users/view", viewUserHandler)
	http.HandleFunc("/users/edit", editUserHandler)
	//http.HandleFunc("/accounts/edit", editAccountHandler)
	// http.HandleFunc("/customers", customersHandler)
	http.HandleFunc("/deposits", makeDepositHandler)
	http.HandleFunc("/withdrawals", makeWithdrawalHandler)
	http.HandleFunc("/transfers", makeTransferHandler)
	http.HandleFunc("/loans", loansHandler)

	http.HandleFunc("/login", loginHandler)

	http.HandleFunc("/activate", activateAccountHandler)
	log.Println("Frontend running at http://localhost:9000/")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}

// Helper to execute API requests with the Bearer token attached
func callAPI(r *http.Request, method, url string, body io.Reader) (*http.Response, error) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+cookie.Value)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return http.DefaultClient.Do(req)
}
