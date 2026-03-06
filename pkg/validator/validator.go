package validator

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Validator struct {
	v *validator.Validate
}

type ValidationErrors map[string]string

func New() *Validator {
	v := validator.New()

	// Use JSON field names in validation error messages.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{v: v}
}

func (cv *Validator) Validate(i interface{}) error {
	if err := cv.v.Struct(i); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			fields := make(ValidationErrors, len(ve))
			for _, fe := range ve {
				fields[fe.Field()] = fieldMessage(fe)
			}
			return echo.NewHTTPError(http.StatusUnprocessableEntity, fields)
		}
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	return nil
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", strings.ReplaceAll(fe.Param(), " ", ", "))
	case "url":
		return "must be a valid URL"
	case "uuid4":
		return "must be a valid UUID"
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}
