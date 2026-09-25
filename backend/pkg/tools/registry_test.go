package tools

import (
	"slices"
	"testing"

	"pentagi/pkg/database"
)

func TestRegistry_ToolType_StringNamesEveryType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		toolType ToolType
		want     string
	}{
		{"the zero type", NoneToolType, "none"},
		{"the environment type", EnvironmentToolType, "environment"},
		{"the network search type", SearchNetworkToolType, "search_network"},
		{"the vector search type", SearchVectorDbToolType, "search_vector_db"},
		{"the agent type", AgentToolType, "agent"},
		{"the agent result type", StoreAgentResultToolType, "store_agent_result"},
		{"the vector store type", StoreVectorDbToolType, "store_vector_db"},
		{"the barrier type", BarrierToolType, "barrier"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.toolType.String(); got != tt.want {
				t.Errorf("ToolType(%d).String() = %q, want %q", tt.toolType, got, tt.want)
			}
		})
	}
}

func TestRegistry_GetToolType_ClassifiesEachTool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		toolName string
		want     ToolType
	}{
		{name: "the terminal acts on the environment", toolName: TerminalToolName, want: EnvironmentToolType},
		{name: "the file tool acts on the environment", toolName: FileToolName, want: EnvironmentToolType},
		{name: "google searches the network", toolName: GoogleToolName, want: SearchNetworkToolType},
		{name: "duckduckgo searches the network", toolName: DuckDuckGoToolName, want: SearchNetworkToolType},
		{name: "tavily searches the network", toolName: TavilyToolName, want: SearchNetworkToolType},
		{name: "the browser searches the network", toolName: BrowserToolName, want: SearchNetworkToolType},
		{name: "perplexity searches the network", toolName: PerplexityToolName, want: SearchNetworkToolType},
		{name: "sploitus searches the network", toolName: SploitusToolName, want: SearchNetworkToolType},
		{name: "the search orchestrator searches the network", toolName: WebSearchToolName, want: SearchNetworkToolType},
		{name: "the memory search reads the vector store", toolName: SearchInMemoryToolName, want: SearchVectorDbToolType},
		{name: "the knowledge graph search reads the vector store", toolName: GraphitiSearchToolName, want: SearchVectorDbToolType},
		{name: "the search agent is an agent", toolName: SearchToolName, want: AgentToolType},
		{name: "the maintenance agent is an agent", toolName: MaintenanceToolName, want: AgentToolType},
		{name: "the coder is an agent", toolName: CoderToolName, want: AgentToolType},
		{name: "the pentester is an agent", toolName: PentesterToolName, want: AgentToolType},
		{name: "finishing is a barrier", toolName: FinalyToolName, want: BarrierToolType},
		{name: "asking the user is a barrier", toolName: AskUserToolName, want: BarrierToolType},
		{name: "reporting the code result stores an agent result", toolName: CodeResultToolName, want: StoreAgentResultToolType},
		{name: "storing a guide writes the vector store", toolName: StoreGuideToolName, want: StoreVectorDbToolType},
		{name: "an unknown tool has no type", toolName: "nonexistent_tool", want: NoneToolType},
		{name: "an empty name has no type", toolName: "", want: NoneToolType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := GetToolType(tt.toolName); got != tt.want {
				t.Errorf("GetToolType(%q) = %v, want %v", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestRegistry_GetRegistryDefinitions_DefinesEveryMappedToolUnderItsOwnName(t *testing.T) {
	t.Parallel()

	mapping := GetToolTypeMapping()
	defs := GetRegistryDefinitions()

	for name := range mapping {
		if _, ok := defs[name]; !ok {
			t.Errorf("tool %q is in toolsTypeMapping but missing from registryDefinitions", name)
		}
	}

	for name, def := range defs {
		if _, ok := mapping[name]; !ok {
			t.Errorf("tool %q is in registryDefinitions but missing from toolsTypeMapping", name)
		}
		if def.Name != name {
			t.Errorf("registryDefinitions[%q].Name = %q, want %q", name, def.Name, name)
		}
	}
}

func TestRegistry_GetRegistryDefinitions_ReturnsACopy(t *testing.T) {
	t.Parallel()
	registryAssertReturnsACopy(t, GetRegistryDefinitions)
}

func TestRegistry_GetToolTypeMapping_ReturnsACopy(t *testing.T) {
	t.Parallel()
	registryAssertReturnsACopy(t, GetToolTypeMapping)
}

func registryAssertReturnsACopy[V any](t *testing.T, get func() map[string]V) {
	t.Helper()

	const sentinelKey = "test_sentinel"

	first := get()
	originalLen := len(first)
	if _, ok := first[sentinelKey]; ok {
		t.Fatalf("precondition failed: sentinel key %q already exists", sentinelKey)
	}
	defer delete(first, sentinelKey)

	var zero V
	first[sentinelKey] = zero

	second := get()
	if len(second) != originalLen {
		t.Errorf("mutation leaked: original len = %d, new len = %d", originalLen, len(second))
	}
	if _, ok := second[sentinelKey]; ok {
		t.Errorf("mutation leaked: %q found in a fresh copy", sentinelKey)
	}
}

func TestRegistry_GetToolsByType_InvertsTheTypeMapping(t *testing.T) {
	t.Parallel()

	forward := GetToolTypeMapping()
	reverse := GetToolsByType()

	expected := make(map[ToolType]map[string]struct{})
	for name, toolType := range forward {
		if expected[toolType] == nil {
			expected[toolType] = make(map[string]struct{})
		}
		expected[toolType][name] = struct{}{}
	}

	for toolType, names := range reverse {
		for _, name := range names {
			if forward[name] != toolType {
				t.Errorf("GetToolsByType()[%v] contains %q, but forward mapping says %v", toolType, name, forward[name])
			}
		}
	}

	for toolType, expectedNames := range expected {
		if len(reverse[toolType]) != len(expectedNames) {
			t.Errorf("GetToolsByType()[%v] has %d entries, want %d", toolType, len(reverse[toolType]), len(expectedNames))
		}
	}

	for toolType, names := range reverse {
		seen := make(map[string]struct{}, len(names))
		for _, name := range names {
			if _, ok := seen[name]; ok {
				t.Errorf("GetToolsByType()[%v] contains duplicate tool %q", toolType, name)
			}
			seen[name] = struct{}{}
		}
	}
}

func TestRegistry_GetMessageType_PicksTheLogTypeOfATool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tool string
		want database.MsglogType
	}{
		{"a terminal call logs as terminal", TerminalToolName, database.MsglogTypeTerminal},
		{"a file call logs as file", FileToolName, database.MsglogTypeFile},
		{"a browser call logs as browser", BrowserToolName, database.MsglogTypeBrowser},
		{"a search engine call logs as search", GoogleToolName, database.MsglogTypeSearch},
		{"a search orchestrator call logs as search", WebSearchToolName, database.MsglogTypeSearch},
		{"asking the adviser logs as advice", AdviceToolName, database.MsglogTypeAdvice},
		{"asking the user logs as ask", AskUserToolName, database.MsglogTypeAsk},
		{"finishing logs as done", FinalyToolName, database.MsglogTypeDone},
		{"an unknown tool logs as thoughts", "unknown_tool", database.MsglogTypeThoughts},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getMessageType(tt.tool); got != tt.want {
				t.Errorf("getMessageType(%q) = %q, want %q", tt.tool, got, tt.want)
			}
		})
	}
}

func TestRegistry_GetMessageResultFormat_PicksTheResultFormatOfATool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tool string
		want database.MsglogResultFormat
	}{
		{"terminal output keeps its terminal format", TerminalToolName, database.MsglogResultFormatTerminal},
		{"the content of a file is plain", FileToolName, database.MsglogResultFormatPlain},
		{"a browsed page is plain", BrowserToolName, database.MsglogResultFormatPlain},
		{"a search result is markdown", GoogleToolName, database.MsglogResultFormatMarkdown},
		{"the result of an unknown tool is markdown", "unknown_tool", database.MsglogResultFormatMarkdown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getMessageResultFormat(tt.tool); got != tt.want {
				t.Errorf("getMessageResultFormat(%q) = %q, want %q", tt.tool, got, tt.want)
			}
		})
	}
}

func TestRegistry_AllowedToolListsNameKnownToolsOnce(t *testing.T) {
	t.Parallel()

	mapping := GetToolTypeMapping()

	validate := func(t *testing.T, listName string, tools []string) {
		t.Helper()

		seen := make(map[string]struct{}, len(tools))
		for _, tool := range tools {
			if _, ok := mapping[tool]; !ok {
				t.Errorf("%s contains unknown tool %q", listName, tool)
			}
			if _, ok := seen[tool]; ok {
				t.Errorf("%s contains duplicate tool %q", listName, tool)
			}
			seen[tool] = struct{}{}
		}
	}

	validate(t, "allowedSummarizingToolsResult", allowedSummarizingToolsResult)
	validate(t, "allowedStoringInMemoryTools", allowedStoringInMemoryTools)

	// Pinned because losing either entry fails no call and only degrades the agents quietly.
	if !slices.Contains(allowedSummarizingToolsResult, BrowserToolName) {
		t.Errorf("allowedSummarizingToolsResult must contain %q", BrowserToolName)
	}
	if !slices.Contains(allowedStoringInMemoryTools, SearchToolName) {
		t.Errorf("allowedStoringInMemoryTools must contain %q", SearchToolName)
	}
}
