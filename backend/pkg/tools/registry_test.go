package tools

import (
	"slices"
	"testing"

	"pentagi/pkg/database"
)

func TestToolTypeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		toolType ToolType
		want     string
	}{
		{NoneToolType, "none"},
		{EnvironmentToolType, "environment"},
		{SearchNetworkToolType, "search_network"},
		{SearchVectorDbToolType, "search_vector_db"},
		{AgentToolType, "agent"},
		{StoreAgentResultToolType, "store_agent_result"},
		{StoreVectorDbToolType, "store_vector_db"},
		{BarrierToolType, "barrier"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			if got := tt.toolType.String(); got != tt.want {
				t.Errorf("ToolType(%d).String() = %q, want %q", tt.toolType, got, tt.want)
			}
		})
	}
}

func TestGetToolType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		toolName string
		want     ToolType
	}{
		{name: "terminal", toolName: TerminalToolName, want: EnvironmentToolType},
		{name: "file", toolName: FileToolName, want: EnvironmentToolType},
		{name: "google", toolName: GoogleToolName, want: SearchNetworkToolType},
		{name: "duckduckgo", toolName: DuckDuckGoToolName, want: SearchNetworkToolType},
		{name: "tavily", toolName: TavilyToolName, want: SearchNetworkToolType},
		{name: "browser", toolName: BrowserToolName, want: SearchNetworkToolType},
		{name: "perplexity", toolName: PerplexityToolName, want: SearchNetworkToolType},
		{name: "sploitus", toolName: SploitusToolName, want: SearchNetworkToolType},
		{name: "search_in_memory", toolName: SearchInMemoryToolName, want: SearchVectorDbToolType},
		{name: "graphiti_search", toolName: GraphitiSearchToolName, want: SearchVectorDbToolType},
		{name: "search agent", toolName: SearchToolName, want: AgentToolType},
		{name: "maintenance", toolName: MaintenanceToolName, want: AgentToolType},
		{name: "coder", toolName: CoderToolName, want: AgentToolType},
		{name: "pentester", toolName: PentesterToolName, want: AgentToolType},
		{name: "done barrier", toolName: FinalyToolName, want: BarrierToolType},
		{name: "ask barrier", toolName: AskUserToolName, want: BarrierToolType},
		{name: "code_result", toolName: CodeResultToolName, want: StoreAgentResultToolType},
		{name: "store_guide", toolName: StoreGuideToolName, want: StoreVectorDbToolType},
		{name: "unknown tool", toolName: "nonexistent_tool", want: NoneToolType},
		{name: "empty string", toolName: "", want: NoneToolType},
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

// TestRegistryDefinitionsCompleteness verifies every tool name in toolsTypeMapping
// has a corresponding entry in registryDefinitions.
func TestRegistryDefinitionsCompleteness(t *testing.T) {
	t.Parallel()

	mapping := GetToolTypeMapping()
	defs := GetRegistryDefinitions()

	for name := range mapping {
		if _, ok := defs[name]; !ok {
			t.Errorf("tool %q is in toolsTypeMapping but missing from registryDefinitions", name)
		}
	}

	// Reverse direction: every tool in registryDefinitions should be in toolsTypeMapping
	for name := range defs {
		if _, ok := mapping[name]; !ok {
			t.Errorf("tool %q is in registryDefinitions but missing from toolsTypeMapping", name)
		}
	}
}

// TestRegistryDefinitionsReturnsCopy verifies that GetRegistryDefinitions returns
// a copy that can be mutated without affecting the original registry.
func TestRegistryDefinitionsReturnsCopy(t *testing.T) {
	t.Parallel()

	const sentinelKey = "test_sentinel"

	defs1 := GetRegistryDefinitions()
	originalLen := len(defs1)
	if _, ok := defs1[sentinelKey]; ok {
		t.Fatalf("precondition failed: sentinel key %q already exists", sentinelKey)
	}
	defer delete(defs1, sentinelKey)

	defs1[sentinelKey] = defs1[TerminalToolName]

	defs2 := GetRegistryDefinitions()
	if len(defs2) != originalLen {
		t.Errorf("mutation leaked: original len = %d, new len = %d", originalLen, len(defs2))
	}
	if _, ok := defs2[sentinelKey]; ok {
		t.Error("mutation leaked: test_sentinel found in fresh copy")
	}
}

// TestToolTypeMappingReturnsCopy verifies that GetToolTypeMapping returns a copy.
func TestToolTypeMappingReturnsCopy(t *testing.T) {
	t.Parallel()

	const sentinelKey = "test_sentinel"

	m1 := GetToolTypeMapping()
	originalLen := len(m1)
	if _, ok := m1[sentinelKey]; ok {
		t.Fatalf("precondition failed: sentinel key %q already exists", sentinelKey)
	}
	defer delete(m1, sentinelKey)

	m1[sentinelKey] = NoneToolType

	m2 := GetToolTypeMapping()
	if len(m2) != originalLen {
		t.Errorf("mutation leaked: original len = %d, new len = %d", originalLen, len(m2))
	}
	if _, ok := m2[sentinelKey]; ok {
		t.Error("mutation leaked: test_sentinel found in fresh mapping copy")
	}
}

// TestGetToolsByType verifies the reverse mapping is consistent with the forward mapping.
func TestGetToolsByType(t *testing.T) {
	t.Parallel()

	forward := GetToolTypeMapping()
	reverse := GetToolsByType()

	// Build expected reverse map from forward map
	expected := make(map[ToolType]map[string]struct{})
	for name, toolType := range forward {
		if expected[toolType] == nil {
			expected[toolType] = make(map[string]struct{})
		}
		expected[toolType][name] = struct{}{}
	}

	// Verify all entries in reverse exist in forward
	for toolType, names := range reverse {
		for _, name := range names {
			if forward[name] != toolType {
				t.Errorf("GetToolsByType()[%v] contains %q, but forward mapping says %v", toolType, name, forward[name])
			}
		}
	}

	// Verify counts match
	for toolType, expectedNames := range expected {
		if len(reverse[toolType]) != len(expectedNames) {
			t.Errorf("GetToolsByType()[%v] has %d entries, want %d", toolType, len(reverse[toolType]), len(expectedNames))
		}
	}

	// Verify there are no duplicates in each reverse slice
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

// TestRegistryDefinitionNames verifies each definition Name field matches its map key.
func TestRegistryDefinitionNames(t *testing.T) {
	t.Parallel()

	defs := GetRegistryDefinitions()
	for key, def := range defs {
		if def.Name != key {
			t.Errorf("registryDefinitions[%q].Name = %q, want %q", key, def.Name, key)
		}
	}
}

func TestGetMessageType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tool string
		want database.MsglogType
	}{
		{"terminal", TerminalToolName, database.MsglogTypeTerminal},
		{"file", FileToolName, database.MsglogTypeFile},
		{"browser", BrowserToolName, database.MsglogTypeBrowser},
		{"search engine", GoogleToolName, database.MsglogTypeSearch},
		{"advice", AdviceToolName, database.MsglogTypeAdvice},
		{"ask", AskUserToolName, database.MsglogTypeAsk},
		{"done", FinalyToolName, database.MsglogTypeDone},
		{"unknown", "unknown_tool", database.MsglogTypeThoughts},
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

func TestGetMessageResultFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tool string
		want database.MsglogResultFormat
	}{
		{"terminal", TerminalToolName, database.MsglogResultFormatTerminal},
		{"file", FileToolName, database.MsglogResultFormatPlain},
		{"browser", BrowserToolName, database.MsglogResultFormatPlain},
		{"search default", GoogleToolName, database.MsglogResultFormatMarkdown},
		{"unknown default", "unknown_tool", database.MsglogResultFormatMarkdown},
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

func TestAllowedToolListsContainKnownUniqueTools(t *testing.T) {
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

	// Minimal invariant checks for critical tools.
	if !slices.Contains(allowedSummarizingToolsResult, BrowserToolName) {
		t.Errorf("allowedSummarizingToolsResult must contain %q", BrowserToolName)
	}
	if !slices.Contains(allowedStoringInMemoryTools, SearchToolName) {
		t.Errorf("allowedStoringInMemoryTools must contain %q", SearchToolName)
	}
}
