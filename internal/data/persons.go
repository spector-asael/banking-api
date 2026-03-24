package data 

import (
	"database/sql"
	"regexp"
	"github.com/spector-asael/banking-api/internal/validator"
	"time"
	"context"
	"fmt"
	"strings"
	"errors"
)

type PersonModel struct {
	DB *sql.DB
}

type Person struct {
	ID                     int64     `json:"id"`
	FirstName              string    `json:"first_name"`
	LastName               string    `json:"last_name"`
	SocialSecurityNumber   string    `json:"social_security_number"`
	Email                  string    `json:"email"`
	DateOfBirth            time.Time `json:"date_of_birth"`
	PhoneNumber            string    `json:"phone_number"`
	LivingAddress          string    `json:"living_address"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// GetByID returns a person by their database ID.
func (m PersonModel) GetByID(id int64) (*Person, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, first_name, last_name, social_security_number, email, date_of_birth, phone_number, living_address, created_at, updated_at
		FROM persons
		WHERE id = $1
	`

	var p Person
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.FirstName,
		&p.LastName,
		&p.SocialSecurityNumber,
		&p.Email,
		&p.DateOfBirth,
		&p.PhoneNumber,
		&p.LivingAddress,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &p, nil
}

// ValidatePerson checks that the person input is valid.
func ValidatePerson(v *validator.Validator, p *Person) {
	// Required fields
	v.Check(p.FirstName != "", "first_name", "must be provided")
	v.Check(p.LastName != "", "last_name", "must be provided")
	v.Check(p.SocialSecurityNumber != "", "social_security_number", "must be provided")
	v.Check(p.Email != "", "email", "must be provided")
	v.Check(!p.DateOfBirth.IsZero(), "date_of_birth", "must be provided")
	v.Check(p.PhoneNumber != "", "phone_number", "must be provided")
	v.Check(p.LivingAddress != "", "living_address", "must be provided")

	// Length constraints
	v.Check(len(p.FirstName) <= 100, "first_name", "must not exceed 100 characters")
	v.Check(len(p.LastName) <= 100, "last_name", "must not exceed 100 characters")
	v.Check(len(p.Email) <= 255, "email", "must not exceed 255 characters")

	// Email format (simple regex)
	emailRX := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	v.Check(emailRX.MatchString(p.Email), "email", "must be a valid email address")

	// Phone number (at least 7 digits)
	phoneRX := regexp.MustCompile(`^\d{7,}$`)
	v.Check(phoneRX.MatchString(p.PhoneNumber), "phone_number", "must be at least 7 digits")

	// SSN basic check (you can refine later)
	v.Check(len(p.SocialSecurityNumber) >= 5, "social_security_number", "must be a valid SSN")

	// DOB sanity check (not in future)
	v.Check(p.DateOfBirth.Before(time.Now()), "date_of_birth", "must be in the past")
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than 0")
	v.Check(f.PageSize > 0, "page_size", "must be greater than 0")
	v.Check(f.PageSize <= 100, "page_size", "must not be greater than 100")

	v.Check(
		validator.PermittedValue(f.Sort, f.SortSafelist...),
		"sort",
		"invalid sort value",
	)
}

func (m PersonModel) GetBySSID(ssn string) (*Person, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, first_name, last_name, social_security_number, email, date_of_birth, phone_number, living_address, created_at, updated_at
		FROM persons
		WHERE social_security_number = $1
	`

	var p Person

	err := m.DB.QueryRowContext(ctx, query, ssn).Scan(
		&p.ID,
		&p.FirstName,
		&p.LastName,
		&p.SocialSecurityNumber,
		&p.Email,
		&p.DateOfBirth,
		&p.PhoneNumber,
		&p.LivingAddress,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &p, nil
}

var ErrDuplicatePerson = errors.New("person with this SSID or email already exists")

func (m PersonModel) Insert(p *Person) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO persons (
			first_name,
			last_name,
			social_security_number,
			email,
			date_of_birth,
			phone_number,
			living_address
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := m.DB.QueryRowContext(
		ctx,
		query,
		p.FirstName,
		p.LastName,
		p.SocialSecurityNumber,
		p.Email,
		p.DateOfBirth,
		p.PhoneNumber,
		p.LivingAddress,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err.Error() != "" && (err.Error() == "UNIQUE constraint failed: persons.social_security_number" || err.Error() == "UNIQUE constraint failed: persons.email" || err.Error() == "pq: duplicate key value violates unique constraint \"persons_social_security_number_key\"" || err.Error() == "pq: duplicate key value violates unique constraint \"persons_email_key\"") {
			return ErrDuplicatePerson
		}
		return err
	}
	return nil
}

// Delete removes a person by social_security_number.
func (m PersonModel) DeleteBySSN(ssn string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		DELETE FROM persons
		WHERE social_security_number = $1
	`

	result, err := m.DB.ExecContext(ctx, query, ssn)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (m PersonModel) Update(p *Person) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE persons
		SET first_name = $1,
		    last_name = $2,
		    email = $3,
		    phone_number = $4,
		    living_address = $5,
		    date_of_birth = $6,
		    updated_at = NOW()
		WHERE social_security_number = $7
	`

	result, err := m.DB.ExecContext(ctx,
		query,
		p.FirstName,
		p.LastName,
		p.Email,
		p.PhoneNumber,
		p.LivingAddress,
		p.DateOfBirth,
		p.SocialSecurityNumber,
	)
	if err != nil {
		return err
	}

	return checkRowsAffected(result)
}

func checkRowsAffected(result sql.Result) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetAll returns a list of persons, with optional name filtering, sorting, and pagination
func (m PersonModel) GetAll(name string, f Filters) ([]*Person, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Base query with filtering on first_name or last_name
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(),
			id,
			first_name,
			last_name,
			social_security_number,
			email,
			date_of_birth,
			phone_number,
			living_address,
			created_at,
			updated_at
		FROM persons
		WHERE (LOWER(first_name) LIKE LOWER($1) OR LOWER(last_name) LIKE LOWER($1))
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3
	`,
		f.sortColumn(),     // resolves "sort" into a safe column
		f.sortDirection(),  // resolves "-" prefix into ASC/DESC
	)

	// Prepare the search string for the LIKE query
	search := "%" + strings.TrimSpace(name) + "%"

	// Execute query with pagination
	rows, err := m.DB.QueryContext(ctx, query, search, f.limit(), f.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	var persons []*Person

	// Iterate through rows
	for rows.Next() {
		var p Person

		// Scan each column into the Person struct
		err := rows.Scan(
			&totalRecords,
			&p.ID,
			&p.FirstName,
			&p.LastName,
			&p.SocialSecurityNumber,
			&p.Email,
			&p.DateOfBirth,
			&p.PhoneNumber,
			&p.LivingAddress,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		persons = append(persons, &p)
	}

	// Check for errors from iteration
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Calculate pagination metadata AFTER looping through rows
	metadata := calculateMetaData(totalRecords, f.Page, f.PageSize)

	return persons, metadata, nil
}