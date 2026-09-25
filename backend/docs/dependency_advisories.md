# Dependency advisories

`govulncheck ./...` is the source of truth here. Run it from `backend/`:

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

It reports two classes separately: vulnerabilities your code **calls**, and ones
that merely sit in the module graph. Only the first class is worth acting on, and
even there the call graph is deliberately conservative — reaching any symbol of a
module is enough to attribute every advisory filed against that module, including
advisories in code the module ships but we never execute.

Everything with a published fix is taken. Fifteen findings stay, in two classes.

## Eight in a server we only talk to

`github.com/ollama/ollama` — GO-2025-4251, 3824, 3695, 3689, 3582, 3559, 3558 and
3557, none with a fixed version: missing authentication on model management,
cross-domain token exposure, denial of service via null dereference, divide by
zero, out-of-bounds read, unthrottled allocation. Every one of them describes the
**Ollama server**. We import `github.com/ollama/ollama/api` — its HTTP client —
and nothing else, and we never run `ollama serve`. The traces govulncheck prints
say as much: `api.AuthorizationError.Error`, i.e. formatting an error we
received, and `api.Client.Chat`, i.e. making a request. The vulnerable code lives
on whatever host serves the endpoint an operator configures.

## Seven the `go` line reports and the binary does not carry

GO-2026-6218, 6091, 6090, 6089, 6088, 5972 and 5026 are filed against `net/url`,
`html/template`, `crypto/tls`, `net/http`, `encoding/xml` and `encoding/asn1`,
all fixed in go1.26.6. govulncheck attributes standard-library advisories to the
version on the `go` line of `go.mod`, which states the floor the module supports
rather than the toolchain that compiles it. The image builds on `golang:1.26`;
`go version` run against the binary inside the built image prints a patch past
those fixes. Raising the `go` line silences all seven, and it also raises the
floor for everyone building from source, so it is a release decision and not a
security fix.

## What would change this

An advisory here gaining a `Fixed in` version, our code starting to run Ollama's
server side in-process, or the `go` line moving. Re-run the command above after
any dependency work and compare the module list, not just the count: a new module
appearing is the signal.
