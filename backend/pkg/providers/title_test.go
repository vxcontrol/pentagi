package providers

import (
	"context"
	"database/sql"
	"sync/atomic"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/mock"
	"pentagi/pkg/templates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTitle_NormalizeTitle_KeepsOnlyThePlainTitleText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		answer string
		title  string
	}{
		{"bold around the whole title", "**Check Kernel & User**", "Check Kernel & User"},
		{"leading blank lines", "\n\nEnumerate SMB Shares", "Enumerate SMB Shares"},
		{"italic around the whole title", "*Scan Open Ports*", "Scan Open Ports"},
		{"bold and italic together", "***Scan Open Ports***", "Scan Open Ports"},
		{"bold inside the title", "Check **Kernel** Version", "Check Kernel Version"},
		{"bold label before the title", "**Title:** Check Kernel", "Title: Check Kernel"},
		{"line breaks inside the title", "Check Kernel\r\n\r\nand User", "Check Kernel and User"},
		{"spaces around and inside", "  Check   Kernel\t", "Check Kernel"},
		{"plain title is kept", "Check Kernel & User", "Check Kernel & User"},
		{"wildcard host is kept", "Enumerate *.example.com", "Enumerate *.example.com"},
		{"lone asterisk in a query is kept", "Test SELECT * FROM users", "Test SELECT * FROM users"},
		{"asterisks around spaces are not emphasis", "Rate ** fast ** scan", "Rate ** fast ** scan"},
		{"underscores in identifiers are kept", "Test __proto__ pollution via user_id", "Test __proto__ pollution via user_id"},
		{"blank answer stays blank", " \n\t ", ""},
		{"wildcard host inside a bold wrapper", "**Audit *.example.com subdomains for takeover**", "Audit *.example.com subdomains for takeover"},
		{"ip range inside a bold wrapper", "**Audit 192.168.*.* network for SMB hosts**", "Audit 192.168.*.* network for SMB hosts"},
		{"ip range after leading line breaks", "\n\nScan 192.168.*.* network for SMB hosts", "Scan 192.168.*.* network for SMB hosts"},
		{"mask abutting the closing bold", "**Scan RDP ports 10.0.*.***", "Scan RDP ports 10.0.*.*"},
		{"path glob inside a bold wrapper", "**Audit /api/v1/users/*/orders/* for IDOR**", "Audit /api/v1/users/*/orders/* for IDOR"},
		{"path globs are kept", "Collect SSH keys from /home/*/.ssh/* on 10.0.0.5", "Collect SSH keys from /home/*/.ssh/* on 10.0.0.5"},
		{"ldap substring filter is kept", "Search LDAP for (cn=*admin*) accounts", "Search LDAP for (cn=*admin*) accounts"},
		{"star after a space is not an opener", "Grep * for passwd in /etc/*", "Grep * for passwd in /etc/*"},
		{"cron schedule is kept", "Inspect */5 * * * * cron entry", "Inspect */5 * * * * cron entry"},
		{"two emphasized words", "Scan **Ports** and **Services**", "Scan Ports and Services"},
		{"space before a closing star is not emphasis", "Rate **fast ** scan", "Rate **fast ** scan"},
		{"stars glued between words are kept", "Rebind a*b to c*d", "Rebind a*b to c*d"},
		{"bold label and bold host", "**Title:** Enumerate SMB on **dc01**", "Title: Enumerate SMB on dc01"},
		{"bold at both ends", "**Nmap** scan of **10.0.0.5**", "Nmap scan of 10.0.0.5"},
		{"italic at both ends", "*Recon* of *example.com*", "Recon of example.com"},
		{"bold at both ends before a comma", "**Scan**, then **Exploit**", "Scan, then Exploit"},
		{"italic word and a trailing mask", "*Enumerate* hosts in 10.0.0.*", "Enumerate hosts in 10.0.0.*"},
		{"bold and italic at both ends", "***Scan*** and ***Exploit***", "Scan and Exploit"},
		{"ip range inside an italic wrapper", "*Audit 192.168.*.* network for SMB hosts*", "Audit 192.168.*.* network for SMB hosts"},
		{"mask abutting an inline bold closer", "Scan **10.0.0.*** subnet", "Scan 10.0.0.* subnet"},
		{"mask abutting a bold closer before a word", "Enumerate SMB on **192.168.1.*** hosts", "Enumerate SMB on 192.168.1.* hosts"},
		{"two-star mask inside inline bold", "Scan hosts in **10.0.*.*** range", "Scan hosts in 10.0.*.* range"},
		{"mask abutting an inline italic closer", "Scan *10.0.0.** subnet", "Scan 10.0.0.* subnet"},
		{"bold wrapper around a leading path", "**/etc/passwd Audit**", "/etc/passwd Audit"},
		{"bold wrapper around a leading dotfile", "**.git Exposure Check**", ".git Exposure Check"},
		{"bold wrapper around a quoted title", "**\"Scan Ports\"**", "\"Scan Ports\""},
		{"bold wrapper around a leading cve tag", "**[CVE-2024-1234] Exploit**", "[CVE-2024-1234] Exploit"},
		{"bold and italic wrapper around a leading path", "***/admin Panel Bruteforce***", "/admin Panel Bruteforce"},
		{"end of sequence token after the title", "Scan Open Ports<|eos|>", "Scan Open Ports"},
		{"end of turn token between words", "Scan<|eot_id|>Open Ports", "Scan Open Ports"},
		{"control token after a bold wrapper", "**Scan Open Ports**<|im_end|>", "Scan Open Ports"},
		{"pipes without angle brackets are kept", "Grep | sort | uniq on /var/log", "Grep | sort | uniq on /var/log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.title, normalizeTitle(tt.answer))
		})
	}
}

type noUserProvidersQuerier struct {
	database.Querier
}

func (noUserProvidersQuerier) GetUserProviderByName(
	context.Context, database.GetUserProviderByNameParams,
) (database.Provider, error) {
	return database.Provider{}, sql.ErrNoRows
}

type defaultImageDocker struct {
	docker.DockerClient
}

func (defaultImageDocker) GetDefaultImage() string { return "debian:latest" }

func TestTitle_NormalizeTitle_IsAppliedToEveryGeneratedTitle(t *testing.T) {
	t.Parallel()

	const input = "check the kernel version and the current user"

	// Each class in this answer is pinned on its own by the NormalizeTitle table.
	const (
		answer = "\n\n**Check Kernel & User**<|eos|>"
		title  = "Check Kernel & User"
	)

	sites := []struct {
		name   string
		prompt string
		title  func(t *testing.T, pc *providerController, prv provider.Provider) string
	}{
		{"the flow title", "Flow Title Generator", func(t *testing.T, pc *providerController, prv provider.Provider) string {
			fp, err := pc.NewFlowProvider(t.Context(), prv.Name(), templates.NewDefaultPrompter(), nil, 1, 1, false, input)
			require.NoError(t, err)

			return fp.Title()
		}},
		{"the assistant title", "Flow Title Generator", func(t *testing.T, pc *providerController, prv provider.Provider) string {
			ap, err := pc.NewAssistantProvider(
				t.Context(), prv.Name(), templates.NewDefaultPrompter(), nil, 1, 1, 1, "debian:latest", input, nil,
			)
			require.NoError(t, err)

			return ap.Title()
		}},
		{"the task title", "Task Title Generator", func(t *testing.T, _ *providerController, prv provider.Provider) string {
			fp := newFlowProvider()
			fp.Provider = prv
			fp.prompter = templates.NewDefaultPrompter()

			got, err := fp.GetTaskTitle(t.Context(), input)
			require.NoError(t, err)

			return got
		}},
	}

	for _, site := range sites {
		t.Run(site.name, func(t *testing.T) {
			t.Parallel()

			prv := mock.NewProvider(provider.ProviderMistral, "mistral", "ministral-8b")
			prv.SetDefaultResponse("English")
			prv.SetResponses([]mock.ResponseConfig{{Key: site.prompt, Response: answer}})

			pc := &providerController{
				cfg:             &config.Config{DockerImageSelectionMode: ImageSelectionModeFixed},
				db:              noUserProvidersQuerier{},
				docker:          defaultImageDocker{},
				startCallNumber: &atomic.Int64{},
				Providers:       provider.Providers{prv.Name(): prv},
			}

			assert.Equal(t, title, site.title(t, pc, prv))
		})
	}
}
