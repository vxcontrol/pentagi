package checker

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"strings"
	"testing"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/vxcontrol/cloud/models"
)

// frame writes one chunk the way the docker daemon does on a non-TTY exec: an
// eight-byte header — the stream number, three zero bytes, then the payload
// length big-endian — followed by the payload.
//
// Written out here rather than taken from a library because the reading half is
// all the client package exposes; the daemon owns the writing half. Spelling the
// format out is also the point of these tests: it is the thing that silently
// corrupts the answer when nobody accounts for it.
func frame(buffer *bytes.Buffer, stream stdcopy.StdType, payload string) {
	var header [8]byte
	header[0] = byte(stream)
	binary.BigEndian.PutUint32(header[4:], uint32(len(payload)))
	buffer.Write(header[:])
	buffer.WriteString(payload)
}

// TestTheExecStreamMustBeDemultiplexedBeforeItIsJSON.
//
// The trap this pins is not hypothetical and it is not visible by reading: the
// daemon frames every chunk of a non-TTY exec with an eight-byte header naming
// the stream and the length. Read the socket directly and the first thing before
// the opening brace is binary, so a perfectly good document does not parse — and
// the symptom is "the product returned nothing", which is also what a product
// without the flag returns.
//
// The test builds the stream the way the daemon does and asserts both halves:
// raw is not JSON, demultiplexed is.
func TestTheExecStreamMustBeDemultiplexedBeforeItIsJSON(t *testing.T) {
	document := `{"schema":1,"version":"0.9.3"}`

	var framed bytes.Buffer
	frame(&framed, stdcopy.Stdout, document)

	if json.Valid(framed.Bytes()) {
		t.Fatal("the framed stream parsed as JSON, so this test proves nothing — " +
			"the framing that makes demultiplexing necessary is not being produced")
	}

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, &framed); err != nil {
		t.Fatalf("demultiplex: %v", err)
	}

	info := productInfoFromOutput(stdout.Bytes())
	if info == nil {
		t.Fatal("a demultiplexed document was rejected")
	}
	if string(info) != document {
		t.Errorf("the document was altered on the way through:\n got %s\nwant %s", info, document)
	}
}

// TestStderrNeverReachesTheDocument.
//
// stdout carries the answer and stderr carries diagnostics, and the two arrive
// interleaved on one socket. Mixing them would put a warning inside the document
// — or, worse, produce something that still parses and is no longer what the
// product said.
func TestStderrNeverReachesTheDocument(t *testing.T) {
	var framed bytes.Buffer
	frame(&framed, stdcopy.Stderr, "cannot reach the database\n")
	frame(&framed, stdcopy.Stdout, `{"schema":1}`)

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, &framed); err != nil {
		t.Fatalf("demultiplex: %v", err)
	}

	if got := string(productInfoFromOutput(stdout.Bytes())); got != `{"schema":1}` {
		t.Errorf("the document picked up something from stderr: %q", got)
	}
	if !strings.Contains(stderr.String(), "cannot reach the database") {
		t.Error("the diagnostic went somewhere other than stderr")
	}
}

// TestNothingThatIsNotADocumentIsCarried.
//
// Every one of these is a state a real installation is in on the day this ships:
// an older product that has never heard of the flag, one whose answer was cut
// short, one that answered with more than the contract will accept. Carrying any
// of them would build a request the service refuses whole — and refusing it costs
// the answer for every component, not just this field.
func TestNothingThatIsNotADocumentIsCarried(t *testing.T) {
	for name, output := range map[string]string{
		"an older build printing usage": "Usage of /opt/pentagi/bin/pentagi:\n  -info\n",
		"nothing at all":                "",
		"whitespace only":               "   \n\t ",
		"a truncated document":          `{"schema":1,"version":`,
		"a log line before the answer":  "level=info msg=starting\n" + `{"schema":1}`,
		"larger than the contract takes": `{"schema":1,"padding":"` +
			strings.Repeat("x", maxProductInfoBytes) + `"}`,
	} {
		if info := productInfoFromOutput([]byte(output)); info != nil {
			t.Errorf("%s was carried through as a document: %q", name, info)
		}
	}
}

// TestTheDocumentBoundIsTheContractsOwn.
//
// The bound here has one job: never build a request the service will refuse whole.
// That makes it the CONTRACT's number, and asserting it as a literal is what let it
// be wrong — it stood at 64 KiB against a 16 KiB contract, and every other test in
// this file measures against the same constant, so they all agreed with each other
// and none of them agreed with the server.
//
// Both halves are pinned: that the constant IS the contract's, and that a document
// one byte over the contract is refused. The second would still fail if somebody
// reintroduced a local number that happened to be larger.
func TestTheDocumentBoundIsTheContractsOwn(t *testing.T) {
	if maxProductInfoBytes != models.MaxProductInfoBytes {
		t.Errorf("the installer accepts %d bytes, the contract accepts %d — a document in "+
			"between costs the whole update check", maxProductInfoBytes, models.MaxProductInfoBytes)
	}

	padding := models.MaxProductInfoBytes // the envelope around it pushes this over
	oversized := `{"schema":1,"padding":"` + strings.Repeat("x", padding) + `"}`
	if len(oversized) <= models.MaxProductInfoBytes {
		t.Fatalf("the fixture is not oversized: %d bytes (test bug)", len(oversized))
	}
	if info := productInfoFromOutput([]byte(oversized)); info != nil {
		t.Errorf("a document of %d bytes was carried into a request the contract bounds at %d",
			len(oversized), models.MaxProductInfoBytes)
	}

	// And one that fits is still carried: a bound that refuses everything would
	// pass the assertion above and quietly drop the document on every check.
	fitting := `{"schema":1,"version":"0.9.3"}`
	if info := productInfoFromOutput([]byte(fitting)); info == nil {
		t.Error("a document well inside the bound was dropped")
	}
}

// TestADocumentIsCarriedWithoutBeingUnderstood.
//
// The installer and the product ship separately, so anything the installer knew
// about the shape would be a second place to change every time the product learns
// to describe more of itself. This pins that a document full of fields this build
// has never heard of survives byte for byte.
func TestADocumentIsCarriedWithoutBeingUnderstood(t *testing.T) {
	future := `{"schema":7,"version":"9.9.9","engine":{"scenarios":12},"unheard_of":[1,2,3]}`

	info := productInfoFromOutput([]byte("  " + future + "\n"))
	if info == nil {
		t.Fatal("a document from a newer product was dropped")
	}
	if string(info) != future {
		t.Errorf("the document was reshaped on the way through:\n got %s\nwant %s", info, future)
	}
}
