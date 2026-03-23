// Filename: cmd/api/persons.go
package handlers

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/spector-asael/banking-api/cmd/api/dependencies/helpers"
	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/validator"
)

func (a *HandlerDependencies) getAllPersonsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	// Input struct
	var input struct {
		Name string
		filters data.Filters
	}

	// Read query params
	qs := r.URL.Query()

	input.Name = a.Helper.ReadString(qs, "name", "")

	input.filters.Page = a.Helper.ReadInt(qs, "page", 1)
	input.filters.PageSize = a.Helper.ReadInt(qs, "page_size", 5)
	input.filters.Sort = a.Helper.ReadString(qs, "sort", "id")

	// Safelist (VERY IMPORTANT for SQL injection protection)
	input.filters.SortSafelist = []string{
		"id", "-id",
		"first_name", "-first_name",
		"last_name", "-last_name",
		"created_at", "-created_at",
	}

	// Validate input
	v := validator.New()

	data.ValidateFilters(v, input.filters)

	if !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// Call model
	persons, metadata, err := a.Models.Persons.GetAll(input.Name, input.filters)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	// Response
	err = a.Helper.WriteJSON(
		w,
		http.StatusOK,
		helpers.Envelope{
			"persons":   persons,
			"@metadata": metadata,
		},
		nil,
	)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

func (a *HandlerDependencies) getPersonBySSIDHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	ssid := params.ByName("ssid")

	if ssid == "" {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"social_security_number": "must be provided"})
		return
	}

	person, err := a.Models.Persons.GetBySSID(ssid)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			a.Helper.WriteJSON(w, http.StatusNotFound, helpers.Envelope{"error": "person not found"}, nil)
		default:
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"person": person}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}


func (a *HandlerDependencies) createPersonHandler(w http.ResponseWriter, r *http.Request) {
	var input data.Person

	err := a.Helper.ReadJSON(w, r, &input)
	if err != nil {
		a.Helper.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidatePerson(v, &input)
	if !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.Models.Persons.Insert(&input)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	err = a.Helper.WriteJSON(w, http.StatusCreated, helpers.Envelope{"person": input}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

func (a *HandlerDependencies) updatePersonHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	ssid := params.ByName("ssid")

	if ssid == "" {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"social_security_number": "must be provided"})
		return
	}

	person, err := a.Models.Persons.GetBySSID(ssid)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			a.Helper.WriteJSON(w, http.StatusNotFound, helpers.Envelope{"error": "person not found"}, nil)
		default:
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}

	var input data.Person
	err = a.Helper.ReadJSON(w, r, &input)
	if err != nil {
		a.Helper.BadRequestResponse(w, r, err)
		return
	}

	// Update fields (only overwrite non-empty values)
	if input.FirstName != "" {
		person.FirstName = input.FirstName
	}
	if input.LastName != "" {
		person.LastName = input.LastName
	}
	if input.Email != "" {
		person.Email = input.Email
	}
	if input.PhoneNumber != "" {
		person.PhoneNumber = input.PhoneNumber
	}
	if input.LivingAddress != "" {
		person.LivingAddress = input.LivingAddress
	}

	v := validator.New()
	data.ValidatePerson(v, person)
	if !v.IsEmpty() {
		a.Helper.FailedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.Models.Persons.Update(person)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
		return
	}

	err = a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"person": person}, nil)
	if err != nil {
		a.Helper.ServerErrorResponse(w, r, err)
	}
}

func (a *HandlerDependencies) deletePersonHandler(w http.ResponseWriter, r *http.Request) {
	// Get the SSID from the URL path
	params := httprouter.ParamsFromContext(r.Context())
	ssid := params.ByName("ssid")

	if ssid == "" {
		a.Helper.FailedValidationResponse(w, r, map[string]string{"social_security_number": "must be provided"})
		return
	}

	// Call the model to delete the person
	err := a.Models.Persons.DeleteBySSN(ssid)
	if err != nil {
		switch {
		case err == data.ErrRecordNotFound:
			// Person not found
			a.Helper.WriteJSON(w, http.StatusNotFound, helpers.Envelope{"error": "person not found"}, nil)
		default:
			// Other database errors
			a.Helper.ServerErrorResponse(w, r, err)
		}
		return
	}

	// Successful deletion
	a.Helper.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "person deleted successfully"}, nil)
}