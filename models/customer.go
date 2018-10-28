package models

import "time"

// Gender is a helper enum that represents customer gender
type Gender string

const (
	// NoGender is a default value for gender
	NoGender Gender = ""
	// Male is a male gender
	Male Gender = "Male"
	// Female is a female gender
	Female Gender = "Female"
)

// IsValidGender checks if the given gender string represents allowed gender
func IsValidGender(gender string) bool {
	return gender == string(Female) || gender == string(Male) || gender == string(NoGender)
}

// Customer represents a basic info for a customer
type Customer struct {
	ID        int
	Revision  int
	FirstName string
	LastName  string
	BirthDate time.Time
	Gender    Gender
	Email     string
	Address   string
}
