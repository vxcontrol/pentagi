package cases

import "testing"

func builtinDefinition(t *testing.T, id string) TestDefinition {
	t.Helper()

	registry, err := LoadBuiltinRegistry()
	if err != nil {
		t.Fatalf("loading the builtin registry: %v", err)
	}
	for _, def := range registry.GetAllTests() {
		if def.ID == id {
			return def
		}
	}
	t.Fatalf("the builtin registry has no case %q", id)
	return TestDefinition{}
}

func testerDefinitions(t *testing.T, yaml string) []TestDefinition {
	t.Helper()

	registry, err := LoadRegistryFromYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return registry.GetAllTests()
}
