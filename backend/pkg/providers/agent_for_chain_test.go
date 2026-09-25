package providers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/providers/pconfig"
)

// Read from models.go, since a hardcoded list would miss a chain type added without a mapping.
func agentForChainTypesInSource(t *testing.T) []database.MsgchainType {
	t.Helper()

	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the test file")
	}
	path := filepath.Join(filepath.Dir(here), "..", "database", "models.go")

	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	var found []database.MsgchainType
	ast.Inspect(file, func(n ast.Node) bool {
		spec, isValue := n.(*ast.ValueSpec)
		if !isValue {
			return true
		}
		typeName, isIdent := spec.Type.(*ast.Ident)
		if !isIdent || typeName.Name != "MsgchainType" {
			return true
		}
		for _, value := range spec.Values {
			literal, isLiteral := value.(*ast.BasicLit)
			if !isLiteral || literal.Kind != token.STRING {
				continue
			}
			unquoted, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatalf("unquote %s: %v", literal.Value, err)
			}
			found = append(found, database.MsgchainType(unquoted))
		}
		return true
	})

	if len(found) == 0 {
		t.Fatalf("no MsgchainType constants found in %s", path)
	}
	return found
}

// The mapping also names no chain type the database model does not declare.
func TestAgentForChain_AgentByChain_NamesAConfiguredAgentForEveryChainType(t *testing.T) {
	agents := make(map[pconfig.ProviderOptionsType]bool, len(pconfig.AllAgentTypes))
	for _, agent := range pconfig.AllAgentTypes {
		agents[agent] = true
	}

	chains := agentForChainTypesInSource(t)
	known := make(map[database.MsgchainType]bool, len(chains))
	for _, chain := range chains {
		known[chain] = true
	}

	for _, chain := range chains {
		agent, mapped := agentByChain[chain]
		if !mapped {
			t.Errorf("chain %q names no agent: the run would fall back to the zero agent type", chain)
			continue
		}
		if !agents[agent] {
			t.Errorf("chain %q names agent %q, which is not a configured agent", chain, agent)
		}
	}

	for chain := range agentByChain {
		if !known[chain] {
			t.Errorf("the mapping carries %q, which is not a chain type the model declares", chain)
		}
	}
}

// Every other chain runs under the agent of its own name.
func TestAgentForChain_AgentByChain_RunsFourChainsUnderAnotherAgentsName(t *testing.T) {
	borrowed := map[database.MsgchainType]pconfig.ProviderOptionsType{
		database.MsgchainTypeReporter:      pconfig.OptionsTypeSimple,
		database.MsgchainTypeSummarizer:    pconfig.OptionsTypeSimple,
		database.MsgchainTypeToolCallFixer: pconfig.OptionsTypeSimpleJSON,
		database.MsgchainTypeMemorist:      pconfig.OptionsTypeSearcher,
	}

	for chain, got := range agentByChain {
		want, isBorrowed := borrowed[chain]
		if !isBorrowed {
			want = pconfig.ProviderOptionsType(chain)
		}
		if got != want {
			t.Errorf("chain %q runs under agent %q, want %q", chain, got, want)
		}
	}
	for chain := range borrowed {
		if _, mapped := agentByChain[chain]; !mapped {
			t.Errorf("chain %q names no agent, want %q", chain, borrowed[chain])
		}
	}
}
