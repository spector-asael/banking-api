
package main

import (
	"net/http"
	"log"
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

	http.HandleFunc("/accounts", accountsHandler)
	http.HandleFunc("/accounts/create", createAccountHandler)
	http.HandleFunc("/accounts/view", viewAccountHandler)
	//http.HandleFunc("/accounts/edit", editAccountHandler)
	// http.HandleFunc("/customers", customersHandler)
	http.HandleFunc("/deposits", makeDepositHandler)
	// http.HandleFunc("/withdrawals", withdrawalsHandler)
	// http.HandleFunc("/loans", loansHandler)
	log.Println("Frontend running at http://localhost:9000/")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}

	