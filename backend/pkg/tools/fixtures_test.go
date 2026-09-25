// Doubles shared across test files; a double for a write refuses a done context, as the real write does.

package tools

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/graph/model"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/providers/embeddings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/vxcontrol/langchaingo/schema"
	"github.com/vxcontrol/langchaingo/vectorstores/pgvector"
)

// --- Arguments and anonymization ---

func mustJSON(v map[string]any) json.RawMessage {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return raw
}

// identityReplacer is an anonymizer that changes nothing.
type identityReplacer struct{}

func (identityReplacer) ReplaceString(s string) string    { return s }
func (identityReplacer) ReplaceBytes(b []byte) []byte     { return b }
func (identityReplacer) WrapReader(r io.Reader) io.Reader { return r }

const secretHost = "10.13.37.5"

// hostReplacer is an anonymizer that replaces secretHost with <REDACTED>.
type hostReplacer struct{}

func (hostReplacer) ReplaceString(s string) string {
	return strings.ReplaceAll(s, secretHost, "<REDACTED>")
}
func (hostReplacer) ReplaceBytes(b []byte) []byte {
	return []byte(hostReplacer{}.ReplaceString(string(b)))
}
func (hostReplacer) WrapReader(r io.Reader) io.Reader { return r }

// --- Operator log ---

// captureLogrus swaps the standard logger's hooks for a test hook until the test ends.
func captureLogrus(t *testing.T) *logtest.Hook {
	t.Helper()

	hook := new(logtest.Hook)
	previous := logrus.StandardLogger().ReplaceHooks(logrus.LevelHooks{})
	logrus.AddHook(hook)
	t.Cleanup(func() { logrus.StandardLogger().ReplaceHooks(previous) })

	return hook
}

// --- Docker daemon ---

// dockerRemoval is one container fakeDockerClient was asked to remove.
type dockerRemoval struct {
	localID string
	dbID    int64
}

// dockerLaunch is one container fakeDockerClient was asked to run.
type dockerLaunch struct {
	name       string
	kind       database.ContainerType
	flowID     int64
	config     container.Config
	hostConfig container.HostConfig
}

// fakeDockerClient stands in for the Docker daemon: it answers as scripted and records every call.
type fakeDockerClient struct {
	// IsContainerRunning answers probeErr, refuses an empty id as the daemon does, then running, else isRunning.
	isRunning bool
	running   func(probe int) (bool, error)
	probeErr  error

	execCreateResp client.ExecCreateResult
	execCreateErr  error
	attachErr      error
	attachOutput   []byte
	// attachDelay holds the attach back, so a command runs for that long before it prints.
	attachDelay time.Duration
	// attachHold keeps the stream open after attachOutput, so a read ends at the deadline, not at EOF.
	attachHold  time.Duration
	inspectResp client.ExecInspectResult
	inspectErr  error

	// CopyFromContainer answers copyFromErr, else archive, else readFileContent as a one-file archive.
	readFileContent string
	archive         []byte
	stat            container.PathStat
	copyFromErr     error
	copyToErr       error

	removeErr error
	launchErr error

	probed         []string // the Docker ids the running probe was asked about
	execContainer  string
	execCreated    client.ExecCreateOptions
	ctxWasCanceled bool // the attach saw its context end
	copyFromCalled bool
	copies         int
	copiedInto     string
	copiedTo       string
	copied         map[string]string // regular files of the last archive copied in, by tar path
	writtenContent string            // the first regular file of the last archive copied in
	removed        []dockerRemoval
	launches       []dockerLaunch
}

var _ docker.DockerClient = (*fakeDockerClient)(nil)

func (d *fakeDockerClient) RunContainer(
	ctx context.Context, name string, kind database.ContainerType, flowID int64,
	cfg *container.Config, hostCfg *container.HostConfig,
) (database.Container, error) {
	if err := ctx.Err(); err != nil {
		return database.Container{}, err
	}
	d.launches = append(d.launches, dockerLaunch{name: name, kind: kind, flowID: flowID, config: *cfg, hostConfig: *hostCfg})
	if d.launchErr != nil {
		return database.Container{}, d.launchErr
	}
	return database.Container{
		ID:      8,
		Status:  database.ContainerStatusRunning,
		LocalID: sql.NullString{String: "fresh", Valid: true},
	}, nil
}

func (d *fakeDockerClient) StopContainer(ctx context.Context, _ string, _ int64) error {
	return ctx.Err()
}

func (d *fakeDockerClient) RemoveContainer(ctx context.Context, localID string, dbID int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	d.removed = append(d.removed, dockerRemoval{localID: localID, dbID: dbID})
	return d.removeErr
}

func (d *fakeDockerClient) IsContainerRunning(_ context.Context, localID string) (bool, error) {
	d.probed = append(d.probed, localID)
	if d.probeErr != nil {
		return false, d.probeErr
	}
	if localID == "" {
		return false, errors.New("invalid container name or ID: value is empty")
	}
	if d.running != nil {
		return d.running(len(d.probed))
	}
	return d.isRunning, nil
}

func (d *fakeDockerClient) KillFlowCommands(ctx context.Context, _ string) error {
	return ctx.Err()
}

func (d *fakeDockerClient) ContainerExecCreate(
	ctx context.Context, name string, options client.ExecCreateOptions,
) (client.ExecCreateResult, error) {
	if err := ctx.Err(); err != nil {
		return client.ExecCreateResult{}, err
	}
	d.execContainer, d.execCreated = name, options
	if d.execCreateErr != nil {
		return client.ExecCreateResult{}, d.execCreateErr
	}
	return d.execCreateResp, nil
}

func (d *fakeDockerClient) ContainerExecAttach(
	ctx context.Context, _ string, _ client.ExecAttachOptions,
) (client.HijackedResponse, error) {
	if d.attachErr != nil {
		return client.HijackedResponse{}, d.attachErr
	}
	if d.attachDelay > 0 {
		select {
		case <-time.After(d.attachDelay):
		case <-ctx.Done():
			d.ctxWasCanceled = true
			return client.HijackedResponse{}, ctx.Err()
		}
	}
	if err := ctx.Err(); err != nil {
		d.ctxWasCanceled = true
		return client.HijackedResponse{}, err
	}

	pr, pw := net.Pipe()
	output, hold := d.attachOutput, d.attachHold
	go func() {
		_, _ = pw.Write(output)
		if hold > 0 {
			time.Sleep(hold)
		}
		pw.Close()
	}()

	return client.HijackedResponse{Conn: pr, Reader: bufio.NewReader(pr)}, nil
}

func (d *fakeDockerClient) ContainerExecInspect(context.Context, string) (client.ExecInspectResult, error) {
	if d.inspectErr != nil {
		return client.ExecInspectResult{}, d.inspectErr
	}
	return d.inspectResp, nil
}

func (d *fakeDockerClient) ContainerStatPath(context.Context, string, string) (container.PathStat, error) {
	return container.PathStat{}, nil
}

func (d *fakeDockerClient) ListContainerDir(context.Context, string, string) (docker.ContainerDirListing, error) {
	return docker.ContainerDirListing{}, nil
}

func (d *fakeDockerClient) CopyToContainer(
	ctx context.Context, name, dstPath string, src io.Reader, _ client.CopyToContainerOptions,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	d.copies++
	d.copiedInto, d.copiedTo = name, dstPath
	if d.copyToErr != nil {
		return d.copyToErr
	}

	d.copied, d.writtenContent = map[string]string{}, ""
	archive := tar.NewReader(src)
	for {
		hdr, err := archive.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		body, err := io.ReadAll(archive)
		if err != nil {
			return err
		}
		if len(d.copied) == 0 {
			d.writtenContent = string(body)
		}
		d.copied[hdr.Name] = string(body)
	}
}

func (d *fakeDockerClient) CopyFromContainer(context.Context, string, string) (io.ReadCloser, container.PathStat, error) {
	d.copyFromCalled = true
	if d.copyFromErr != nil {
		return nil, container.PathStat{}, d.copyFromErr
	}

	archive := d.archive
	if archive == nil && d.readFileContent != "" {
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)
		_ = tw.WriteHeader(&tar.Header{Name: "file", Mode: 0o600, Size: int64(len(d.readFileContent))})
		_, _ = tw.Write([]byte(d.readFileContent))
		_ = tw.Close()
		archive = buf.Bytes()
	}

	return io.NopCloser(&chunkedReader{src: bytes.NewReader(archive)}), d.stat, nil
}

func (d *fakeDockerClient) Cleanup(context.Context) error { return nil }
func (d *fakeDockerClient) GetDefaultImage() string       { return "test-image" }

// dockerReadChunk is the most one Read of a container archive returns, as a socket read does.
const dockerReadChunk = 4096

// chunkedReader reads like a socket, so a read loop that ignores its byte count is caught.
type chunkedReader struct {
	src io.Reader
}

func (r *chunkedReader) Read(p []byte) (int, error) {
	if len(p) > dockerReadChunk {
		p = p[:dockerReadChunk]
	}
	return r.src.Read(p)
}

// --- Containers table ---

// fakeContainerDB stands in for the containers table: the flow's primary row and its status updates.
type fakeContainerDB struct {
	database.Querier

	row       *database.Container // nil: the flow has no container row
	lookupErr error
	updateErr error
	updates   []database.UpdateContainerStatusParams
}

func (q *fakeContainerDB) GetFlowPrimaryContainer(context.Context, int64) (database.Container, error) {
	if q.lookupErr != nil {
		return database.Container{}, q.lookupErr
	}
	if q.row == nil {
		return database.Container{}, sql.ErrNoRows
	}
	return *q.row, nil
}

func (q *fakeContainerDB) UpdateContainerStatus(
	ctx context.Context, arg database.UpdateContainerStatusParams,
) (database.Container, error) {
	if err := ctx.Err(); err != nil {
		return database.Container{}, err
	}
	q.updates = append(q.updates, arg)
	if q.updateErr != nil {
		return database.Container{}, q.updateErr
	}
	return database.Container{ID: arg.ID, Status: arg.Status}, nil
}

// --- Terminal log ---

// termLogState says which TermLogProvider write a termLogEntry records.
type termLogState string

const (
	termLogMessage    termLogState = ""
	termLogRunning    termLogState = "running"
	termLogNotRunning termLogState = "not running"
)

// termLogEntry is one write recordingTermLog accepted.
type termLogEntry struct {
	state       termLogState
	msgType     database.TermlogType
	msg         string
	containerID int64
	taskID      *int64
	subtaskID   *int64
}

// recordingTermLog stands in for the terminal log: it keeps every write it accepts, where it is filed.
type recordingTermLog struct {
	failOn   database.TermlogType // PutMsg refuses the messages of this type
	stateErr error                // ContainerRunning and ContainerNotRunning fail with it

	mu      sync.Mutex
	entries []termLogEntry
}

var _ TermLogProvider = (*recordingTermLog)(nil)

func (l *recordingTermLog) PutMsg(
	ctx context.Context, msgType database.TermlogType, msg string, containerID int64, taskID, subtaskID *int64,
) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if msgType == l.failOn {
		return 0, errors.New("connection reset by peer")
	}
	l.record(termLogEntry{
		state: termLogMessage, msgType: msgType, msg: msg, containerID: containerID, taskID: taskID, subtaskID: subtaskID,
	})
	return 1, nil
}

func (l *recordingTermLog) ContainerRunning(ctx context.Context, containerID int64, taskID, subtaskID *int64) error {
	return l.recordState(ctx, termLogRunning, containerID, taskID, subtaskID)
}

func (l *recordingTermLog) ContainerNotRunning(ctx context.Context, containerID int64, taskID, subtaskID *int64) error {
	return l.recordState(ctx, termLogNotRunning, containerID, taskID, subtaskID)
}

func (l *recordingTermLog) recordState(
	ctx context.Context, state termLogState, containerID int64, taskID, subtaskID *int64,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if l.stateErr != nil {
		return l.stateErr
	}
	l.record(termLogEntry{state: state, containerID: containerID, taskID: taskID, subtaskID: subtaskID})
	return nil
}

func (l *recordingTermLog) record(entry termLogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, entry)
}

func (l *recordingTermLog) all() []termLogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]termLogEntry(nil), l.entries...)
}

// states is every accepted write of the given state, nil when there is none.
func (l *recordingTermLog) states(state termLogState) []termLogEntry {
	var out []termLogEntry
	for _, entry := range l.all() {
		if entry.state == state {
			out = append(out, entry)
		}
	}
	return out
}

// written joins every message the terminal logged, to tell "printed nothing" from "printed but dropped".
func (l *recordingTermLog) written() string {
	var b strings.Builder
	for _, entry := range l.all() {
		if entry.state == termLogMessage {
			b.WriteString(entry.msg)
		}
	}
	return b.String()
}

// --- Embedding API ---

// fakeEmbedder stands in for the embedding API: it keeps every text and answers one fixed vector.
type fakeEmbedder struct {
	err         error // EmbedDocuments fails with it
	none        bool  // EmbedDocuments answers no vectors
	unavailable bool  // IsAvailable reports false

	embedded []string // every text EmbedDocuments was sent
	queries  []string // every text EmbedQuery was sent
}

var _ embeddings.Embedder = (*fakeEmbedder)(nil)

func (e *fakeEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	e.embedded = append(e.embedded, texts...)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if e.err != nil {
		return nil, e.err
	}
	if e.none {
		return nil, nil
	}

	vectors := make([][]float32, len(texts))
	for i := range vectors {
		vectors[i] = []float32{0.5, -1, 0.25}
	}
	return vectors, nil
}

func (e *fakeEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	e.queries = append(e.queries, text)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []float32{0.5, -1, 0.25}, nil
}

func (e *fakeEmbedder) IsAvailable() bool { return !e.unavailable }

// --- Vector store ---

// newFakeVectorStore is a real pgvector.Store over conn that embeds with embedder.
func newFakeVectorStore(t *testing.T, conn *fakeVectorConn, embedder *fakeEmbedder) *pgvector.Store {
	t.Helper()

	conn.embedder = embedder
	store, err := pgvector.New(t.Context(), pgvector.WithConn(conn), pgvector.WithEmbedder(embedder))
	if err != nil {
		t.Fatalf("building the vector store: %v", err)
	}

	return &store
}

// vectorResult is what one similarity query finds, or why it fails.
type vectorResult struct {
	docs []schema.Document
	err  error
}

// vectorQuery is one similarity query as it reached the table, with the question it embedded.
type vectorQuery struct {
	question string
	sql      string
	args     []any
}

// vectorDoc is one document AddDocuments wrote.
type vectorDoc struct {
	id      string
	content string
	meta    map[string]any
	args    []any // every argument of the insert: uuid, document, embedding, metadata, collection
}

// vectorSearch is a similarity query read back from its SQL.
type vectorSearch struct {
	question string
	filters  map[string]string
	limit    int     // how many documents it asked for
	minScore float64 // the lowest score it accepted; 0: no threshold
}

var (
	vectorFilterPattern    = regexp.MustCompile(`cmetadata ->> '(\w+)'\) = \$(\d+)`)
	vectorLimitPattern     = regexp.MustCompile(`LIMIT \$(\d+)`)
	vectorThresholdPattern = regexp.MustCompile(`data\.distance < \$(\d+)`)
)

// fakeVectorConn stands in for the table behind a pgvector.Store: it keeps every query and write.
type fakeVectorConn struct {
	pgvector.PGXConn

	results []vectorResult // the answer to each similarity query in turn; nothing found once spent
	addErr  error          // AddDocuments fails with it

	embedder *fakeEmbedder // the store's, so a query records the question it was asked
	queries  []vectorQuery
	added    []vectorDoc
	closed   int // how often the store closed its connection
}

func (c *fakeVectorConn) Ping(ctx context.Context) error { return ctx.Err() }

func (c *fakeVectorConn) Begin(ctx context.Context) (pgx.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return fakeVectorTx{}, nil
}

func (c *fakeVectorConn) Close() { c.closed++ }

func (c *fakeVectorConn) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	query := vectorQuery{sql: sql, args: args}
	if c.embedder != nil && len(c.embedder.queries) > 0 {
		query.question = c.embedder.queries[len(c.embedder.queries)-1]
	}
	c.queries = append(c.queries, query)
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(c.results) == 0 {
		return &fakeVectorRows{}, nil
	}
	next := c.results[0]
	c.results = c.results[1:]
	if next.err != nil {
		return nil, next.err
	}

	return &fakeVectorRows{docs: next.docs}, nil
}

func (c *fakeVectorConn) SendBatch(ctx context.Context, batch *pgx.Batch) pgx.BatchResults {
	if err := ctx.Err(); err != nil {
		return fakeVectorBatch{err: err}
	}
	for _, query := range batch.QueuedQueries {
		// VALUES(uuid, document, embedding, cmetadata, collection_id)
		c.added = append(c.added, vectorDoc{
			id:      query.Arguments[0].(string),
			content: query.Arguments[1].(string),
			meta:    query.Arguments[3].(map[string]any),
			args:    query.Arguments,
		})
	}

	return fakeVectorBatch{err: c.addErr}
}

func (c *fakeVectorConn) searches() []vectorSearch {
	var out []vectorSearch
	for _, query := range c.queries {
		out = append(out, query.search())
	}
	return out
}

func (c *fakeVectorConn) search(t *testing.T, i int) vectorSearch {
	t.Helper()

	if i >= len(c.queries) {
		t.Fatalf("only %d similarity queries reached the store", len(c.queries))
	}
	return c.queries[i].search()
}

func (q vectorQuery) search() vectorSearch {
	arg := func(placeholder string) any {
		n, _ := strconv.Atoi(placeholder)
		return q.args[n-1]
	}

	search := vectorSearch{question: q.question, filters: map[string]string{}}
	for _, match := range vectorFilterPattern.FindAllStringSubmatch(q.sql, -1) {
		search.filters[match[1]] = fmt.Sprint(arg(match[2]))
	}
	if match := vectorLimitPattern.FindStringSubmatch(q.sql); match != nil {
		search.limit, _ = arg(match[1]).(int)
	}
	if match := vectorThresholdPattern.FindStringSubmatch(q.sql); match != nil {
		// pgvector binds 1 - threshold through float32 (0.2 arrives as 0.19999998807907104); six places undo it.
		search.minScore = math.Round((1-arg(match[1]).(float64))*1e6) / 1e6
	}

	return search
}

// fakeVectorTx answers the schema bootstrap pgvector.New runs in a transaction.
type fakeVectorTx struct{ pgx.Tx }

func (fakeVectorTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (fakeVectorTx) QueryRow(context.Context, string, ...any) pgx.Row { return fakeCollectionRow{} }
func (fakeVectorTx) Commit(context.Context) error                     { return nil }
func (fakeVectorTx) Rollback(context.Context) error                   { return nil }

// fakeCollectionRow is the collection the bootstrap finds.
type fakeCollectionRow struct{}

func (fakeCollectionRow) Scan(dest ...any) error {
	*dest[0].(*string) = "collection"
	return nil
}

// fakeVectorRows are the documents one similarity query found.
type fakeVectorRows struct {
	pgx.Rows
	docs []schema.Document
	at   int
}

func (r *fakeVectorRows) Next() bool {
	r.at++
	return r.at <= len(r.docs)
}

func (r *fakeVectorRows) Scan(dest ...any) error {
	doc := r.docs[r.at-1]
	*dest[0].(*string) = doc.PageContent
	*dest[1].(*map[string]any) = doc.Metadata
	*dest[2].(*float32) = doc.Score
	return nil
}

func (*fakeVectorRows) Close()     {}
func (*fakeVectorRows) Err() error { return nil }

// fakeVectorBatch is the outcome of one AddDocuments batch.
type fakeVectorBatch struct {
	pgx.BatchResults
	err error
}

func (b fakeVectorBatch) Close() error { return b.err }

// --- Vector store log ---

// vectorStoreLogEntry is one entry recordingVectorStoreLog accepted.
type vectorStoreLogEntry struct {
	initiator database.MsgchainType
	executor  database.MsgchainType
	filter    string
	query     string
	action    database.VecstoreActionType
	result    string
	taskID    *int64
	subtaskID *int64
}

// recordingVectorStoreLog stands in for the vector store log: it keeps every entry it is given, whole.
type recordingVectorStoreLog struct {
	err     error // every write fails with it, after it is kept
	entries []vectorStoreLogEntry
}

var _ VectorStoreLogProvider = (*recordingVectorStoreLog)(nil)

func (l *recordingVectorStoreLog) PutLog(
	ctx context.Context,
	initiator database.MsgchainType,
	executor database.MsgchainType,
	filter json.RawMessage,
	query string,
	action database.VecstoreActionType,
	result string,
	taskID *int64,
	subtaskID *int64,
) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	l.entries = append(l.entries, vectorStoreLogEntry{
		initiator: initiator, executor: executor, filter: string(filter), query: query,
		action: action, result: result, taskID: taskID, subtaskID: subtaskID,
	})
	if l.err != nil {
		return 0, l.err
	}
	return int64(len(l.entries)), nil
}

// --- Knowledge table and page ---

// recordingKnowledgeDB stands in for the knowledge table: it counts inserts and keeps the last one.
type recordingKnowledgeDB struct {
	database.Querier

	err error // every insert fails with it

	inserts   int // every insert sent, accepted or not
	document  string
	embedding any
	meta      map[string]any
}

func (d *recordingKnowledgeDB) InsertKnowledgeDocument(
	ctx context.Context, arg database.InsertKnowledgeDocumentParams,
) (string, error) {
	d.inserts++
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if d.err != nil {
		return "", d.err
	}

	d.document, d.embedding = arg.Document.String, arg.Embedding
	d.meta = map[string]any{}
	if err := json.Unmarshal(arg.Cmetadata, &d.meta); err != nil {
		return "", err
	}
	return "doc-1", nil
}

// recordingKnowledgeProvider stands in for the knowledge page's publisher, dropping a done context's event.
type recordingKnowledgeProvider struct {
	created []*model.KnowledgeDocument
}

var _ KnowledgeProvider = (*recordingKnowledgeProvider)(nil)

func (k *recordingKnowledgeProvider) KnowledgeDocumentCreated(ctx context.Context, doc *model.KnowledgeDocument) {
	if ctx.Err() != nil {
		return
	}
	k.created = append(k.created, doc)
}

// knowledgeRig wires a knowledge door (search.go, guide.go, code.go) to the doubles above.
type knowledgeRig struct {
	conn         *fakeVectorConn
	store        *pgvector.Store
	embedder     *fakeEmbedder       // the vector store's
	doorEmbedder embeddings.Embedder // the door's: embedder, unless a test takes it away
	db           *recordingKnowledgeDB
	log          *recordingVectorStoreLog
	known        *recordingKnowledgeProvider
}

func newKnowledgeRig(t *testing.T) *knowledgeRig {
	t.Helper()

	obs.InitObserver(context.Background(), nil, nil, []logrus.Level{})

	rig := &knowledgeRig{
		conn:     &fakeVectorConn{},
		embedder: &fakeEmbedder{},
		db:       &recordingKnowledgeDB{},
		log:      &recordingVectorStoreLog{},
		known:    &recordingKnowledgeProvider{},
	}
	rig.store = newFakeVectorStore(t, rig.conn, rig.embedder)
	rig.doorEmbedder = rig.embedder

	return rig
}

// handle runs a call under an agent context: a door logs to the vector store only when it knows the agent.
func (r *knowledgeRig) handle(t *testing.T, tool Tool, name string, args json.RawMessage) string {
	t.Helper()

	result, err := tool.Handle(PutAgentContext(t.Context(), database.MsgchainTypeSearcher), name, args)
	if err != nil {
		t.Fatalf("%s returned an error: %v", name, err)
	}

	return result
}

// stored is the one document written: via the vector store within the embedding limit, else to the database.
func (r *knowledgeRig) stored(t *testing.T) (id, content string, meta map[string]any) {
	t.Helper()

	switch {
	case len(r.conn.added) == 1 && r.db.document == "":
		return r.conn.added[0].id, r.conn.added[0].content, r.conn.added[0].meta
	case len(r.conn.added) == 0 && r.db.document != "":
		return "doc-1", r.db.document, r.db.meta
	default:
		t.Fatalf("expected one written document, got %d through the vector store and %.60q in the database",
			len(r.conn.added), r.db.document)
		return "", "", nil
	}
}

func (r *knowledgeRig) announced(t *testing.T) *model.KnowledgeDocument {
	t.Helper()

	if len(r.known.created) != 1 {
		t.Fatalf("announced %d documents to the knowledge page, want the one it stored", len(r.known.created))
	}

	return r.known.created[0]
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

// scraperAnswer is what fakeScraper answers on one path.
type scraperAnswer struct {
	status int
	body   string
}

// fakeScraper stands in for the scraper; it refuses a path off pages, a url other than target, a partial screenshot.
func fakeScraper(t *testing.T, target string, pages map[string]scraperAnswer) *httptest.Server {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		answer, ok := pages[r.URL.Path]
		query := r.URL.Query()
		switch {
		case !ok:
			w.WriteHeader(http.StatusNotFound)
			return
		case query.Get("url") != target:
			w.WriteHeader(http.StatusBadRequest)
			return
		case r.URL.Path == "/screenshot" && query.Get("fullPage") != "true":
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(answer.status)
		fmt.Fprint(w, answer.body)
	}))
	t.Cleanup(ts.Close)

	return ts
}
