package validation

import (
	"errors"
	"strings"
)

type PasswordValidator struct{}

func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{}
}

func (v *PasswordValidator) Validate(password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("password is required")
	}

	return nil
}

func (v *PasswordValidator) ValidateConfirmation(password, confirmation string) error {
	if err := v.Validate(password); err != nil {
		return err
	}

	if strings.TrimSpace(confirmation) == "" {
		return errors.New("password confirmation is required")
	}

	if password != confirmation {
		return errors.New("passwords do not match")
	}

	return nil
}
