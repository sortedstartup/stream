package validation

import (
	"strings"
	"testing"
)

// TestEmailValidation contains test cases for the email validator
func TestEmailValidation(t *testing.T) {

	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"Valid email", "user@example.com", true},
		{"Valid email with dots", "user.name@example.com", true},
		{"Valid email with plus", "user+tag@example.com", true},
		{"Valid email with numbers", "user123@example.com", true},
		{"Empty string", "", false},
		{"Missing @", "userexample.com", false},
		{"Multiple @", "user@domain@example.com", false},
		{"Invalid domain", "user@.com", false},
		{"Invalid local part", "@example.com", false},
		{"Too long local part", strings.Repeat("a", 65) + "@example.com", false},
		{"Too long email", strings.Repeat("a", 255) + "@example.com", false},
		{"Invalid characters", "user!@example.com", false},
		{"Domain starts with dot", "user@.example.com", false},
		{"Domain ends with dot", "user@example.com.", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateEmail(tt.email)
			if got != tt.want {
				t.Errorf("Validate(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}
