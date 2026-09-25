package tester

import (
	"encoding/json"

	"pentagi/pkg/tools"

	"github.com/vxcontrol/langchaingo/llms"
)

// fileToolCallResponse omits empty arguments, as a model sends only the fields it needs.
func fileToolCallResponse(id string, args map[string]string) *llms.ContentResponse {
	argMap := make(map[string]any, len(args))
	for k, v := range args {
		if v != "" {
			argMap[k] = v
		}
	}
	argsJSON, _ := json.Marshal(argMap)

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{
			ToolCalls: []llms.ToolCall{{
				ID:   id,
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      tools.FileToolName,
					Arguments: string(argsJSON),
				},
			}},
		}},
	}
}

// correctFileEditDiff turns FileEditTestOldLine, the third line of FileEditTestContent, into FileEditTestNewLine.
const correctFileEditDiff = "@@ -3,1 +3,1 @@\n-" + FileEditTestOldLine + "\n+" + FileEditTestNewLine + "\n"
