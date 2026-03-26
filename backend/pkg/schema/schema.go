package schema

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/xeipuuv/gojsonschema"
)

var validate *validator.Validate

// IValid is interface to control all models from user code
type IValid interface {
	Valid() error
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

func init() {
	validate = validator.New()

	_ = validate.RegisterValidation("valid", deepValidator())

	_, _ = reflect.ValueOf(Schema{}).Interface().(IValid)
}

// Schema is the root schema.
// RFC draft-wright-json-schema-00, section 4.5
type Schema struct {
	ID   string `json:"$id,omitempty"`
	Type `json:""`
}

// getValidator is internal function to get validator object
func (sh Schema) getValidator() (*gojsonschema.Schema, error) {
	sl := gojsonschema.NewSchemaLoader()
	sl.Draft = gojsonschema.Draft7
	sl.AutoDetect = false

	var err error
	var rs *gojsonschema.Schema
	if rs, err = sl.Compile(gojsonschema.NewGoLoader(sh)); err != nil {
		return nil, err
	}
	return rs, nil
}

// validate is function to validate input JSON document in bytes
func (sh Schema) validate(l gojsonschema.JSONLoader) (*gojsonschema.Result, error) {
	if rs, err := sh.getValidator(); err != nil {
		return nil, err
	} else if res, err := rs.Validate(l); err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

// GetValidator is function to return validator object
func (sh Schema) GetValidator() (*gojsonschema.Schema, error) {
	return sh.getValidator()
}

// ValidateString is function to validate input string of JSON document
func (sh Schema) ValidateString(doc string) (*gojsonschema.Result, error) {
	docl := gojsonschema.NewStringLoader(string(doc))
	return sh.validate(docl)
}

// ValidateBytes is function to validate input bytes of JSON document
func (sh Schema) ValidateBytes(doc []byte) (*gojsonschema.Result, error) {
	docl := gojsonschema.NewStringLoader(string(doc))
	return sh.validate(docl)
}

// ValidateGo is function to validate input interface of golang object as a JSON document
func (sh Schema) ValidateGo(doc interface{}) (*gojsonschema.Result, error) {
	docl := gojsonschema.NewGoLoader(doc)
	return sh.validate(docl)
}

// Valid is function to control input/output data
func (sh Schema) Valid() error {
	if err := validate.Struct(sh); err != nil {
		return err
	}
	if _, err := sh.getValidator(); err != nil {
		return err
	}
	return nil
}

// Value is interface function to return current value to store to DB
func (sh Schema) Value() (driver.Value, error) {
	b, err := json.Marshal(sh)
	return string(b), err
}

// Scan is interface function to parse DB value when getting from DB
func (sh *Schema) Scan(input interface{}) error {
	return scanFromJSON(input, sh)
}

// Definitions hold schema definitions.
// http://json-schema.org/latest/json-schema-validation.html#rfc.section.5.26
// RFC draft-wright-json-schema-validation-00, section 5.26
type Definitions map[string]*Type

// Type represents a JSON Schema object type.
type Type struct {
	// RFC draft-wright-json-schema-00
	Version string `json:"$schema,omitempty"` // section 6.1
	Ref     string `json:"$ref,omitempty"`    // section 7
	// RFC draft-wright-json-schema-validation-00, section 5
	MultipleOf           int              `json:"multipleOf,omitempty"`           // section 5.1
	Maximum              float64          `json:"maximum,omitempty"`              // section 5.2
	ExclusiveMaximum     bool             `json:"exclusiveMaximum,omitempty"`     // section 5.3
	Minimum              float64          `json:"minimum,omitempty"`              // section 5.4
	ExclusiveMinimum     bool             `json:"exclusiveMinimum,omitempty"`     // section 5.5
	MaxLength            int              `json:"maxLength,omitempty"`            // section 5.6
	MinLength            int              `json:"minLength,omitempty"`            // section 5.7
	Pattern              string           `json:"pattern,omitempty"`              // section 5.8
	AdditionalItems      *Type            `json:"additionalItems,omitempty"`      // section 5.9
	Items                *Type            `json:"items,omitempty"`                // section 5.9
	MaxItems             int              `json:"maxItems,omitempty"`             // section 5.10
	MinItems             int              `json:"minItems,omitempty"`             // section 5.11
	UniqueItems          bool             `json:"uniqueItems,omitempty"`          // section 5.12
	MaxProperties        int              `json:"maxProperties,omitempty"`        // section 5.13
	MinProperties        int              `json:"minProperties,omitempty"`        // section 5.14
	Required             []string         `json:"required,omitempty"`             // section 5.15
	Properties           map[string]*Type `json:"properties,omitempty"`           // section 5.16
	PatternProperties    map[string]*Type `json:"patternProperties,omitempty"`    // section 5.17
	AdditionalProperties json.RawMessage  `json:"additionalProperties,omitempty"` // section 5.18
	Dependencies         map[string]*Type `json:"dependencies,omitempty"`         // section 5.19
	Enum                 []interface{}    `json:"enum,omitempty"`                 // section 5.20
	Type                 string           `json:"type,omitempty"`                 // section 5.21
	AllOf                []*Type          `json:"allOf,omitempty"`                // section 5.22
	AnyOf                []*Type          `json:"anyOf,omitempty"`                // section 5.23
	OneOf                []*Type          `json:"oneOf,omitempty"`                // section 5.24
	Not                  *Type            `json:"not,omitempty"`                  // section 5.25
	Definitions          Definitions      `json:"definitions,omitempty"`          // section 5.26
	// RFC draft-wright-json-schema-validation-00, section 6, 7
	Title       string      `json:"title,omitempty"`       // section 6.1
	Description string      `json:"description,omitempty"` // section 6.1
	Default     interface{} `json:"default,omitempty"`     // section 6.2
	Format      string      `json:"format,omitempty"`      // section 7
	// RFC draft-wright-json-schema-hyperschema-00, section 4
	Media          *Type  `json:"media,omitempty"`          // section 4.3
	BinaryEncoding string `json:"binaryEncoding,omitempty"` // section 4.3

	// extended properties
	ExtProps map[string]interface{} `json:"-"`
}

// MarshalJSON is a JSON interface function to make JSON data bytes array from the struct object
func (t Type) MarshalJSON() ([]byte, error) {
	var err error
	var data []byte
	raw := make(map[string]interface{})
	type tn Type
	if data, err = json.Marshal((*tn)(&t)); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	for k, v := range t.ExtProps {
		raw[k] = v
	}
	if _, ok := raw["properties"]; t.Type == "object" && !ok {
		raw["properties"] = make(map[string]*Type)
	}
	if _, ok := raw["required"]; t.Type == "object" && !ok {
		raw["required"] = []string{}
	}
	return json.Marshal(raw)
}

// UnmarshalJSON is a JSON interface function to parse JSON data bytes array and to get struct object
func (t *Type) UnmarshalJSON(input []byte) error {
	var excludeKeys []string
	tp := reflect.TypeOf(Type{})
	for i := 0; i < tp.NumField(); i++ {
		field := tp.Field(i)
		excludeKeys = append(excludeKeys, strings.Split(field.Tag.Get("json"), ",")[0])
	}
	type tn Type
	if err := json.Unmarshal(input, (*tn)(t)); err != nil {
		return err
	}
	raw := make(map[string]interface{})
	if err := json.Unmarshal(input, &raw); err != nil {
		return err
	}
	t.ExtProps = make(map[string]interface{})
	for k, v := range raw {
		if !slices.Contains(excludeKeys, k) {
			t.ExtProps[k] = v
		}
	}
	return nil
}
