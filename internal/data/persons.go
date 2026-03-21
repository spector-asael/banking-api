package data 

import (
	"database/sql"
)

type PersonModel struct {
	DB *sql.DB
}

type Person struct {
	ID                     int    `json:"id"`
	FirstName              string `json:"first_name"`
	LastName               string `json:"last_name"`
	SocialSecurityNumber   string `json:"social_security_number"`
	Email                  string `json:"email"`
	DateOfBirth            string `json:"date_of_birth"`
	PhoneNumber            string `json:"phone_number"`
	LivingAddress          string `json:"living_address"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`

}

