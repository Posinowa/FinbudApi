package validator

import (
	"errors"
	"unicode"
)

// ValidatePasswordStrength şifrenin güvenlik kurallarına uygun olup olmadığını kontrol eder.
// Kurallar: min 8 karakter, en az 1 büyük harf, 1 küçük harf, 1 rakam.
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("şifre en az 8 karakter olmalıdır")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}

	if !hasUpper {
		return errors.New("şifre en az 1 büyük harf içermelidir")
	}
	if !hasLower {
		return errors.New("şifre en az 1 küçük harf içermelidir")
	}
	if !hasDigit {
		return errors.New("şifre en az 1 rakam içermelidir")
	}

	return nil
}
