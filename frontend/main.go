package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/persons", personsHandler)
	http.HandleFunc("/persons/create", createPersonHandler)
	http.HandleFunc("/persons/view", viewPersonHandler)
	http.HandleFunc("/persons/edit", editPersonHandler)

	http.HandleFunc("/customers", customersHandler)
	http.HandleFunc("/customers/create", createCustomerHandler)
	http.HandleFunc("/customers/view", viewCustomerHandler)
	http.HandleFunc("/customers/kyc", updateKYCHandler)

	http.HandleFunc("/employees", employeesHandler)
	http.HandleFunc("/employees/create", createEmployeeHandler)
	http.HandleFunc("/employees/view", viewEmployeeHandler)
	http.HandleFunc("/employees/edit", editEmployeeHandler)

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

	http.HandleFunc("/activate", activateAccountHandler)
	log.Println("Frontend running at http://localhost:9000/")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}
