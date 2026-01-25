package model

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
	"go.uber.org/zap"
)

var (
	schemaDecoder = schema.NewDecoder()
	validate      *validator.Validate
	usernameRegex = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]{2,39}$")
)

func init() {
	schemaDecoder.SetAliasTag("json")
	schemaDecoder.IgnoreUnknownKeys(true)

	validate = validator.New(validator.WithRequiredStructEnabled())
	err := validate.RegisterValidation("validusername", validateUsername)
	if err != nil {
		zap.S().Errorf("reg validation error: %s", err)
	}
}

func validateUsername(fl validator.FieldLevel) bool {
	return usernameRegex.MatchString(fl.Field().String())
}
