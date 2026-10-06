package smallmodel

import (
	"fmt"
	"net"
	"strings"
)

// Decision is the verifier's verdict on a proposed action.
type Decision int

const (
	// Allow lets the action run unchanged.
	Allow Decision = iota
	// Confirm requires human (or large-model) confirmation before running.
	Confirm
	// Deny blocks the action outright.
	Deny
)

func (d Decision) String() string {
	switch d {
	case Allow:
		return "allow"
	case Confirm:
		return "confirm"
	case Deny:
		return "deny"
	default:
		return "unknown"
	}
}

// Verdict is a Decision with a human-readable reason, suitable for surfacing to
// the operator or feeding back to the model.
type Verdict struct {
	Decision Decision
	Reason   string
}

// defaultDeniedSubstrings mark a shell command as destructive enough to require
// confirmation regardless of scope.
var defaultDeniedSubstrings = []string{
	"rm -rf /",
	"rm -fr /",
	"mkfs",
	"dd if=",
	" dd of=",
	"shutdown",
	"reboot",
	"halt",
	":(){:|:&};:",
	"> /dev/sda",
	"chmod -r 000",
	"wipefs",
}

// Verifier classifies proposed actions before execution. It is deterministic and
// cheap; it is the floor under a small, less reliable model, keeping it inside
// the authorized scope and away from destructive commands.
type Verifier struct {
	scopeHosts []string
	scopeNets  []*net.IPNet
	denied     []string
}

// NewVerifier builds a Verifier from comma-separated config. scopeCSV lists
// allowed IPs, CIDRs or hostnames; an empty scope disables scope enforcement.
// deniedCSV extends the built-in destructive-command list.
func NewVerifier(scopeCSV, deniedCSV string) *Verifier {
	v := &Verifier{denied: append([]string{}, defaultDeniedSubstrings...)}

	for _, entry := range splitCSV(scopeCSV) {
		if _, ipnet, err := net.ParseCIDR(entry); err == nil {
			v.scopeNets = append(v.scopeNets, ipnet)
			continue
		}
		v.scopeHosts = append(v.scopeHosts, strings.ToLower(entry))
	}
	for _, d := range splitCSV(deniedCSV) {
		v.denied = append(v.denied, strings.ToLower(d))
	}
	return v
}

// CheckCommand verdicts a shell command: Deny-worthy destructive patterns become
// Confirm (never a silent block of the agent's own plan), and a target outside
// the configured scope becomes Confirm as well. Everything else is allowed.
func (v *Verifier) CheckCommand(cmd string) Verdict {
	low := strings.ToLower(cmd)

	for _, d := range v.denied {
		if strings.Contains(low, d) {
			return Verdict{Confirm, fmt.Sprintf("destructive pattern %q requires confirmation", strings.TrimSpace(d))}
		}
	}

	if out, ok := v.outOfScopeTarget(cmd); ok {
		return Verdict{Confirm, fmt.Sprintf("target %q is outside the authorized scope", out)}
	}

	return Verdict{Allow, ""}
}

// outOfScopeTarget returns the first IP in the command that is not in scope.
// Hostnames are checked against the host allowlist. With no scope configured it
// always returns ok=false.
func (v *Verifier) outOfScopeTarget(cmd string) (string, bool) {
	if len(v.scopeNets) == 0 && len(v.scopeHosts) == 0 {
		return "", false
	}

	for _, ipStr := range reIPv4.FindAllString(cmd, -1) {
		if !v.ipInScope(ipStr) {
			return ipStr, true
		}
	}
	return "", false
}

func (v *Verifier) ipInScope(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return true
	}
	for _, n := range v.scopeNets {
		if n.Contains(ip) {
			return true
		}
	}
	for _, h := range v.scopeHosts {
		if h == ipStr {
			return true
		}
	}
	return false
}

func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
