package tester

import (
	"pentagi/pkg/providers/tester/testdata"
)

type AgentTestResults []testdata.TestResult

type ProviderTestResults struct {
	Simple       AgentTestResults `json:"simple"`
	SimpleJSON   AgentTestResults `json:"simpleJson"`
	PrimaryAgent AgentTestResults `json:"primary_agent"`
	Assistant    AgentTestResults `json:"assistant"`
	Generator    AgentTestResults `json:"generator"`
	Refiner      AgentTestResults `json:"refiner"`
	Adviser      AgentTestResults `json:"adviser"`
	Reflector    AgentTestResults `json:"reflector"`
	Searcher     AgentTestResults `json:"searcher"`
	Enricher     AgentTestResults `json:"enricher"`
	Coder        AgentTestResults `json:"coder"`
	Installer    AgentTestResults `json:"installer"`
	Pentester    AgentTestResults `json:"pentester"`
}
