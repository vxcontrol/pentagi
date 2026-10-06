package smallmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifier_CheckCommand_AllowsOrdinaryCommand(t *testing.T) {
	v := NewVerifier("", "")
	assert.Equal(t, Allow, v.CheckCommand("nmap -sV 10.0.0.5").Decision)
}

func TestVerifier_CheckCommand_ConfirmsDestructivePattern(t *testing.T) {
	v := NewVerifier("", "")
	got := v.CheckCommand("rm -rf / --no-preserve-root")
	assert.Equal(t, Confirm, got.Decision)
	assert.Contains(t, got.Reason, "destructive")
}

func TestVerifier_CheckCommand_ConfirmsOperatorDeniedSubstring(t *testing.T) {
	v := NewVerifier("", "msfconsole")
	assert.Equal(t, Confirm, v.CheckCommand("msfconsole -q -x run").Decision)
}

func TestVerifier_CheckCommand_ConfirmsOutOfScopeIP(t *testing.T) {
	v := NewVerifier("10.0.0.0/24", "")
	got := v.CheckCommand("nmap -sV 8.8.8.8")
	assert.Equal(t, Confirm, got.Decision)
	assert.Contains(t, got.Reason, "8.8.8.8")
}

func TestVerifier_CheckCommand_AllowsInScopeIP(t *testing.T) {
	v := NewVerifier("10.0.0.0/24", "")
	assert.Equal(t, Allow, v.CheckCommand("nmap -sV 10.0.0.5").Decision)
}

func TestVerifier_CheckCommand_AllowsSingleHostScope(t *testing.T) {
	v := NewVerifier("10.0.0.5", "")
	assert.Equal(t, Allow, v.CheckCommand("curl http://10.0.0.5:8443").Decision)
	assert.Equal(t, Confirm, v.CheckCommand("curl http://10.0.0.6").Decision)
}

func TestVerifier_CheckCommand_NoScopeMeansNoScopeEnforcement(t *testing.T) {
	v := NewVerifier("", "")
	assert.Equal(t, Allow, v.CheckCommand("nmap 1.2.3.4").Decision)
}
