// Filename: internal/validator/validator.go
package validator
 
import (
  "slices"
)

// We will create a new type named Validator
type Validator struct {
    Errors map[string]string
} 

// Construct a new Validator and return a pointer to it
// All validation errors go into this one Validator instance
func New() *Validator {
    return &Validator {
        Errors: make(map[string]string),
    }
}