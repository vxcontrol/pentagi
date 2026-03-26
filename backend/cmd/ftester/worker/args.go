package worker

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"pentagi/pkg/tools"
)

// FunctionInfo represents information about a function and its arguments
type FunctionInfo struct {
	Name        string
	Description string
	Arguments   []ArgumentInfo
}

// ArgumentInfo represents information about a function argument
type ArgumentInfo struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Default     any
	Enum        []any
}

// DescribeParams contains the parameters for the describe function
type DescribeParams struct {
	Verbose bool `json:"verbose"`
}

var describeFuncInfo = FunctionInfo{
	Name:        "describe",
	Description: "Display information about tasks and subtasks for the given flow ID with optional filtering",
	Arguments: []ArgumentInfo{
		{
			Name:        "verbose",
			Type:        "boolean",
			Description: "Display full descriptions and results",
			Required:    false,
		},
	},
}

// GetAvailableFunctions returns all available functions with their descriptions
func GetAvailableFunctions() []FunctionInfo {
	funcInfos := []FunctionInfo{}

	for name, def := range tools.GetRegistryDefinitions() {
		// Skip functions that are not available for user invocation
		if !isToolAvailableForCall(name) {
			continue
		}

		funcInfo := FunctionInfo{
			Name:        name,
			Description: def.Description,
		}
		funcInfos = append(funcInfos, funcInfo)
	}

	// Add custom ftester functions
	funcInfos = append(funcInfos, describeFuncInfo)

	return funcInfos
}

// GetFunctionInfo returns information about a specific function
func GetFunctionInfo(funcName string) (FunctionInfo, error) {
	// Check for custom ftester functions
	if funcName == "describe" {
		return describeFuncInfo, nil
	}

	definitions := tools.GetRegistryDefinitions()

	def, ok := definitions[funcName]
	if !ok {
		return FunctionInfo{}, fmt.Errorf("function not found: %s", funcName)
	}

	// Check if the function is available for user invocation
	if !isToolAvailableForCall(funcName) {
		return FunctionInfo{}, fmt.Errorf("function not available for user invocation: %s", funcName)
	}

	fi := FunctionInfo{
		Name:        def.Name,
		Description: def.Description,
		Arguments:   []ArgumentInfo{},
	}

	// Extract argument info from the schema
	if def.Parameters == nil {
		return fi, nil
	}

	// Handle the schema based on its actual type
	var schemaObj map[string]any

	// Check if it's already a map
	if rawMap, ok := def.Parameters.(map[string]any); ok {
		schemaObj = rawMap
	} else {
		// It might be a jsonschema.Schema or something else that needs to be marshaled
		schemaBytes, err := json.Marshal(def.Parameters)
		if err != nil {
			return fi, fmt.Errorf("error marshaling schema: %w", err)
		}

		if err := json.Unmarshal(schemaBytes, &schemaObj); err != nil {
			return fi, fmt.Errorf("error unmarshaling schema: %w", err)
		}
	}

	// Now parse the properties
	if properties, ok := schemaObj["properties"].(map[string]any); ok {
		for propName, propInfo := range properties {
			propMap, ok := propInfo.(map[string]any)
			if !ok {
				continue
			}

			argType := "string"
			if typeInfo, ok := propMap["type"]; ok {
				argType = fmt.Sprintf("%v", typeInfo)
			}

			description := ""
			if descInfo, ok := propMap["description"]; ok {
				description = fmt.Sprintf("%v", descInfo)
			}

			required := false
			if requiredFields, ok := schemaObj["required"].([]any); ok {
				for _, reqField := range requiredFields {
					if reqField.(string) == propName {
						required = true
						break
					}
				}
			}

			defaultVal := ""
			if defaultInfo, ok := propMap["default"]; ok {
				defaultVal = fmt.Sprintf("%v", defaultInfo)
			}

			enumValues := []any{}
			if enumInfo, ok := propMap["enum"]; ok {
				enumValues = enumInfo.([]any)
			}

			fi.Arguments = append(fi.Arguments, ArgumentInfo{
				Name:        propName,
				Type:        argType,
				Description: description,
				Required:    required,
				Default:     defaultVal,
				Enum:        enumValues,
			})
		}

		slices.SortFunc(fi.Arguments, func(a, b ArgumentInfo) int {
			return strings.Compare(a.Name, b.Name)
		})
	}

	return fi, nil
}

// ParseFunctionArgs parses command-line arguments into a structured object for the function
func ParseFunctionArgs(funcName string, args []string) (any, error) {
	// Handle describe function specially
	if funcName == "describe" {
		params := &DescribeParams{}

		// Parse the command-line arguments for describe
		for i := 0; i < len(args); i++ {
			arg := args[i]

			// Check if the arg starts with '-'
			if !strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("invalid argument format (expected '-name'): %s", arg)
			}

			// Get the argument name without '-'
			argName := strings.TrimPrefix(arg, "-")

			switch argName {
			case "verbose":
				params.Verbose = true
			default:
				return nil, fmt.Errorf("unknown argument for describe: %s", argName)
			}
		}

		return params, nil
	}

	// Get function info to check required arguments
	funcInfo, err := GetFunctionInfo(funcName)
	if err != nil {
		return nil, err
	}

	// Create a map to store parsed args
	parsedArgs := make(map[string]any)

	// Parse the command-line arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Check if the arg starts with '-'
		if !strings.HasPrefix(arg, "-") {
			return nil, fmt.Errorf("invalid argument format (expected '-name'): %s", arg)
		}

		// Get the argument name without '-'
		argName := strings.TrimPrefix(arg, "-")

		// Find the argument info
		var argInfo *ArgumentInfo
		for _, ai := range funcInfo.Arguments {
			if ai.Name == argName {
				argInfo = &ai
				break
			}
		}

		if argInfo == nil {
			return nil, fmt.Errorf("unknown argument: %s", argName)
		}

		// Check if there's a value for the argument
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			// Next arg is the value
			parsedArgs[argName] = args[i+1]
			i++ // Skip the value in the next iteration
		} else {
			// Boolean flag (no value)
			parsedArgs[argName] = true
		}
	}

	// Check if all required arguments are provided
	for _, arg := range funcInfo.Arguments {
		if arg.Required {
			if _, ok := parsedArgs[arg.Name]; !ok {
				if arg.Name == "message" {
					parsedArgs[arg.Name] = "dummy message"
					continue
				}
				return nil, fmt.Errorf("missing required argument: %s", arg.Name)
			}
		}
	}

	// Find the appropriate struct type for the function
	structType, err := getStructTypeForFunction(funcName)
	if err != nil {
		return nil, err
	}

	// Create a new instance of the struct
	structValue := reflect.New(structType).Interface()

	// Convert parsedArgs to JSON
	jsonData, err := json.Marshal(parsedArgs)
	if err != nil {
		return nil, fmt.Errorf("error marshaling arguments: %w", err)
	}

	// Unmarshal JSON into the struct
	err = json.Unmarshal(jsonData, structValue)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling arguments: %w", err)
	}

	return structValue, nil
}

// getStructTypeForFunction finds the appropriate struct type for a function
func getStructTypeForFunction(funcName string) (reflect.Type, error) {
	// Map function names to struct types
	typeMap := map[string]any{
		tools.TerminalToolName:          &tools.TerminalAction{},
		tools.FileToolName:              &tools.FileAction{},
		tools.BrowserToolName:           &tools.Browser{},
		tools.GoogleToolName:            &tools.SearchAction{},
		tools.DuckDuckGoToolName:        &tools.SearchAction{},
		tools.TavilyToolName:            &tools.SearchAction{},
		tools.TraversaalToolName:        &tools.SearchAction{},
		tools.PerplexityToolName:        &tools.SearchAction{},
		tools.SearxngToolName:           &tools.SearchAction{},
		tools.SploitusToolName:          &tools.SploitusAction{},
		tools.MemoristToolName:          &tools.MemoristAction{},
		tools.SearchInMemoryToolName:    &tools.SearchInMemoryAction{},
		tools.SearchGuideToolName:       &tools.SearchGuideAction{},
		tools.StoreGuideToolName:        &tools.StoreGuideAction{},
		tools.SearchAnswerToolName:      &tools.SearchAnswerAction{},
		tools.StoreAnswerToolName:       &tools.StoreAnswerAction{},
		tools.SearchCodeToolName:        &tools.SearchCodeAction{},
		tools.StoreCodeToolName:         &tools.StoreCodeAction{},
		tools.GraphitiSearchToolName:    &tools.GraphitiSearchAction{},
		tools.SearchToolName:            &tools.ComplexSearch{},
		tools.MaintenanceToolName:       &tools.MaintenanceAction{},
		tools.CoderToolName:             &tools.CoderAction{},
		tools.PentesterToolName:         &tools.PentesterAction{},
		tools.AdviceToolName:            &tools.AskAdvice{},
		tools.FinalyToolName:            &tools.Done{},
		tools.AskUserToolName:           &tools.AskUser{},
		tools.SearchResultToolName:      &tools.SearchResult{},
		tools.MemoristResultToolName:    &tools.MemoristResult{},
		tools.MaintenanceResultToolName: &tools.TaskResult{},
		tools.CodeResultToolName:        &tools.CodeResult{},
		tools.HackResultToolName:        &tools.HackResult{},
		tools.EnricherResultToolName:    &tools.EnricherResult{},
		tools.ReportResultToolName:      &tools.TaskResult{},
		tools.SubtaskListToolName:       &tools.SubtaskList{},
	}

	structType, ok := typeMap[funcName]
	if !ok {
		return nil, fmt.Errorf("no struct type found for function: %s", funcName)
	}

	return reflect.TypeOf(structType).Elem(), nil
}

// IsToolAvailableForCall checks if a tool is available for call from the command line
func isToolAvailableForCall(toolName string) bool {
	toolsMapping := tools.GetToolsByType()
	availableTools := map[string]struct{}{}
	for toolType, toolsList := range toolsMapping {
		switch toolType {
		case tools.NoneToolType, tools.StoreAgentResultToolType,
			tools.StoreVectorDbToolType, tools.BarrierToolType:
			continue
		default:
			for _, tool := range toolsList {
				availableTools[tool] = struct{}{}
			}
		}
	}
	_, ok := availableTools[toolName]
	return ok
}
