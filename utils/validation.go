package utils

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
func ValidateDocumentID(documentID string) error {
	if documentID == "" {
		return &ValidationError{
			Field:   "document_id",
			Message: "Document ID is required",
		}
	}
	cleanDocumentID := strings.ReplaceAll(documentID, " ", "")
	cleanDocumentID = strings.ReplaceAll(cleanDocumentID, "-", "")

	if len(cleanDocumentID) < 10 {
		return &ValidationError{
			Field:   "document_id",
			Message: "Document ID must be at least 10 characters long",
		}
	}

	return nil
}
func ValidateFileNumber(fileNo string) error {
	if fileNo == "" {
		return &ValidationError{
			Field:   "file_no",
			Message: "File number is required",
		}
	}
	if len(fileNo) < 3 {
		return &ValidationError{
			Field:   "file_no",
			Message: "File number must be at least 3 characters long",
		}
	}

	return nil
}
func ValidateEmail(email string) error {
	if email == "" {
		return nil
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return &ValidationError{
			Field:   "email",
			Message: "Invalid email format",
		}
	}

	return nil
}
func ValidateContact(contact string) error {
	if contact == "" {
		return &ValidationError{
			Field:   "contact",
			Message: "Contact number is required",
		}
	}
	cleanContact := strings.ReplaceAll(contact, " ", "")
	cleanContact = strings.ReplaceAll(cleanContact, "-", "")
	cleanContact = strings.ReplaceAll(cleanContact, "(", "")
	cleanContact = strings.ReplaceAll(cleanContact, ")", "")
	cleanContact = strings.ReplaceAll(cleanContact, "+", "")

	if len(cleanContact) < 10 {
		return &ValidationError{
			Field:   "contact",
			Message: "Contact number must be at least 10 digits",
		}
	}

	return nil
}
func ValidateGender(gender string) error {
	if gender == "" {
		return nil
	}

	validGenders := map[string]bool{
		"male":    true,
		"female":  true,
		"other":   true,
		"unknown": true,
	}

	if !validGenders[strings.ToLower(gender)] {
		return &ValidationError{
			Field:   "gender",
			Message: "Gender must be one of: male, female, other, unknown",
		}
	}

	return nil
}

func ValidateMaritalStatus(maritalStatus string) error {
	if maritalStatus == "" {
		return nil
	}

	validStatuses := map[string]bool{
		"D": true, // Divorced
		"L": true, // Legally Separated
		"M": true, // Married
		"U": true, // Unmarried
		"W": true, // Widowed
	}

	// Convert to uppercase for case-insensitive comparison
	upperStatus := strings.ToUpper(maritalStatus)

	if !validStatuses[upperStatus] {
		return &ValidationError{
			Field:   "martial_status",
			Message: "Marital status must be one of: D (Divorced), L (Legally Separated), M (Married), U (Unmarried), W (Widowed)",
		}
	}

	return nil
}
func ValidatePatientData(fullName, documentID, fileNo, email, contact, gender, maritalStatus string) []error {
	var errors []error
	if fullName == "" {
		errors = append(errors, &ValidationError{
			Field:   "full_name",
			Message: "Full name is required",
		})
	} else if len(fullName) < 3 {
		errors = append(errors, &ValidationError{
			Field:   "full_name",
			Message: "Full name must be at least 3 characters long",
		})
	}
	if err := ValidateDocumentID(documentID); err != nil {
		errors = append(errors, err)
	}
	if err := ValidateFileNumber(fileNo); err != nil {
		errors = append(errors, err)
	}
	if err := ValidateEmail(email); err != nil {
		errors = append(errors, err)
	}
	if err := ValidateContact(contact); err != nil {
		errors = append(errors, err)
	}
	if err := ValidateGender(gender); err != nil {
		errors = append(errors, err)
	}
	if err := ValidateMaritalStatus(maritalStatus); err != nil {
		errors = append(errors, err)
	}

	return errors
}
func FormatValidationErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	var messages []string
	for _, err := range errors {
		messages = append(messages, err.Error())
	}

	return errors[0]
}

func CombineValidationErrors(errors []error) string {
	if len(errors) == 0 {
		return ""
	}
	var messages []string
	for _, err := range errors {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}
