package helpers

import (
	"regexp"
	"slices"
	"unicode"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func In(value string, list ...string) bool {
	if slices.Contains(list, value) == true {
		return true
	}
	return false
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func VerifyPassword(value string) bool {
	var number, upper, special bool
	numCharacters := 0
	for _, c := range value {
		switch {
		case unicode.IsNumber(c):
			number = true
			numCharacters++
		case unicode.IsUpper(c):
			upper = true
			numCharacters++
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
			numCharacters++
		case unicode.IsLetter(c) || c == ' ':
			continue
			numCharacters++
		default:
			return false
		}
	}
	if (number && upper && special) == true && (numCharacters >= 8) {
		return true
	}
	return false
}
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}
