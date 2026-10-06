package smallmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFacts_ExtractsServiceLinesWithBanner(t *testing.T) {
	out := `Nmap scan report for 10.0.0.5
PORT     STATE SERVICE VERSION
22/tcp   open  ssh     OpenSSH 8.2p1
8443/tcp open  https   Apache Tomcat 9.0.31`

	facts := Facts(out)

	assert.Contains(t, facts, "22/tcp open: ssh (OpenSSH 8.2p1)")
	assert.Contains(t, facts, "8443/tcp open: https (Apache Tomcat 9.0.31)")
	assert.Contains(t, facts, "host seen: 10.0.0.5")
}

func TestFacts_ExtractsCVEsCaseInsensitiveAndNormalizes(t *testing.T) {
	facts := Facts("vulnerable to cve-2020-1938 and CVE-2017-12615")
	assert.Contains(t, facts, "CVE referenced: CVE-2020-1938")
	assert.Contains(t, facts, "CVE referenced: CVE-2017-12615")
}

func TestFacts_IsDeduplicatedAndSorted(t *testing.T) {
	facts := Facts("CVE-2020-1938 CVE-2020-1938 CVE-2020-1938")
	assert.Equal(t, []string{"CVE referenced: CVE-2020-1938"}, facts)
}

func TestFacts_DoesNotDoubleCountPortWithBanner(t *testing.T) {
	facts := Facts("22/tcp open ssh OpenSSH 8.2p1")
	count := 0
	for _, f := range facts {
		if len(f) >= 10 && f[:10] == "22/tcp ope" {
			count++
		}
	}
	assert.Equal(t, 1, count)
}

func TestFacts_EmptyInputYieldsNoFacts(t *testing.T) {
	assert.Empty(t, Facts(""))
}
