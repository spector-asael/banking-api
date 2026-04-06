// Filename: cmd/api/employees.go
package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/spector-asael/banking-api/cmd/api/dependencies/helpers"
	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/validator"
)

// GET /employees
func (a *HandlerDependencies) getAllEmployeesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		filters data.Filters
	}

	qs := r.URL.Query()

	input.filters.Page = a.Helper.ReadInt(qs, "page", 1)
	input.filters.PageSize = a.Helper.ReadInt(qs, "page_size", 10)
	input.filters.Sort = a.Helper.ReadString(qs, "sort", "id")

	input.filters.SortSafelist = []string{
		"id", "-id",
		"first_name", "-first_name",
		"last_name", "-last_name",
		"status", "-status",
	}

	v := validator.New()
	if data.ValidateFilters(v, input.filters); !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}

	employees, metadata, err := a.Models.Employees.GetAll(input.filters)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"employees": employees, "@metadata": metadata}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

// GET /employees/:id
func (a *HandlerDependencies) getEmployeeByIDHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id <= 0 {
		a.Helper.NotFoundResponse(w, r)
		return
	}

	employee, err := a.Models.Employees.GetByID(id)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			a.Helper.NotFoundResponse(w, r)
		default:
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"employee": employee}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

// POST /employees
func (a *HandlerDependencies) createEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SSID        string `json:"ssid"`
		Username    string `json:"username"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		BranchID    int64  `json:"branch_id"`
		PositionID  int64  `json:"position_id"`
		AccountType int    `json:"account_type"`
	}

	err := a.Helper.ReadJSON(w, r, &input)
	if err != nil {
		log.Printf("[DEBUG] Failed to read JSON: %v\n", err)
		a.Helper.BadRequestResponse(w, r, err)
		return
	}

	// Validate input
	v := validator.New()
	v.Check(input.SSID != "", "ssid", "must be provided")
	v.Check(input.BranchID > 0, "branch_id", "must be a valid branch")
	v.Check(input.PositionID > 0, "position_id", "must be a valid position")
	v.Check(input.AccountType >= 1 && input.AccountType <= 4, "account_type", "must be a valid employee account type")
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.IsEmpty() {
		log.Printf("[DEBUG] Validation failed: %v\n", v.Errors)
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// 1. Verify Person exists
	person, err := a.Models.Persons.GetBySSID(input.SSID)
	if err != nil {
		log.Printf("[DEBUG] Failed to find person with SSID %s: %v\n", input.SSID, err)
		a.Helper.FailedValidationResponse(w, r, map[string]string{"ssid": "no person found with this SSID"})
		return
	}

	// 2. Prepare Employee object
	employee := &data.Employee{
		PersonID:   person.ID,
		BranchID:   input.BranchID,
		PositionID: input.PositionID,
		HireDate:   time.Now(),
		Status:     "active",
	}

	// 3. Start Transaction
	tx, err := a.Models.Employees.DB.Begin()
	if err != nil {
		log.Printf("[DEBUG] Failed to begin DB transaction: %v\n", err)
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}
	defer tx.Rollback() // Safe to defer, will do nothing if tx.Commit() succeeds

	// 4. Insert Employee
	err = a.Models.Employees.InsertTx(tx, employee)
	if err != nil {
		log.Printf("[DEBUG] DB Error inserting Employee: %v\n", err)
		if err == data.ErrDuplicateEmployee {
			a.Helper.FailedValidationResponse(w, r, map[string]string{"ssid": "employee record already exists for this person"})
			return
		}
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	// 5. Insert Associated User
	employeeID := int(employee.ID)
	user := &data.User{
		Username:     input.Username,
		Email:        input.Email,
		Activated:    true,
		Employee_id:  &employeeID,
		Customer_id:  nil,
		Account_type: input.AccountType,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		log.Printf("[DEBUG] Error hashing password: %v\n", err)
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	err = a.Models.Users.InsertTx(tx, user)
	if err != nil {
		log.Printf("[DEBUG] DB Error inserting User: %v\n", err)
		switch {
		case err == data.ErrDuplicateEmail:
			a.Helper.FailedValidationResponse(w, r, map[string]string{"email": "email already in use"})
		default:
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}

	// 6. Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("[DEBUG] Failed to commit transaction: %v\n", err)
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	// --- Email Logic ---
	roleTitles := map[int]string{
		1: "Teller",
		2: "Customer Service Representative",
		3: "Manager",
		4: "Administrator",
	}
	roleName := roleTitles[input.AccountType]

	templateData := map[string]any{
		"Account_type":  roleName,
		"SSID":          input.SSID,
		"Username":      user.Username,
		"Email":         user.Email,
		"CreatedAt":     time.Now().Format("January 02, 2006 at 15:04 MST"),
		"PasswordSetAt": time.Now().Format("January 02, 2006 at 15:04 MST"),
	}

	a.Helper.Background(func() {
		sendErr := a.Mailer.Send(user.Email, "employee_welcome.tmpl", templateData)
		if sendErr != nil {
			a.Logger.Error(sendErr.Error())
		}
	})

	log.Printf("[DEBUG] SUCCESS! Employee %d and User created.\n", employee.ID)
	a.Helper.WriteJSON(w, http.StatusCreated, helpers.Envelope{"employee": employee, "user": user}, nil)
}

// PATCH /employees/:id
func (a *HandlerDependencies) updateEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, _ := strconv.ParseInt(params.ByName("id"), 10, 64)

	employee, err := a.Models.Employees.GetByID(id)
	if err != nil {
		a.Helper.NotFoundResponse(w, r)
		return
	}

	var input struct {
		BranchID   *int64  `json:"branch_id"`
		PositionID *int64  `json:"position_id"`
		Status     *string `json:"status"`
	}

	if err := a.Helper.ReadJSON(w, r, &input); err != nil {
		a.Helper.BadRequestResponse(w, r, err)
		return
	}

	if input.BranchID != nil {
		employee.BranchID = *input.BranchID
	}
	if input.PositionID != nil {
		employee.PositionID = *input.PositionID
	}
	if input.Status != nil {
		employee.Status = *input.Status
	}

	v := validator.New()
	if data.ValidateEmployee(v, employee); !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}

	if err := a.Models.Employees.Update(employee); err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"employee": employee}, nil)
}

// DELETE /employees/:id
// DELETE /employees/:id
func (a *HandlerDependencies) deleteEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id <= 0 {
		a.Helper.NotFoundResponse(w, r)
		return
	}

	// Create a context with a 3-second timeout
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// 1. Start a transaction
	tx, err := a.Models.Employees.DB.BeginTx(ctx, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}
	defer tx.Rollback()

	// 2. Delete the associated User account first (to avoid foreign key constraint errors)
	_, err = tx.ExecContext(ctx, "DELETE FROM users WHERE employee_id = $1", id)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	// 3. Delete the Employee
	result, err := tx.ExecContext(ctx, "DELETE FROM employees WHERE id = $1", id)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	if rowsAffected == 0 {
		a.Helper.NotFoundResponse(w, r)
		return
	}

	// 4. Commit the transaction
	if err = tx.Commit(); err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "employee and associated user account deleted successfully"}, nil)
}
