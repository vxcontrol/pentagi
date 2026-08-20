package deafguard

import (
	"fmt"
	"regexp"
)

// RiskLevel represents the severity of a classified command.
type RiskLevel string

const (
	RiskCritical RiskLevel = "critical"
	RiskHigh     RiskLevel = "high"
	RiskMedium   RiskLevel = "medium"
	RiskLow      RiskLevel = "low"
	RiskNone     RiskLevel = "none"
)

// Action represents what the Deaf Guard should do with a command.
type Action string

const (
	ActionBlock Action = "block"
	ActionWarn  Action = "warn"
	ActionLog   Action = "log"
)

// Category represents the classification tier.
type Category string

const (
	CategoryContainerEscape  Category = "container_escape"
	CategoryDestructiveFile  Category = "destructive_file"
	CategoryNetworkDos       Category = "network_dos"
	CategoryPersistence      Category = "persistence"
	CategoryReverseShellExfil Category = "reverse_shell_exfil"
	CategoryCredentialAbuse  Category = "credential_abuse"
	CategoryAggressiveFlags  Category = "aggressive_flags"
	CategoryStandardPentest  Category = "standard_pentest"
	CategoryLocalUtility     Category = "local_utility"
)

// Tier numbers for Settings UI toggle ordering.
var CategoryTier = map[Category]int{
	CategoryContainerEscape:  1,
	CategoryDestructiveFile:  2,
	CategoryNetworkDos:       3,
	CategoryPersistence:      4,
	CategoryReverseShellExfil: 5,
	CategoryCredentialAbuse:  6,
	CategoryAggressiveFlags:  7,
	CategoryStandardPentest:  8,
	CategoryLocalUtility:     9,
}

// CategoryInfo provides human-readable descriptions for the Settings UI.
var CategoryInfo = map[Category]struct {
	Name        string
	Description string
	DefaultAction Action
}{
	CategoryContainerEscape:   {"Container Escape", "Attempts to break out of the Docker sandbox (mount, docker socket, nsenter, chroot)", ActionBlock},
	CategoryDestructiveFile:   {"Destructive File Ops", "Irreversible filesystem destruction (rm -rf /, shred, dd to devices, mkfs)", ActionBlock},
	CategoryNetworkDos:        {"Network DoS", "Flood attacks, fork bombs, stress tools, iptables DROP rules", ActionBlock},
	CategoryPersistence:       {"Persistence / Implants", "Creating backdoors, user accounts, cron jobs, SSH keys, systemd services", ActionBlock},
	CategoryReverseShellExfil: {"Reverse Shells / Exfil", "Reverse shell patterns, file uploads to non-target hosts, DNS exfiltration", ActionWarn},
	CategoryCredentialAbuse:   {"Credential Abuse", "Brute force tools targeting non-scope hosts (hydra, medusa, crackmapexec)", ActionWarn},
	CategoryAggressiveFlags:   {"Aggressive Pentest Flags", "Standard tools with dangerous options (sqlmap --risk=3, nmap DoS scripts)", ActionWarn},
	CategoryStandardPentest:   {"Standard Pentesting", "Normal recon and testing tools with default flags (nmap, nuclei, curl, ffuf)", ActionLog},
	CategoryLocalUtility:      {"Local Utilities", "Container-internal file ops and scripting (cat, grep, python, jq)", ActionLog},
}

// Rule is a single classification rule with a compiled regex pattern.
type Rule struct {
	Pattern  *regexp.Regexp
	Category Category
	Risk     RiskLevel
	Action   Action
	Reason   string
}

// rules is the full set of classification rules, evaluated in order.
// First match wins — rules are ordered from most dangerous to least.
var rules []Rule

func init() {
	type rawRule struct {
		Pattern  string
		Category Category
		Risk     RiskLevel
		Action   Action
		Reason   string
	}

	raw := []rawRule{
		// ── Tier 1: Container Escape ──
		{`\bmount\s+.*/dev/`, CategoryContainerEscape, RiskCritical, ActionBlock, "Mounting block devices may escape container sandbox"},
		{`/var/run/docker\.sock`, CategoryContainerEscape, RiskCritical, ActionBlock, "Docker socket access enables container escape"},
		{`\bnsenter\b`, CategoryContainerEscape, RiskCritical, ActionBlock, "nsenter can enter host namespaces"},
		{`\bchroot\s+`, CategoryContainerEscape, RiskCritical, ActionBlock, "chroot can pivot to host filesystem"},
		{`/proc/1/(environ|root|cgroup|ns)`, CategoryContainerEscape, RiskCritical, ActionBlock, "Accessing host init process information"},
		{`\b(insmod|modprobe|rmmod)\b`, CategoryContainerEscape, RiskCritical, ActionBlock, "Kernel module operations not permitted"},
		{`\bunshare\b`, CategoryContainerEscape, RiskHigh, ActionBlock, "Namespace manipulation may enable escape"},
		{`\bdocker\s+(run|exec|cp|build|pull|push)`, CategoryContainerEscape, RiskCritical, ActionBlock, "Docker CLI commands from within container"},
		{`\b(kubectl|crictl|ctr)\s+`, CategoryContainerEscape, RiskCritical, ActionBlock, "Container orchestration commands not permitted"},
		{`/dev/(mem|kmem|port)`, CategoryContainerEscape, RiskCritical, ActionBlock, "Direct memory/port access not permitted"},
		{`echo\s+.*>\s*/proc/sys/`, CategoryContainerEscape, RiskCritical, ActionBlock, "Writing to kernel parameters not permitted"},

		// ── Tier 2: Destructive File Operations ──
		{`rm\s+(-[a-zA-Z]*[rf][a-zA-Z]*\s+)*(\/($|\s)|\/etc|\/var|\/usr|\/boot|\/dev|\/sys|\/proc)`, CategoryDestructiveFile, RiskCritical, ActionBlock, "Recursive/forced deletion of system directories"},
		{`\bshred\s+`, CategoryDestructiveFile, RiskCritical, ActionBlock, "Secure file erasure has no pentesting purpose"},
		{`\bdd\s+.*of\s*=\s*/dev/`, CategoryDestructiveFile, RiskCritical, ActionBlock, "Writing to block devices"},
		{`\b(mkfs|wipefs)\b`, CategoryDestructiveFile, RiskCritical, ActionBlock, "Filesystem creation/wiping not permitted"},
		{`\btruncate\s+.*-s\s*0\s+/(etc|var|usr)`, CategoryDestructiveFile, RiskHigh, ActionBlock, "Truncating system files to zero"},

		// ── Tier 3: Network DoS ──
		{`hping3?\s+.*--(flood|fast)`, CategoryNetworkDos, RiskCritical, ActionBlock, "Network flood attacks not permitted"},
		{`\bslowloris\b`, CategoryNetworkDos, RiskCritical, ActionBlock, "Slowloris DoS tool not permitted"},
		{`\bslowhttp(test)?\b`, CategoryNetworkDos, RiskCritical, ActionBlock, "Slow HTTP DoS tool not permitted"},
		{`:\(\)\s*\{.*\|.*&\s*\}`, CategoryNetworkDos, RiskCritical, ActionBlock, "Fork bomb detected"},
		{`\bstress(-ng)?\s+`, CategoryNetworkDos, RiskHigh, ActionBlock, "Stress testing tool not permitted"},
		{`\bping\s+.*-f\b`, CategoryNetworkDos, RiskHigh, ActionBlock, "Flood ping not permitted"},
		{`\biptables\s+.*(DROP|REJECT)`, CategoryNetworkDos, RiskHigh, ActionBlock, "Firewall manipulation not permitted"},
		{`\btc\s+.*netem`, CategoryNetworkDos, RiskHigh, ActionBlock, "Traffic shaping not permitted"},
		{`nmap\s+.*--script\s*=?\s*[^\s]*(dos)`, CategoryNetworkDos, RiskHigh, ActionBlock, "Nmap DoS scripts not permitted"},

		// ── Tier 4: Persistence / Implants ──
		{`\b(useradd|adduser)\s+`, CategoryPersistence, RiskHigh, ActionBlock, "Creating user accounts not permitted"},
		{`\busermod\s+.*-[a-zA-Z]*G`, CategoryPersistence, RiskHigh, ActionBlock, "Modifying user groups not permitted"},
		{`>>\s*.*authorized_keys`, CategoryPersistence, RiskHigh, ActionBlock, "SSH key injection not permitted"},
		{`>>\s*/etc/(passwd|shadow|sudoers)`, CategoryPersistence, RiskCritical, ActionBlock, "System credential file modification not permitted"},
		{`\bcrontab\s+`, CategoryPersistence, RiskHigh, ActionBlock, "Cron job manipulation not permitted"},
		{`/etc/cron\.(d|daily|hourly|weekly|monthly)/`, CategoryPersistence, RiskHigh, ActionBlock, "Cron directory manipulation not permitted"},
		{`/etc/systemd/system/`, CategoryPersistence, RiskHigh, ActionBlock, "Systemd service installation not permitted"},

		// ── Tier 5: Reverse Shells / Exfiltration ──
		{`(bash|sh|zsh)\s+.*>&\s*/dev/tcp/`, CategoryReverseShellExfil, RiskHigh, ActionWarn, "Bash reverse shell pattern detected"},
		{`\b(nc|ncat|netcat)\s+.*-[a-zA-Z]*e\s+`, CategoryReverseShellExfil, RiskHigh, ActionWarn, "Netcat reverse shell pattern detected"},
		{`python[23]?\s+-c\s+.*socket.*connect`, CategoryReverseShellExfil, RiskHigh, ActionWarn, "Python reverse shell pattern detected"},
		{`\bsocat\s+.*EXEC:.*TCP:`, CategoryReverseShellExfil, RiskHigh, ActionWarn, "Socat reverse shell pattern detected"},
		{`php\s+-r\s+.*fsockopen`, CategoryReverseShellExfil, RiskHigh, ActionWarn, "PHP reverse shell pattern detected"},
		{`ruby\s+-e\s+.*TCPSocket`, CategoryReverseShellExfil, RiskHigh, ActionWarn, "Ruby reverse shell pattern detected"},
		{`perl\s+-e\s+.*socket.*connect`, CategoryReverseShellExfil, RiskHigh, ActionWarn, "Perl reverse shell pattern detected"},
		{`curl\s+.*--upload-file`, CategoryReverseShellExfil, RiskMedium, ActionWarn, "File upload via curl detected"},
		{`curl\s+.*-d\s+@`, CategoryReverseShellExfil, RiskMedium, ActionWarn, "File content POST via curl detected"},
		{`wget\s+.*--post-file`, CategoryReverseShellExfil, RiskMedium, ActionWarn, "File upload via wget detected"},
		{`\bscp\s+`, CategoryReverseShellExfil, RiskMedium, ActionWarn, "SCP file transfer detected"},
		{`\brsync\s+`, CategoryReverseShellExfil, RiskMedium, ActionWarn, "Rsync transfer detected"},

		// ── Tier 6: Credential Abuse ──
		{`\b(hydra|medusa|ncrack|patator)\s+`, CategoryCredentialAbuse, RiskHigh, ActionWarn, "Brute force tool detected — verify target is in scope"},
		{`\b(crackmapexec|nxc|netexec)\s+`, CategoryCredentialAbuse, RiskHigh, ActionWarn, "Network credential tool detected — verify target is in scope"},
		{`\bsshpass\s+`, CategoryCredentialAbuse, RiskHigh, ActionWarn, "Automated SSH login detected — verify target is in scope"},

		// ── Tier 7: Aggressive Pentest Flags ──
		{`sqlmap\s+.*--risk\s*=?\s*3`, CategoryAggressiveFlags, RiskHigh, ActionWarn, "sqlmap risk level 3 can modify target data (UPDATE/INSERT)"},
		{`sqlmap\s+.*--(os-shell|os-pwn|os-cmd)`, CategoryAggressiveFlags, RiskHigh, ActionWarn, "sqlmap OS command execution modifies target state"},
		{`sqlmap\s+.*--file-write`, CategoryAggressiveFlags, RiskHigh, ActionWarn, "sqlmap file write modifies target filesystem"},
		{`nmap\s+.*--script\s*=?\s*[^\s]*(exploit|brute)`, CategoryAggressiveFlags, RiskMedium, ActionWarn, "Nmap exploit/brute scripts may modify target state"},
		{`(gobuster|ffuf|feroxbuster|dirsearch)\s+.*-t\s+[0-9]{3,}`, CategoryAggressiveFlags, RiskMedium, ActionWarn, "Very high thread count may overwhelm target"},
		{`\bwpscan\s+.*--passwords`, CategoryAggressiveFlags, RiskMedium, ActionWarn, "WordPress brute force may lock accounts"},
	}

	rules = make([]Rule, 0, len(raw))
	for _, r := range raw {
		compiled := regexp.MustCompile(r.Pattern)
		rules = append(rules, Rule{
			Pattern:  compiled,
			Category: r.Category,
			Risk:     r.Risk,
			Action:   r.Action,
			Reason:   r.Reason,
		})
	}

	// Verify CategoryTier and CategoryInfo have identical key sets.
	for cat := range CategoryTier {
		if _, ok := CategoryInfo[cat]; !ok {
			panic(fmt.Sprintf("deafguard: CategoryTier has %q but CategoryInfo does not", cat))
		}
	}
	for cat := range CategoryInfo {
		if _, ok := CategoryTier[cat]; !ok {
			panic(fmt.Sprintf("deafguard: CategoryInfo has %q but CategoryTier does not", cat))
		}
	}
}
