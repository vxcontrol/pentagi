package main

import (
	"testing"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester"
	"pentagi/pkg/providers/tester/cases"
)

type namedModelProvider struct{ provider.Provider }

func (namedModelProvider) Model(pconfig.ProviderOptionsType) string { return "probe-model" }

func TestMain_ConvertToAgentResults_KeepsStopReasonsAndCountsFilteredFailures(t *testing.T) {
	results := tester.ProviderTestResults{
		Simple: tester.AgentTestResults{
			{Name: "refused call", Group: cases.TestGroupBasic, StopReason: "content_filter"},
			{Name: "moderated", Group: cases.TestGroupAdvanced, ContentFiltered: true},
			{Name: "broken", Group: cases.TestGroupAdvanced},
			{Name: "fine", Group: cases.TestGroupBasic, Success: true},
		},
	}

	converted := convertToAgentResults(results, namedModelProvider{})
	if len(converted) != 1 {
		t.Fatalf("converted = %+v, want one agent", converted)
	}

	agent := converted[0]
	if len(agent.BasicTests) != 2 || agent.BasicTests[0].StopReason != "content_filter" {
		t.Errorf("basic tests = %+v, want the runner's stop reason kept on the refused call", agent.BasicTests)
	}
	if agent.TotalFiltered != 1 {
		t.Errorf("TotalFiltered = %d, want 1: a stop reason alone is not the vendor's filter", agent.TotalFiltered)
	}
	if agent.TotalTests != 4 || agent.TotalSuccess != 1 {
		t.Errorf("tally = %d/%d, want 1/4: a filtered case is still a failure", agent.TotalSuccess, agent.TotalTests)
	}
	if len(agent.AdvancedTests) != 2 || !agent.AdvancedTests[0].ContentFiltered || agent.AdvancedTests[1].ContentFiltered {
		t.Errorf("advanced tests = %+v, want the mark on the moderated one only", agent.AdvancedTests)
	}
}
