package smallmodel

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	reCVE  = regexp.MustCompile(`(?i)CVE-\d{4}-\d{4,7}`)
	rePort = regexp.MustCompile(`\b(\d{1,5})/(tcp|udp)\s+open\b`)
	reIPv4 = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4]\d|1?\d?\d)\.){3}(?:25[0-5]|2[0-4]\d|1?\d?\d)\b`)
	// reService matches an nmap-style line "80/tcp open http Apache httpd 2.4.41".
	reService = regexp.MustCompile(`(?m)^(\d{1,5})/(tcp|udp)\s+open\s+(\S+)(?:\s+(.*))?$`)
)

// Facts extracts structured, high-value facts from free text (typically tool
// output) using regexes only — no model call, no cost, no drift. The result is
// a stable, de-duplicated, sorted list suitable for feeding Store.MergeFacts.
func Facts(text string) []string {
	set := map[string]struct{}{}

	for _, m := range reService.FindAllStringSubmatch(text, -1) {
		port, proto, svc := m[1], m[2], m[3]
		banner := strings.TrimSpace(m[4])
		fact := fmt.Sprintf("%s/%s open: %s", port, proto, svc)
		if banner != "" {
			fact += " (" + collapseSpaces(banner) + ")"
		}
		set[fact] = struct{}{}
	}

	// Ports seen in the terse form, not already captured with a service banner.
	for _, m := range rePort.FindAllStringSubmatch(text, -1) {
		fact := fmt.Sprintf("%s/%s open", m[1], m[2])
		if !hasPortFact(set, m[1], m[2]) {
			set[fact] = struct{}{}
		}
	}

	for _, cve := range reCVE.FindAllString(text, -1) {
		set["CVE referenced: "+strings.ToUpper(cve)] = struct{}{}
	}

	for _, ip := range reIPv4.FindAllString(text, -1) {
		set["host seen: "+ip] = struct{}{}
	}

	return toSortedSlice(set)
}

func hasPortFact(set map[string]struct{}, port, proto string) bool {
	prefix := fmt.Sprintf("%s/%s open", port, proto)
	for k := range set {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func toSortedSlice(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
