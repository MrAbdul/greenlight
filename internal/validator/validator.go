package validator

import (
	"regexp"
	"slices"
)

// regex matching for email checks
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	Errors map[string]any
}

func New() *Validator {
	return &Validator{Errors: make(map[string]any)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}
func (v *Validator) AddError(key string, msg string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = msg
	}
}
func (v *Validator) Check(ok bool, key, mesage string) {
	if !ok {
		v.AddError(key, mesage)
	}
}
func PermittedValues[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
func Unique[T comparable](value []T) bool {
	uniqValues := make(map[T]bool)
	for _, v := range value {
		uniqValues[v] = true
	}
	return len(value) == len(uniqValues)
}
