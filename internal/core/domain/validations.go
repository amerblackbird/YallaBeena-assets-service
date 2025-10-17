package domain

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a validation error
type ValidationError struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Params  []interface{} `json:"params,omitempty"`
	Value   interface{}   `json:"value,omitempty"`
}

type ValidationErrors map[string][]ValidationError

func NewValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("phone", validatePhoneNumber)
	return validate
}

func validatePhoneNumber(fl validator.FieldLevel) bool {
	// Saudi Arabia phone number validation regex
	pattern := `^\+?966[5-9][0-9]{8}$`
	phone := fl.Field().String()
	if phone == "" {
		return false
	}
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

func GetValidationErrors(errs validator.ValidationErrors) ValidationErrors {
	validationErrors := make(ValidationErrors)

	for _, e := range errs {
		fieldName := strings.ToLower(e.Field())
		validationError := GetValidationError(e)

		if _, exists := validationErrors[fieldName]; !exists {
			validationErrors[fieldName] = []ValidationError{}
		}
		validationErrors[fieldName] = append(validationErrors[fieldName], validationError)
	}

	return validationErrors
}

func GetValidationError(e validator.FieldError) ValidationError {
	tag := e.Tag()
	param := e.Param()
	value := e.Value()

	switch tag {
	case "required":
		return ValidationError{
			Code:    "required",
			Message: "This field is required",
		}
	case "min":
		if e.Kind().String() == "string" {
			return ValidationError{
				Code:    "min",
				Message: fmt.Sprintf("Minimum length is %s characters", param),
				Params:  []interface{}{param},
				Value:   value,
			}
		}
		return ValidationError{
			Code:    "min",
			Message: fmt.Sprintf("Minimum value is %s", param),
			Params:  []interface{}{param},
		}
	case "max":
		if e.Kind().String() == "string" {
			return ValidationError{
				Code:    "max",
				Message: fmt.Sprintf("Maximum length is %s characters", param),
				Params:  []interface{}{param},
				Value:   value,
			}
		}
		return ValidationError{
			Code:    "max",
			Message: fmt.Sprintf("Maximum value is %s", param),
			Params:  []interface{}{param},
			Value:   value,
		}
	case "email":
		return ValidationError{
			Code:    "email",
			Message: "Please enter a valid email address",
			Value:   value,
		}
	case "numeric":
		return ValidationError{
			Code:    "numeric",
			Message: "This field must contain only numbers",
			Value:   value,
		}
	case "alpha":
		return ValidationError{
			Code:    "alpha",
			Message: "This field must contain only letters",
			Value:   value,
		}
	case "alphanum":
		return ValidationError{
			Code:    "alphanum",
			Message: "This field must contain only letters and numbers",
			Value:   value,
		}
	case "len":
		return ValidationError{
			Code:    "len",
			Message: fmt.Sprintf("This field must be exactly %s characters long", param),
			Params:  []interface{}{param},
			Value:   value,
		}
	case "oneof":
		values := strings.Split(param, " ")
		params := make([]interface{}, len(values))
		for i, v := range values {
			params[i] = v
		}
		return ValidationError{
			Code:    "oneof",
			Message: fmt.Sprintf("This field must be one of: %s", strings.Join(values, ", ")),
			Params:  params,
			Value:   value,
		}
	case "gte":
		return ValidationError{
			Code:    "gte",
			Message: fmt.Sprintf("This field must be greater than or equal to %s", param),
			Params:  []interface{}{param},
			Value:   value,
		}
	case "lte":
		return ValidationError{
			Code:    "lte",
			Message: fmt.Sprintf("This field must be less than or equal to %s", param),
			Params:  []interface{}{param},
			Value:   value,
		}
	case "gt":
		return ValidationError{
			Code:    "gt",
			Message: fmt.Sprintf("This field must be greater than %s", param),
			Params:  []interface{}{param},
			Value:   value,
		}
	case "lt":
		return ValidationError{
			Code:    "lt",
			Message: fmt.Sprintf("This field must be less than %s", param),
			Params:  []interface{}{param},
			Value:   value,
		}
	case "phone":
		return ValidationError{
			Code:    "phone",
			Message: "Please enter a valid phone number in the format +9665XXXXXXXX",
			Value:   value,
		}
	case "url":
		return ValidationError{
			Code:    "url",
			Message: "Please enter a valid URL",
			Params:  []interface{}{param},
			Value:   value,
		}
	case "md5":
		return ValidationError{
			Code:    "md5",
			Message: "This field must be a valid MD5 hash",
			Value:   value,
		}
	case "uuid":
		return ValidationError{
			Code:    "uuid",
			Message: "This field must be a valid UUID",
			Value:   value,
		}
	case "datetime":
		return ValidationError{
			Code:    "datetime",
			Message: "This field must be a valid date and time",
			Value:   value,
		}
	case "json":
		return ValidationError{
			Code:    "json",
			Message: "This field must be a valid JSON object",
			Value:   value,
		}
	case "ip":
		return ValidationError{
			Code:    "ip",
			Message: "This field must be a valid IP address",
			Value:   value,
		}
	default:
		return ValidationError{
			Code:    tag,
			Message: fmt.Sprintf("Validation failed for '%s' rule", tag),
			Value:   value,
		}
	}
}
