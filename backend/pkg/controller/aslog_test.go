package controller

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/providers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// persistFakeQuerier fails the Nth write of each kind, every read from failGetFrom on, and on a done context.
type persistFakeQuerier struct {
	database.Querier

	mx            sync.Mutex
	contentWrites []string
	contentTypes  []database.MsglogType
	fullWrites    []string
	fullResults   []string
	fullTypes     []database.MsglogType
	deleted       []int64
	failContentOn int
	contentCalls  int
	failFullOn    int
	fullCalls     int
	failGetFrom   int
	getCalls      int
	readBack      int
}

func (q *persistFakeQuerier) GetFlowAssistantLog(ctx context.Context, id int64) (database.Assistantlog, error) {
	q.mx.Lock()
	defer q.mx.Unlock()
	q.getCalls++

	if err := ctx.Err(); err != nil {
		return database.Assistantlog{}, err
	}
	if q.failGetFrom > 0 && q.getCalls >= q.failGetFrom {
		return database.Assistantlog{}, errors.New("connection reset")
	}

	return database.Assistantlog{ID: id, FlowID: 7, AssistantID: 3}, nil
}

func (q *persistFakeQuerier) UpdateAssistantLogContent(
	ctx context.Context, arg database.UpdateAssistantLogContentParams,
) (database.Assistantlog, error) {
	q.mx.Lock()
	defer q.mx.Unlock()
	q.contentCalls++
	q.readBack++

	if err := ctx.Err(); err != nil {
		return database.Assistantlog{}, err
	}
	if q.contentCalls == q.failContentOn {
		return database.Assistantlog{}, errors.New("connection reset")
	}
	q.contentWrites = append(q.contentWrites, arg.Message)
	q.contentTypes = append(q.contentTypes, arg.Type)

	return database.Assistantlog{ID: arg.ID, Message: arg.Message}, nil
}

func (q *persistFakeQuerier) SaveAssistantLogContent(
	ctx context.Context, arg database.SaveAssistantLogContentParams,
) error {
	q.mx.Lock()
	defer q.mx.Unlock()
	q.contentCalls++

	if err := ctx.Err(); err != nil {
		return err
	}
	if q.contentCalls == q.failContentOn {
		return errors.New("connection reset")
	}
	q.contentWrites = append(q.contentWrites, arg.Message)
	q.contentTypes = append(q.contentTypes, arg.Type)

	return nil
}

func (q *persistFakeQuerier) SaveAssistantLog(ctx context.Context, arg database.SaveAssistantLogParams) error {
	q.mx.Lock()
	defer q.mx.Unlock()
	q.fullCalls++

	if err := ctx.Err(); err != nil {
		return err
	}
	if q.fullCalls == q.failFullOn {
		return errors.New("connection reset")
	}
	q.fullWrites = append(q.fullWrites, arg.Message)
	q.fullResults = append(q.fullResults, arg.Result)
	q.fullTypes = append(q.fullTypes, arg.Type)

	return nil
}

func (q *persistFakeQuerier) UpdateAssistantLog(
	ctx context.Context, arg database.UpdateAssistantLogParams,
) (database.Assistantlog, error) {
	q.mx.Lock()
	defer q.mx.Unlock()
	q.fullCalls++
	q.readBack++

	if err := ctx.Err(); err != nil {
		return database.Assistantlog{}, err
	}
	if q.fullCalls == q.failFullOn {
		return database.Assistantlog{}, errors.New("connection reset")
	}
	q.fullWrites = append(q.fullWrites, arg.Message)
	q.fullResults = append(q.fullResults, arg.Result)
	q.fullTypes = append(q.fullTypes, arg.Type)

	return database.Assistantlog{ID: arg.ID, Type: arg.Type, Message: arg.Message, Result: arg.Result}, nil
}

func (q *persistFakeQuerier) DeleteFlowAssistantLog(ctx context.Context, id int64) error {
	q.mx.Lock()
	defer q.mx.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}
	q.deleted = append(q.deleted, id)

	return nil
}

func (q *persistFakeQuerier) rowsReadBack() int {
	q.mx.Lock()
	defer q.mx.Unlock()

	return q.readBack
}

func (q *persistFakeQuerier) results() []string {
	q.mx.Lock()
	defer q.mx.Unlock()

	return slices.Clone(q.fullResults)
}

func (q *persistFakeQuerier) resultTypes() []database.MsglogType {
	q.mx.Lock()
	defer q.mx.Unlock()

	return slices.Clone(q.fullTypes)
}

func (q *persistFakeQuerier) deletions() []int64 {
	q.mx.Lock()
	defer q.mx.Unlock()

	return slices.Clone(q.deleted)
}

func (q *persistFakeQuerier) types() []database.MsglogType {
	q.mx.Lock()
	defer q.mx.Unlock()

	return slices.Clone(q.contentTypes)
}

func (q *persistFakeQuerier) snapshot() ([]string, []string) {
	q.mx.Lock()
	defer q.mx.Unlock()

	return slices.Clone(q.contentWrites), slices.Clone(q.fullWrites)
}

type persistFakePublisher struct {
	subscriptions.FlowPublisher

	mx     sync.Mutex
	frames []bool // appendPart of each published frame
	logs   []database.Assistantlog
}

func (p *persistFakePublisher) AssistantLogUpdated(_ context.Context, log database.Assistantlog, appendPart bool) {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.frames = append(p.frames, appendPart)
	p.logs = append(p.logs, log)
}

func (p *persistFakePublisher) published() []database.Assistantlog {
	p.mx.Lock()
	defer p.mx.Unlock()

	return slices.Clone(p.logs)
}

func (p *persistFakePublisher) snapshot() []bool {
	p.mx.Lock()
	defer p.mx.Unlock()

	return slices.Clone(p.frames)
}

// aslogStartUpdater runs the updater of message 11 on stream 42 of flow 7, assistant 3.
func aslogStartUpdater(
	db *persistFakeQuerier,
) (*flowAssistantLogWorker, chan *providers.StreamMessageChunk, *persistFakePublisher) {
	pub := &persistFakePublisher{}
	worker := NewFlowAssistantLogWorker(db, 7, 3, pub).(*flowAssistantLogWorker)
	ch := make(chan *providers.StreamMessageChunk, 16)
	worker.results[42] = ch

	go worker.workerMsgUpdater(11, 42, ch)

	return worker, ch, pub
}

func aslogChunk(
	kind providers.StreamMessageChunkType, msgType database.MsglogType, content string,
) *providers.StreamMessageChunk {
	return &providers.StreamMessageChunk{Type: kind, MsgType: msgType, Content: content}
}

func aslogResult(result string) *providers.StreamMessageChunk {
	return &providers.StreamMessageChunk{
		Type:         providers.StreamMessageChunkTypeResult,
		MsgType:      database.MsglogTypeAnswer,
		Result:       result,
		ResultFormat: database.MsglogResultFormatPlain,
	}
}

func waitFor(t *testing.T, within time.Duration, what string, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %s", what)
}

func aslogStored(db *persistFakeQuerier) func() bool {
	return func() bool {
		content, _ := db.snapshot()

		return len(content) > 0
	}
}

func TestAslog_WorkerMsgUpdater_PeriodicSaveStoresTheRunningMessage(t *testing.T) {
	db := &persistFakeQuerier{}
	_, ch, pub := aslogStartUpdater(db)

	for _, part := range []string{"Hel", "lo ", "World"} {
		ch <- aslogChunk(providers.StreamMessageChunkTypeContent, database.MsglogTypeAnswer, part)
	}

	waitFor(t, 15*time.Second, "the message to reach storage without a closing chunk", aslogStored(db))

	content, _ := db.snapshot()
	assert.Equal(t, "Hello World", content[len(content)-1], "the whole running message is stored")
	assert.Zero(t, db.rowsReadBack(), "the periodic save does not ask the database for the row back")
	for i, appendPart := range pub.snapshot() {
		assert.True(t, appendPart, "frame %d was published as a snapshot during accumulation", i)
	}
}

func TestAslog_WorkerMsgUpdater_RewritesKeepTheStreamedType(t *testing.T) {
	for _, tc := range []struct {
		name   string
		result bool
	}{
		{name: "the periodic save of a failed update"},
		{name: "the result write", result: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &persistFakeQuerier{failContentOn: 1}
			worker, ch, _ := aslogStartUpdater(db)

			ch <- aslogChunk(providers.StreamMessageChunkTypeUpdate, database.MsglogTypeTerminal, "ls -la /etc")

			written := db.types
			if tc.result {
				require.NoError(t, worker.UpdateMsgResult(context.Background(), 11, 42, "total 0", database.MsglogResultFormatPlain))
				written = db.resultTypes
			}

			waitFor(t, 15*time.Second, "the rewrite to reach storage", func() bool { return len(written()) > 0 })
			assert.Equal(t, database.MsglogTypeTerminal, written()[0])
		})
	}
}

func TestAslog_WorkerMsgUpdater_FramesAfterAFailedWriteKeepTheirFlow(t *testing.T) {
	db := &persistFakeQuerier{failContentOn: 1}
	_, ch, pub := aslogStartUpdater(db)

	ch <- aslogChunk(providers.StreamMessageChunkTypeUpdate, database.MsglogTypeTerminal, "ls -la /etc")
	ch <- aslogChunk(providers.StreamMessageChunkTypeContent, database.MsglogTypeAnswer, "and here is what it means")

	waitFor(t, 15*time.Second, "a frame published after the failed write", func() bool {
		return len(pub.published()) > 0
	})
	for i, log := range pub.published() {
		assert.Equal(t, [2]int64{7, 3}, [2]int64{log.FlowID, log.AssistantID}, "flow and assistant of frame %d", i)
	}
}

func TestAslog_WorkerMsgUpdater_RetriesAFailedWriteOfAReplyOfTheSameLength(t *testing.T) {
	db := &persistFakeQuerier{failContentOn: 2}
	_, ch, _ := aslogStartUpdater(db)

	ch <- aslogChunk(providers.StreamMessageChunkTypeContent, database.MsglogTypeAnswer, "0123456789")
	waitFor(t, 15*time.Second, "the first reply to reach storage", aslogStored(db))

	ch <- aslogChunk(providers.StreamMessageChunkTypeUpdate, database.MsglogTypeAnswer, "abcdefghij")
	waitFor(t, 15*time.Second, "the rewritten reply to reach storage", func() bool {
		content, _ := db.snapshot()

		return content[len(content)-1] == "abcdefghij"
	})
}

func TestAslog_WorkerMsgUpdater_SavesADeliveredResultWhoseWriteFailed(t *testing.T) {
	if testing.Short() {
		t.Skip("waits out the worker's idle timeout")
	}
	t.Parallel()

	for _, tc := range []struct {
		name string
		db   *persistFakeQuerier
	}{
		{name: "after the row was read", db: &persistFakeQuerier{failFullOn: 1}},
		{name: "when the row cannot be read back", db: &persistFakeQuerier{failFullOn: 1, failGetFrom: 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, ch, _ := aslogStartUpdater(tc.db)

			ch <- aslogChunk(providers.StreamMessageChunkTypeContent, database.MsglogTypeAnswer, "the reply body")
			waitFor(t, 15*time.Second, "the content to reach storage", aslogStored(tc.db))

			ch <- aslogResult("the tool's answer")
			waitFor(t, 60*time.Second, "a write that stores the result", func() bool { return len(tc.db.results()) > 0 })
			assert.Equal(t, "the tool's answer", tc.db.results()[0])
		})
	}
}

func TestAslog_WorkerMsgUpdater_KeepsAReplyOnlyThePeriodicSaveStored(t *testing.T) {
	if testing.Short() {
		t.Skip("waits out the worker's idle timeout")
	}
	t.Parallel()

	db := &persistFakeQuerier{failContentOn: 1}
	_, ch, _ := aslogStartUpdater(db)

	ch <- aslogChunk(providers.StreamMessageChunkTypeUpdate, database.MsglogTypeAnswer, "a reply nobody else managed to store")
	waitFor(t, 15*time.Second, "the periodic write to rescue the reply", aslogStored(db))

	waitFor(t, updateMsgTimeout+15*time.Second, "the worker to give up on the idle stream", func() bool {
		_, full := db.snapshot()

		return len(full) > 0 || len(db.deletions()) > 0
	})
	assert.Empty(t, db.deletions(), "the worker deleted a log the periodic write had stored")
}
