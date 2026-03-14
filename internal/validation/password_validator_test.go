package validation_test

import (
	"testing"

	"github.com/sp/pwpdf/internal/validation"
)

func TestPasswordValidatorRejectsEmptyPassword(t *testing.T) {
	validator := validation.NewPasswordValidator()

	if err := validator.Validate("   "); err == nil {
		t.Fatal("expected an error for an empty password")
	}
}

func TestPasswordValidatorRequiresMatchingConfirmation(t *testing.T) {
	validator := validation.NewPasswordValidator()

	if err := validator.ValidateConfirmation("secret", "different"); err == nil {
		t.Fatal("expected an error for mismatched passwords")
	}
}
