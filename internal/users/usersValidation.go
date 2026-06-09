package users

import (
	"Librorum/internal/helpers"
)

func ValidateEmail(v *helpers.Validator, email string) {
	v.Check(email != "", "email", "email must be provided\n")
	v.Check(helpers.Matches(email, helpers.EmailRX), "email", "email must be a valid email address\n")
}

func ValidatePassword(v *helpers.Validator, password string) {
	v.Check(password != "", "password", "password must be provided\n")
	v.Check(helpers.VerifyPassword(password), "password",
		"password must contain atleast 8 characters, one of them been 1 special character, 1 upper case letter and 1 number to be valid\n")
}

func ValidateUser(v *helpers.Validator, user *UserProfile) {
	v.Check(user.Username != "", "username", "username must be provided\n")
	v.Check(user.DisplayName != "", "display_name", "display name must be provided\n")

	ValidateEmail(v, user.Email)

	if user.Password.Plaintext != nil {
		ValidatePassword(v, *user.Password.Plaintext)
	}
}
