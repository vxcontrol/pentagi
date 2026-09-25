package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"pentagi/pkg/password"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
)

const (
	solidRegexString    = "^[a-z0-9_\\-]+$"
	clDateRegexString   = "^[0-9]{2}[.-][0-9]{2}[.-][0-9]{4}$"
	semverRegexString   = "^[0-9]+\\.[0-9]+(\\.[0-9]+)?$"
	semverexRegexString = "^(v)?[0-9]+\\.[0-9]+(\\.[0-9]+)?(\\.[0-9]+)?(-[a-zA-Z0-9]+)?$"
)

var (
	validate *validator.Validate
)

func GetValidator() *validator.Validate {
	return validate
}

// IValid is interface to control all models from user code
type IValid interface {
	Valid() error
}

func templateValidatorString(regexpString string) validator.Func {
	regexpValue := regexp.MustCompile(regexpString)
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()
		matchString := func(str string) bool {
			if str == "" && fl.Param() == "omitempty" {
				return true
			}
			return regexpValue.MatchString(str)
		}

		switch field.Kind() {
		case reflect.String:
			return matchString(fl.Field().String())
		case reflect.Slice, reflect.Array:
			for i := 0; i < field.Len(); i++ {
				if !matchString(field.Index(i).String()) {
					return false
				}
			}
			return true
		case reflect.Map:
			for _, k := range field.MapKeys() {
				if !matchString(field.MapIndex(k).String()) {
					return false
				}
			}
			return true
		default:
			return false
		}
	}
}

func strongPasswordValidatorString() validator.Func {
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()

		switch field.Kind() {
		case reflect.String:
			return password.IsStrong(field.String())
		default:
			return false
		}
	}
}

// MaxPasswordBytes is the longest password bcrypt.GenerateFromPassword accepts.
const MaxPasswordBytes = password.MaxBytes

func passwordLengthValidatorString() validator.Func {
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()

		switch field.Kind() {
		case reflect.String:
			return password.FitsHashLimit(field.String())
		default:
			return false
		}
	}
}

var emailFormatRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isRealEmail(email string) bool {
	return len(email) > 4 && emailFormatRegex.MatchString(email)
}

func emailValidatorString() validator.Func {
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()

		switch field.Kind() {
		case reflect.String:
			email := fl.Field().String()
			if email == "admin" {
				return true
			}
			if err := validate.Var(email, "required,uuid"); err == nil {
				return true
			}
			return isRealEmail(email)
		default:
			return false
		}
	}
}

func strictEmailValidatorString() validator.Func {
	return func(fl validator.FieldLevel) bool {
		return fl.Field().Kind() == reflect.String && isRealEmail(fl.Field().String())
	}
}

func oauthMinScope() validator.Func {
	scopeParts := []string{
		"openid",
		"email",
	}
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()

		switch field.Kind() {
		case reflect.String:
			scope := strings.ToLower(fl.Field().String())
			for _, part := range scopeParts {
				if !strings.Contains(scope, part) {
					return false
				}
			}
			return true
		default:
			return false
		}
	}
}

func deepValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		if iv, ok := fl.Field().Interface().(IValid); ok {
			if err := iv.Valid(); err != nil {
				return false
			}
		}

		return true
	}
}

func scanFromJSON(input interface{}, output interface{}) error {
	if v, ok := input.(string); ok {
		return json.Unmarshal([]byte(v), output)
	} else if v, ok := input.([]byte); ok {
		if err := json.Unmarshal(v, output); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("unsupported type of input value to scan")
}

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("solid", templateValidatorString(solidRegexString))
	_ = validate.RegisterValidation("cldate", templateValidatorString(clDateRegexString))
	_ = validate.RegisterValidation("semver", templateValidatorString(semverRegexString))
	_ = validate.RegisterValidation("semverex", templateValidatorString(semverexRegexString))
	_ = validate.RegisterValidation("stpass", strongPasswordValidatorString())
	_ = validate.RegisterValidation("passlen", passwordLengthValidatorString())
	_ = validate.RegisterValidation("vmail", emailValidatorString())
	_ = validate.RegisterValidation("realemail", strictEmailValidatorString())
	_ = validate.RegisterValidation("oauth_min_scope", oauthMinScope())
	_ = validate.RegisterValidation("valid", deepValidator())
	_ = validate.RegisterValidation("notblank", validators.NotBlank)

	// Check validation interface for all models
	_, _ = reflect.ValueOf(Login{}).Interface().(IValid)
	_, _ = reflect.ValueOf(AuthCallback{}).Interface().(IValid)

	_, _ = reflect.ValueOf(User{}).Interface().(IValid)
	_, _ = reflect.ValueOf(Password{}).Interface().(IValid)
	_, _ = reflect.ValueOf(EmailChange{}).Interface().(IValid)

	_, _ = reflect.ValueOf(Role{}).Interface().(IValid)
	_, _ = reflect.ValueOf(Prompt{}).Interface().(IValid)
	_, _ = reflect.ValueOf(Assistant{}).Interface().(IValid)
	_, _ = reflect.ValueOf(Flow{}).Interface().(IValid)
	_, _ = reflect.ValueOf(Provider{}).Interface().(IValid)
}
