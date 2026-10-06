package smallmodel

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// Compacted is the result of pre-processing a raw tool output: a short form fit
// for a small model's context, the facts extracted deterministically, and
// whether the raw output was shortened (so the caller can offer a tool to
// re-read the full version from disk).
type Compacted struct {
	Summary   string
	Facts     []string
	Truncated bool
}

// CompactOutput turns raw tool output into a form a small model reads well. It
// prefers a dedicated parser (nmap XML today); otherwise it keeps the head and
// tail of the output, because the useful part of a log is often at the end and
// a plain byte-truncation keeps only the start. maxBytes <= 0 disables shortening.
func CompactOutput(toolName, raw string, maxBytes int) Compacted {
	if summary, facts, ok := summarizeNmapXML(raw); ok {
		return Compacted{Summary: summary, Facts: facts, Truncated: len(raw) > len(summary)}
	}

	facts := Facts(raw)
	if maxBytes <= 0 || len(raw) <= maxBytes {
		return Compacted{Summary: raw, Facts: facts, Truncated: false}
	}

	return Compacted{Summary: headTail(raw, maxBytes), Facts: facts, Truncated: true}
}

// headTail keeps the first and last portions of s within budget, joined by a
// marker that names how many bytes were dropped.
func headTail(s string, budget int) string {
	const marker = "\n... [%d bytes omitted] ...\n"
	if budget <= len(marker)+16 {
		return s[:budget]
	}
	room := budget - len(marker)
	head := room * 6 / 10
	tail := room - head
	omitted := len(s) - head - tail
	return s[:head] + fmt.Sprintf(marker, omitted) + s[len(s)-tail:]
}

type nmapRun struct {
	XMLName xml.Name   `xml:"nmaprun"`
	Hosts   []nmapHost `xml:"host"`
}

type nmapHost struct {
	Addresses []struct {
		Addr string `xml:"addr,attr"`
	} `xml:"address"`
	Ports struct {
		Ports []struct {
			PortID   string `xml:"portid,attr"`
			Protocol string `xml:"protocol,attr"`
			State    struct {
				State string `xml:"state,attr"`
			} `xml:"state"`
			Service struct {
				Name    string `xml:"name,attr"`
				Product string `xml:"product,attr"`
				Version string `xml:"version,attr"`
			} `xml:"service"`
		} `xml:"port"`
	} `xml:"ports"`
}

// summarizeNmapXML parses nmap -oX output into a few lines plus structured facts.
// ok is false when the input is not nmap XML, so the caller falls back.
func summarizeNmapXML(raw string) (summary string, facts []string, ok bool) {
	if !strings.Contains(raw, "<nmaprun") {
		return "", nil, false
	}
	var run nmapRun
	if err := xml.Unmarshal([]byte(raw), &run); err != nil {
		return "", nil, false
	}

	var b strings.Builder
	factSet := map[string]struct{}{}
	for _, h := range run.Hosts {
		addr := ""
		if len(h.Addresses) > 0 {
			addr = h.Addresses[0].Addr
		}
		open := 0
		var lines []string
		for _, p := range h.Ports.Ports {
			if p.State.State != "open" {
				continue
			}
			open++
			svc := strings.TrimSpace(strings.Join([]string{p.Service.Name, p.Service.Product, p.Service.Version}, " "))
			line := fmt.Sprintf("%s/%s open: %s", p.PortID, p.Protocol, svc)
			lines = append(lines, line)
			factSet[line] = struct{}{}
		}
		if addr != "" {
			factSet["host seen: "+addr] = struct{}{}
		}
		fmt.Fprintf(&b, "host %s: %d open ports\n", addr, open)
		for _, l := range lines {
			fmt.Fprintf(&b, "  - %s\n", l)
		}
	}

	return strings.TrimRight(b.String(), "\n"), toSortedSlice(factSet), true
}
