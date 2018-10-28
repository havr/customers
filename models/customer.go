package models

import (
	"time"
)

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
	ID        int    `json:"id"`
	Revision  int    `json:"revision"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	BirthDate Date   `json:"birthDate"`
	Gender    Gender `json:"gender"`
	Email     string `json:"email"`
	Address   string `json:"address"`
}

// Date is a time.Time with custom String() to display only date, without time
// This is relevant for template rendering only
type Date time.Time

func (d Date) String() string {
	return time.Time(d).Format("02 Jan 06")
}
