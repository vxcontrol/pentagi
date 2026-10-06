package smallmodel

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const nmapXML = `<?xml version="1.0"?>
<nmaprun scanner="nmap">
  <host>
    <address addr="10.0.0.5" addrtype="ipv4"/>
    <ports>
      <port protocol="tcp" portid="22">
        <state state="open"/>
        <service name="ssh" product="OpenSSH" version="8.2p1"/>
      </port>
      <port protocol="tcp" portid="8443">
        <state state="open"/>
        <service name="https" product="Apache Tomcat" version="9.0.31"/>
      </port>
      <port protocol="tcp" portid="23">
        <state state="closed"/>
      </port>
    </ports>
  </host>
</nmaprun>`

func TestCompactOutput_ParsesNmapXMLIntoSummaryAndFacts(t *testing.T) {
	c := CompactOutput("terminal", nmapXML, 0)

	assert.Contains(t, c.Summary, "host 10.0.0.5: 2 open ports")
	assert.Contains(t, c.Summary, "22/tcp open: ssh OpenSSH 8.2p1")
	assert.NotContains(t, c.Summary, "23/tcp", "closed ports are dropped")
	assert.Contains(t, c.Facts, "host seen: 10.0.0.5")
	assert.True(t, c.Truncated)
}

func TestCompactOutput_KeepsHeadAndTailWhenOverBudget(t *testing.T) {
	raw := "HEAD-" + strings.Repeat("x", 5000) + "-TAIL"
	c := CompactOutput("terminal", raw, 400)

	assert.True(t, c.Truncated)
	assert.LessOrEqual(t, len(c.Summary), 400)
	assert.True(t, strings.HasPrefix(c.Summary, "HEAD-"))
	assert.True(t, strings.HasSuffix(c.Summary, "-TAIL"))
	assert.Contains(t, c.Summary, "bytes omitted")
}

func TestCompactOutput_PassesThroughWhenUnderBudget(t *testing.T) {
	raw := "short output"
	c := CompactOutput("terminal", raw, 400)

	assert.False(t, c.Truncated)
	assert.Equal(t, raw, c.Summary)
}

func TestCompactOutput_FallsBackWhenNmapXMLIsMalformed(t *testing.T) {
	c := CompactOutput("terminal", "<nmaprun> not really xml", 0)
	assert.Equal(t, "<nmaprun> not really xml", c.Summary)
}
