package validation

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

const EMAIL_ADDRESS_MAX_SIZE = 256

// Validate checks if the provided email address is valid
func ValidateEmail(email string) bool {
	if len(email) > EMAIL_ADDRESS_MAX_SIZE {
		return false
	}

	email = strings.TrimSpace(email)

	if len(email) == 0 {
		return false
	}

	// Check basic pattern
	if !emailRegex.MatchString(email) {
		return false
	}

	// Additional checks
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local, domain := parts[0], parts[1]

	// Local part checks
	if len(local) > 64 {
		return false
	}

	// Domain checks
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	return true
}
