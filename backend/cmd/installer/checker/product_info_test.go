package checker

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/vxcontrol/cloud/models"
)

// frame writes one chunk the way the daemon multiplexes a non-TTY exec: stream, three zero bytes, big-endian length.
func frame(buffer *bytes.Buffer, stream stdcopy.StdType, payload string) {
	var header [8]byte
	header[0] = byte(stream)
	binary.BigEndian.PutUint32(header[4:], uint32(len(payload)))
	buffer.Write(header[:])
	buffer.WriteString(payload)
}

// productInfoExecStream answers an exec start as the daemon does: upgrade the connection, write the raw stream.
func productInfoExecStream(t *testing.T, stream []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		conn, _, err := http.NewResponseController(w).Hijack()
		if err != nil {
			t.Errorf("hijack: %v", err)
			return
		}
		defer conn.Close()
		upgrade := "HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.multiplexed-stream\r\n" +
			"Connection: Upgrade\r\nUpgrade: tcp\r\n\r\n"
		if _, err := conn.Write(append([]byte(upgrade), stream...)); err != nil {
			t.Errorf("write the exec stream: %v", err)
			return
		}
		// A half-close, then waiting for the client to hang up, lets it read to EOF instead of a reset.
		if tcp, ok := conn.(*net.TCPConn); ok {
			_ = tcp.CloseWrite()
		}
		_, _ = io.Copy(io.Discard, conn)
	}
}

func TestProductInfo_GatherProductInfo_DemultiplexesTheExecStream(t *testing.T) {
	const document = `{"schema":1,"version":"0.9.3"}`
	var stream bytes.Buffer
	frame(&stream, stdcopy.Stderr, "cannot reach the database\n")
	frame(&stream, stdcopy.Stdout, document)
	if json.Valid(stream.Bytes()) {
		t.Fatal("the framed stream parses as JSON, so it cannot tell a demultiplexed read from a raw one")
	}
	tests := []struct {
		name     string
		running  bool
		exitCode int
		want     string
	}{
		{"a running product's answer arrives without its stderr", true, 0, document},
		{"a product that exits with an error gave no answer", true, 1, ""},
		{"a stopped product is not asked", false, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docker := checkerFakeDocker(t, map[string]http.HandlerFunc{
				"GET /containers/pentagi/json":  checkerJSON(http.StatusOK, fmt.Sprintf(`{"State":{"Running":%t}}`, tt.running)),
				"POST /containers/pentagi/exec": checkerJSON(http.StatusCreated, `{"Id":"info"}`),
				"POST /exec/info/start":         productInfoExecStream(t, stream.Bytes()),
				"GET /exec/info/json":           checkerJSON(http.StatusOK, fmt.Sprintf(`{"ExitCode":%d}`, tt.exitCode)),
			})
			handler := &defaultCheckHandler{mx: &sync.Mutex{}, dockerClient: docker}

			if got := string(handler.gatherProductInfo(t.Context())); got != tt.want {
				t.Errorf("product info = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProductInfo_ProductInfoFromOutput_CarriesOnlyADocumentTheContractAccepts(t *testing.T) {
	future := `{"schema":7,"version":"9.9.9","engine":{"scenarios":12},"unheard_of":[1,2,3]}`
	atTheBound := `{"p":"` + strings.Repeat("x", models.MaxProductInfoBytes-8) + `"}`
	tests := []struct {
		name, output, want string
	}{
		{"a document from a newer product, byte for byte", "  " + future + "\n", future},
		{"a document exactly at the contract's bound", atTheBound, atTheBound},
		{"an older build printing usage", "Usage of /opt/pentagi/bin/pentagi:\n  -info\n", ""},
		{"nothing at all", "", ""},
		{"whitespace only", "   \n\t ", ""},
		{"a truncated document", `{"schema":1,"version":`, ""},
		{"a log line before the answer", "level=info msg=starting\n" + `{"schema":1}`, ""},
		{"one byte over the contract's bound", `{"p":"` + strings.Repeat("x", models.MaxProductInfoBytes-7) + `"}`, ""},
	}

	for _, tt := range tests {
		got := productInfoFromOutput([]byte(tt.output))
		if string(got) != tt.want || (got == nil) != (tt.want == "") {
			t.Errorf("%s: carried %d bytes, want %d", tt.name, len(got), len(tt.want))
		}
	}
}
