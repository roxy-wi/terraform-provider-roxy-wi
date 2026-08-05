package roxywi

import (
	"errors"
	"fmt"
	"net/mail"
)

// Utility function to check if the error is a 404 not found error
func isNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 404
}

// Utility function to validate email format
func validateEmail(val interface{}, key string) (warns []string, errs []error) {
	_, err := mail.ParseAddress(val.(string))
	if err != nil {
		errs = append(errs, fmt.Errorf("%q must be a valid email address: %v", key, err))
	}
	return
}
