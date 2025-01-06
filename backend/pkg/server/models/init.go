package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/xeipuuv/gojsonschema"
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
	numberRegex := regexp.MustCompile("[0-9]")
	alphaLRegex := regexp.MustCompile("[a-z]")
	alphaURegex := regexp.MustCompile("[A-Z]")
	specRegex := regexp.MustCompile("[!@#$&*]")
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()

		switch field.Kind() {
		case reflect.String:
			password := fl.Field().String()
			return len(password) > 15 || (len(password) >= 8 &&
				numberRegex.MatchString(password) &&
				alphaLRegex.MatchString(password) &&
				alphaURegex.MatchString(password) &&
				specRegex.MatchString(password))
		default:
			return false
		}
	}
}

func emailValidatorString() validator.Func {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
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
			return len(email) > 4 && emailRegex.MatchString(email)
		default:
			return false
		}
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

func getMapKeys(kvmap interface{}) string {
	kl := []interface{}{}
	val := reflect.ValueOf(kvmap)
	if val.Kind() == reflect.Map {
		for _, e := range val.MapKeys() {
			v := val.MapIndex(e)
			kl = append(kl, v.Interface())
		}
	}
	kld, _ := json.Marshal(kl)
	return string(kld)
}

func mismatchLenError(tag string, wants, current int) string {
	return fmt.Sprintf("%s wants len %d but current is %d", tag, wants, current)
}

func keyIsNotExtistInMap(tag, key string, kvmap interface{}) string {
	return fmt.Sprintf("%s must present key %s in keys list %s", tag, key, getMapKeys(kvmap))
}

func keyIsNotExtistInSlice(tag, key string, klist interface{}) string {
	kld, _ := json.Marshal(klist)
	return fmt.Sprintf("%s must present key %s in keys list %s", tag, key, string(kld))
}

func keysAreNotExtistInSlice(tag, lkeys, rkeys interface{}) string {
	lkeysd, _ := json.Marshal(lkeys)
	rkeysd, _ := json.Marshal(rkeys)
	return fmt.Sprintf("%s must all keys present %s in keys list %s", tag, string(lkeysd), string(rkeysd))
}

func contextError(tag string, id string, ctx interface{}) string {
	ctxd, _ := json.Marshal(ctx)
	return fmt.Sprintf("%s with %s ctx %s", tag, id, string(ctxd))
}

func caughtValidationError(tag string, err error) string {
	return fmt.Sprintf("%s caught error %s", tag, err.Error())
}

func caughtSchemaValidationError(tag string, errs []gojsonschema.ResultError) string {
	var arr []string
	for _, err := range errs {
		arr = append(arr, err.String())
	}
	errd, _ := json.Marshal(arr)
	return fmt.Sprintf("%s caught errors %s", tag, string(errd))
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
	_ = validate.RegisterValidation("vmail", emailValidatorString())
	_ = validate.RegisterValidation("oauth_min_scope", oauthMinScope())
	_ = validate.RegisterValidation("valid", deepValidator())

	// Check validation interface for all models
	_, _ = reflect.ValueOf(Login{}).Interface().(IValid)
	_, _ = reflect.ValueOf(AuthCallback{}).Interface().(IValid)

	_, _ = reflect.ValueOf(User{}).Interface().(IValid)
	_, _ = reflect.ValueOf(Password{}).Interface().(IValid)

	_, _ = reflect.ValueOf(Role{}).Interface().(IValid)
}
