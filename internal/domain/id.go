package domain

import (
	"errors"
	"fmt"
	"regexp"
)

var idRegex = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// ErrInvalidID indicates that a string does not conform to the identifier format.
var ErrInvalidID = errors.New("invalid identifier: must match ^[a-z][a-z0-9_]*$")

// IsValidID reports whether the string is a valid identifier.
func IsValidID(id string) bool {
	return idRegex.MatchString(id)
}

// ValidateID returns an error if id does not match ^[a-z][a-z0-9_]*$.
func ValidateID(field, id string) error {
	if !IsValidID(id) {
		return fmt.Errorf("%s %q is invalid: must match ^[a-z][a-z0-9_]*$", field, id)
	}
	return nil
}
