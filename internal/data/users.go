// Filename: internal/data/users.go
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/spector-asael/banking-api/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

// Specificy a custom duplicate email error message
var ErrDuplicateEmail = errors.New("duplicate email")
var ErrEditConflict = errors.New("edit conflict")

type UserModel struct {
	DB *sql.DB
}

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"name"`
	Email        string    `json:"email"`
	Password     password  `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	Activated    bool      `json:"activated"`
	Version      int       `json:"version"`
	Employee_id  *int      `json:"employee_id,omitempty"`
	Customer_id  *int      `json:"customer_id,omitempty"`
	Account_type int       `json:"account_type"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	// Hash the password using bcrypt with a cost of 12.
	// The cost determines the computational complexity of the hashing process, making it more resistant to brute-force attacks.
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	// Store the plaintext password in memory for validation purposes.
	p.hash = hash
	// Store the hashed password in the struct for later comparison during authentication.
	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	// Compare the provided plaintext password with the stored hash using bcrypt's comparison function.
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil // Password does not match
		default:
			return false, err // An error occurred during comparison
		}
	}
	return true, nil // Password matches
}

// Validate the email address for a user.
// It must be provided and match a valid email format.
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

// Check that a valid password is provided
// It must fit the following criteria:
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Username != "", "username", "must be provided")
	v.Check(len(user.Username) <= 200, "username", "must not be more than 200 bytes long")

	// Validate email for user
	ValidateEmail(v, user.Email)
	// validate the plain text password
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	// check if we messed up in our codebase
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func (u UserModel) InsertTx(tx *sql.Tx, user *User) error {
	query := `
        INSERT INTO users (username, email, password_hash, activated, employee_id, customer_id, account_type)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, version
    `

	err := tx.QueryRow(
		query,
		user.Username,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.Employee_id,
		user.Customer_id,
		user.Account_type,
	).Scan(&user.ID, &user.CreatedAt, &user.Version)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" && (pqErr.Constraint == "users_email_key" || pqErr.Constraint == "users_username_key") {
				return ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}

// Insert a new user into the database.
func (u UserModel) Insert(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := u.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = u.InsertTx(tx, user)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetByID returns a user from the database by their ID.
func (u UserModel) GetByID(id int64) (*User, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT id, created_at, username, email, password_hash, version, activated, employee_id, customer_id, account_type
        FROM users
        WHERE id = $1`

	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.Version,
		&user.Activated,
		&user.Employee_id,
		&user.Customer_id,
		&user.Account_type,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Get a user from the database by their email address.
func (u UserModel) GetByEmail(email string) (*User, error) {
	query := `
        SELECT id, created_at, username, email, password_hash, version, activated, employee_id, customer_id, account_type
        FROM users
        WHERE email = $1`

	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.Version,
		&user.Activated,
		&user.Employee_id,
		&user.Customer_id,
		&user.Account_type,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// GetAll returns a list of users with filtering, sorting, and pagination support.
func (u UserModel) GetAll(username string, email string, f Filters) ([]*User, Metadata, error) {
	// 1. Construct the SQL query with filtering and COUNT(*) OVER() for metadata.
	// We use LOWER() for case-insensitive search.
	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(), id, created_at, username, email, password_hash, version, activated, employee_id, customer_id, account_type
        FROM users
        WHERE (STRPOS(LOWER(username), LOWER($1)) > 0 OR $1 = '')
        AND (STRPOS(LOWER(email), LOWER($2)) > 0 OR $2 = '')
        ORDER BY %s %s, id ASC
        LIMIT $3 OFFSET $4`, f.sortColumn(), f.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 2. Execute the query using limit() and offset() helpers.
	rows, err := u.DB.QueryContext(ctx, query, username, email, f.limit(), f.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	users := []*User{}

	for rows.Next() {
		var user User
		err := rows.Scan(
			&totalRecords, // Scan the COUNT(*) OVER() value
			&user.ID,
			&user.CreatedAt,
			&user.Username,
			&user.Email,
			&user.Password.hash,
			&user.Version,
			&user.Activated,
			&user.Employee_id,
			&user.Customer_id,
			&user.Account_type,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// 3. Calculate Metadata based on the total records found.
	metadata := calculateMetaData(totalRecords, f.Page, f.PageSize)

	return users, metadata, nil
}

// Update a User. If the version number is different
// Than what it was before we ran the query, it means
// someone did a previous edit or is doing an edit, so
// our query will fail and we would need to try again a bit later
// This is an optimistic locking strategy to prevent lost updates
// when multiple clients are trying to update the same record at the same time.
func (u UserModel) Update(user *User) error {
	query := `
            UPDATE users
            SET username = $1, email = $2, password_hash = $3, activated = $4, employee_id = $5, customer_id = $6, account_type = $7, version = version + 1
            WHERE id = $8 AND version = $9
            RETURNING version
        `
	args := []any{user.Username, user.Email, user.Password.hash, user.Activated, user.Employee_id, user.Customer_id, user.Account_type, user.ID, user.Version}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)

	// Check for errors during update
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
