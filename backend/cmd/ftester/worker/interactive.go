package worker

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"pentagi/pkg/terminal"
)

// InteractiveFillArgs interactively fills in missing function arguments
func InteractiveFillArgs(ctx context.Context, funcName string, taskID, subtaskID *int64) (any, error) {
	// Get function information
	funcInfo, err := GetFunctionInfo(funcName)
	if err != nil {
		return nil, err
	}

	// Special handling for the describe function
	if funcName == "describe" {
		params := &DescribeParams{}

		result, err := terminal.GetYesNoInputContext(ctx, "Enable verbose mode", os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("input cancelled: %w", err)
		}
		params.Verbose = result

		return params, nil
	}

	// Get the structure type for the function
	structType, err := getStructTypeForFunction(funcName)
	if err != nil {
		return nil, err
	}

	// Create a new instance of the structure
	structValue := reflect.New(structType).Interface()

	// Create a map to store argument values
	parsedArgs := make(map[string]any)

	terminal.PrintHeader("Interactive argument input for function: " + funcName)
	terminal.PrintInfo("Please enter values for the following arguments:")
	fmt.Println()

	// Request values for each argument
	for _, arg := range funcInfo.Arguments {
		description := arg.Description
		if arg.Default != "" {
			description += fmt.Sprintf(" (default: %v)", arg.Default)
		}
		if len(arg.Enum) > 0 {
			description += fmt.Sprintf(" (enum: %v)", arg.Enum)
		}
		terminal.PrintHeader(description)

		title := arg.Name
		if arg.Required && arg.Name != "message" {
			title += " (required)"
		}

		// Request value from the user
		var value any

		switch arg.Type {
		case "boolean":
			result, err := terminal.GetYesNoInputContext(ctx, title, os.Stdin)
			if err != nil {
				return nil, fmt.Errorf("input cancelled for '%s': %w", arg.Name, err)
			}
			value = result

		case "integer", "number":
			if arg.Name == "task_id" && taskID != nil {
				terminal.PrintKeyValueFormat("Task ID", "%d", *taskID)
				value = *taskID
				break
			}
			if arg.Name == "subtask_id" && subtaskID != nil {
				terminal.PrintKeyValueFormat("Subtask ID", "%d", *subtaskID)
				value = *subtaskID
				break
			}
			for {
				strValue, err := terminal.InteractivePromptContext(ctx, title, os.Stdin)
				if err != nil {
					return nil, fmt.Errorf("input cancelled for '%s': %w", arg.Name, err)
				}

				if strValue == "" && !arg.Required {
					break
				}

				intValue, err := strconv.Atoi(strValue)
				if err != nil {
					terminal.PrintError("Please enter a valid number")
					continue
				}

				value = intValue
				break
			}

		case "array":
			var items []string
			for {
				itemTitle := fmt.Sprintf("%s item #%d (leave empty to finish)", title, len(items)+1)
				strValue, err := terminal.InteractivePromptContext(ctx, itemTitle, os.Stdin)
				if err != nil {
					return nil, fmt.Errorf("input cancelled for '%s': %w", arg.Name, err)
				}

				if strValue == "" {
					if len(items) == 0 && arg.Required {
						terminal.PrintError("At least one value is required")
						continue
					}
					break
				}

				items = append(items, strValue)
			}

			if len(items) > 0 {
				value = items
			}

		default: // string and other types
			strValue, err := terminal.InteractivePromptContext(ctx, title, os.Stdin)
			if err != nil {
				return nil, fmt.Errorf("input cancelled for '%s': %w", arg.Name, err)
			}

			value = strValue
			if value == "" && arg.Required && arg.Name == "message" {
				value = "dummy message"
			}
		}

		// If a value is entered, add it to the map
		if value != nil {
			parsedArgs[arg.Name] = value
		}
	}

	// Check that all required arguments are provided
	for _, arg := range funcInfo.Arguments {
		if arg.Required {
			if _, ok := parsedArgs[arg.Name]; !ok {
				return nil, fmt.Errorf("missing required argument: %s", arg.Name)
			}
		}
	}

	// Convert parsedArgs to a structure
	err = fillStructFromMap(structValue, parsedArgs)
	if err != nil {
		return nil, fmt.Errorf("error filling structure: %w", err)
	}

	return structValue, nil
}

// fillStructFromMap fills a structure with data from a map. Every conversion
// uses a safe (two-value) type assertion so a mismatch between the collected
// value and the field's Go type returns a clear error instead of panicking.
func fillStructFromMap(structPtr any, data map[string]any) error {
	val := reflect.ValueOf(structPtr).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldName := field.Tag.Get("json")

		// If the json tag is not set, use the field name
		if fieldName == "" {
			fieldName = field.Name
		}

		// Remove optional parts of the json tag
		if comma := strings.Index(fieldName, ","); comma != -1 {
			fieldName = fieldName[:comma]
		}

		value, ok := data[fieldName]
		if !ok {
			continue
		}

		fieldValue := val.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		switch fieldValue.Kind() {
		case reflect.String:
			strValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("unexpected type for argument '%s': expected string, got %T", fieldName, value)
			}
			fieldValue.SetString(strValue)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue, ok := value.(int)
			if !ok {
				return fmt.Errorf("unexpected type for argument '%s': expected int, got %T", fieldName, value)
			}
			fieldValue.SetInt(int64(intValue))

		case reflect.Bool:
			boolValue, ok := value.(bool)
			if !ok {
				return fmt.Errorf("unexpected type for argument '%s': expected bool, got %T", fieldName, value)
			}
			fieldValue.SetBool(boolValue)

		case reflect.Slice:
			if fieldValue.Type().Elem().Kind() != reflect.String {
				return fmt.Errorf("unsupported slice element type for argument '%s': %s", fieldName, fieldValue.Type().Elem().Kind())
			}

			// Build the slice explicitly against the field's own type (e.g.
			// tools.Strings, not a plain []string) rather than relying on Go's
			// slice-assignability rules, so this keeps working correctly even
			// if the field is a named type with custom (un)marshaling.
			var items []string
			switch v := value.(type) {
			case []string:
				items = v
			case string:
				items = []string{v}
			default:
				return fmt.Errorf("unexpected type for argument '%s': expected []string or string, got %T", fieldName, value)
			}

			converted := reflect.MakeSlice(fieldValue.Type(), len(items), len(items))
			for idx, item := range items {
				converted.Index(idx).SetString(item)
			}
			fieldValue.Set(converted)

		case reflect.Struct:
			// For special types that may be in the tools package
			// This is a simplified version that may require refinement
			// depending on specific types
			fmt.Printf("Complex structure field detected: %s\n", fieldName)
		}
	}

	return nil
}
