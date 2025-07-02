package helper

import "unicode"

func IsValidPassword(password string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	if len(password) >= 8 {
		hasMinLen = true
	}

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasNumber = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

func IsSameDigitPhone(phone string)bool{
	if len(phone) != 10 {
		return false
	}
	first := phone[0]
	for i := 1; i < 10; i++ {
		if phone[i] != first {
			return false
		}
	}
	return true
}

func IsName(name string)bool{

	var (
		hasNumber = false
		hasSpecial = false 
	)

	for _, ch := range name{
		switch  {
		case unicode.IsDigit(ch):
			hasNumber = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	return hasNumber || hasSpecial
}