package controller

import (
	"bytes"
	"context"
	"sync"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/providers"

	lru "github.com/hashicorp/golang-lru/v2/expirable"
)

const (
	updateMsgTimeout = 30 * time.Second
	streamCacheSize  = 1000
	streamCacheTTL   = 2 * time.Hour
)

type FlowAssistantLogWorker interface {
	PutMsg(
		ctx context.Context,
		msgType database.MsglogType,
		taskID, subtaskID *int64,
		streamID int64,
		thinking, msg string,
	) (int64, error)
	PutFlowAssistantMsg(
		ctx context.Context,
		msgType database.MsglogType,
		thinking, msg string,
	) (int64, error)
	PutFlowAssistantMsgResult(
		ctx context.Context,
		msgType database.MsglogType,
		thinking, msg, result string,
		resultFormat database.MsglogResultFormat,
	) (int64, error)
	StreamFlowAssistantMsg(
		ctx context.Context,
		chunk *providers.StreamMessageChunk,
	) error
	UpdateMsgResult(
		ctx context.Context,
		msgID, streamID int64,
		result string,
		resultFormat database.MsglogResultFormat,
	) error
}

type flowAssistantLogWorker struct {
	db          database.Querier
	mx          *sync.Mutex
	flowID      int64
	assistantID int64
	results     map[int64]chan *providers.StreamMessageChunk
	streamCache *lru.LRU[int64, int64] // streamID -> msgID
	pub         subscriptions.FlowPublisher
}

func NewFlowAssistantLogWorker(
	db database.Querier, flowID int64, assistantID int64, pub subscriptions.FlowPublisher,
) FlowAssistantLogWorker {
	return &flowAssistantLogWorker{
		db:          db,
		mx:          &sync.Mutex{},
		flowID:      flowID,
		assistantID: assistantID,
		results:     make(map[int64]chan *providers.StreamMessageChunk),
		streamCache: lru.NewLRU[int64, int64](streamCacheSize, nil, streamCacheTTL),
		pub:         pub,
	}
}

func (aslw *flowAssistantLogWorker) PutMsg(
	ctx context.Context,
	msgType database.MsglogType,
	taskID, subtaskID *int64,
	streamID int64,
	thinking, msg string,
) (int64, error) {
	aslw.mx.Lock()
	defer aslw.mx.Unlock()

	return aslw.putMsg(ctx, msgType, taskID, subtaskID, streamID, thinking, msg)
}

func (aslw *flowAssistantLogWorker) PutFlowAssistantMsg(
	ctx context.Context, msgType database.MsglogType, thinking, msg string,
) (int64, error) {
	aslw.mx.Lock()
	defer aslw.mx.Unlock()

	return aslw.putMsg(ctx, msgType, nil, nil, 0, thinking, msg)
}

func (aslw *flowAssistantLogWorker) PutFlowAssistantMsgResult(
	ctx context.Context, msgType database.MsglogType, thinking, msg, result string,
	resultFormat database.MsglogResultFormat,
) (int64, error) {
	aslw.mx.Lock()
	defer aslw.mx.Unlock()

	return aslw.putMsgResult(ctx, msgType, nil, nil, thinking, msg, result, resultFormat)
}

func (aslw *flowAssistantLogWorker) StreamFlowAssistantMsg(
	ctx context.Context, chunk *providers.StreamMessageChunk,
) error {
	aslw.mx.Lock()
	defer aslw.mx.Unlock()

	return aslw.appendMsgResult(ctx, chunk)
}

func (aslw *flowAssistantLogWorker) UpdateMsgResult(
	ctx context.Context,
	msgID, streamID int64,
	result string,
	resultFormat database.MsglogResultFormat,
) error {
	aslw.mx.Lock()
	defer aslw.mx.Unlock()

	msgLog, err := aslw.db.GetFlowAssistantLog(ctx, msgID)
	if err != nil {
		return err
	}

	ch, workerFound := aslw.results[streamID]
	if workerFound {
		ch <- &providers.StreamMessageChunk{
			Type:         providers.StreamMessageChunkTypeResult,
			MsgType:      msgLog.Type,
			Content:      msgLog.Message,
			Thinking:     msgLog.Thinking.String,
			Result:       result,
			ResultFormat: resultFormat,
			StreamID:     streamID,
		}
		return nil
	}

	msgLog, err = aslw.db.UpdateAssistantLogResult(ctx, database.UpdateAssistantLogResultParams{
		Result:       sanitizeUTF8(result),
		ResultFormat: resultFormat,
		ID:           msgID,
	})
	if err != nil {
		return err
	}

	aslw.pub.AssistantLogUpdated(ctx, msgLog, false)

	return nil
}

func (aslw *flowAssistantLogWorker) putMsg(
	ctx context.Context,
	msgType database.MsglogType,
	taskID, subtaskID *int64,
	streamID int64,
	thinking, msg string,
) (int64, error) {
	if len(msg) > defaultMaxMessageLength {
		msg = msg[:defaultMaxMessageLength] + "..."
	}

	msgID, msgFound := aslw.streamCache.Get(streamID)
	ch, workerFound := aslw.results[streamID]

	if msgFound && workerFound {
		ch <- &providers.StreamMessageChunk{
			Type:     providers.StreamMessageChunkTypeUpdate,
			MsgType:  msgType,
			Content:  msg,
			Thinking: thinking,
			StreamID: streamID,
		}
		return msgID, nil
	} else if msgFound {
		msgLog, err := aslw.db.UpdateAssistantLogContent(ctx, database.UpdateAssistantLogContentParams{
			Type:     msgType,
			Message:  sanitizeUTF8(msg),
			Thinking: database.StringToNullString(sanitizeUTF8(thinking)),
			ID:       msgID,
		})
		if err == nil {
			aslw.pub.AssistantLogUpdated(ctx, msgLog, false)
		}
		return msgID, err
	} else {
		msgLog, err := aslw.db.CreateAssistantLog(ctx, database.CreateAssistantLogParams{
			Type:        msgType,
			Message:     sanitizeUTF8(msg),
			Thinking:    database.StringToNullString(sanitizeUTF8(thinking)),
			FlowID:      aslw.flowID,
			AssistantID: aslw.assistantID,
		})
		if err == nil {
			if streamID != 0 {
				aslw.streamCache.Add(streamID, msgLog.ID)
			}
			aslw.pub.AssistantLogAdded(ctx, msgLog)
			return msgLog.ID, nil
		}
		return 0, err
	}
}

func (aslw *flowAssistantLogWorker) putMsgResult(
	ctx context.Context,
	msgType database.MsglogType,
	taskID, subtaskID *int64,
	thinking, msg, result string,
	resultFormat database.MsglogResultFormat,
) (int64, error) {
	if len(msg) > defaultMaxMessageLength {
		msg = msg[:defaultMaxMessageLength] + "..."
	}

	msgLog, err := aslw.db.CreateResultAssistantLog(ctx, database.CreateResultAssistantLogParams{
		Type:         msgType,
		Message:      sanitizeUTF8(msg),
		Thinking:     database.StringToNullString(sanitizeUTF8(thinking)),
		Result:       sanitizeUTF8(result),
		ResultFormat: resultFormat,
		FlowID:       aslw.flowID,
		AssistantID:  aslw.assistantID,
	})
	if err != nil {
		return 0, err
	}

	aslw.pub.AssistantLogAdded(ctx, msgLog)

	return msgLog.ID, nil
}

func (aslw *flowAssistantLogWorker) appendMsgResult(
	ctx context.Context, chunk *providers.StreamMessageChunk,
) error {
	var (
		err    error
		msgLog database.Assistantlog
	)

	if chunk == nil {
		return nil
	}

	ch, ok := aslw.results[chunk.StreamID]
	if ok {
		ch <- chunk
		return nil
	}

	msgLog, err = aslw.db.CreateAssistantLog(ctx, database.CreateAssistantLogParams{
		Type:        chunk.MsgType,
		Message:     "", // special case for completion answer
		FlowID:      aslw.flowID,
		AssistantID: aslw.assistantID,
	})
	if err != nil {
		return err
	}

	aslw.streamCache.Add(chunk.StreamID, msgLog.ID)
	ch = make(chan *providers.StreamMessageChunk, 50) // safe capacity to avoid deadlock
	aslw.results[chunk.StreamID] = ch                 // it's safe because mutex is used in parent method
	ch <- chunk

	go aslw.workerMsgUpdater(msgLog.ID, chunk.StreamID, ch)

	return nil
}

func (aslw *flowAssistantLogWorker) workerMsgUpdater(
	msgID, streamID int64,
	ch chan *providers.StreamMessageChunk,
) {
	timer := time.NewTimer(updateMsgTimeout)
	defer timer.Stop()

	ctx := context.Background()
	result := ""
	resultFormat := database.MsglogResultFormatPlain
	contentData := make([]byte, 0, defaultMaxMessageLength)
	contentBuf := bytes.NewBuffer(contentData)
	thinkingData := make([]byte, 0, defaultMaxMessageLength)
	thinkingBuf := bytes.NewBuffer(thinkingData)

	msgLog, err := aslw.db.GetFlowAssistantLog(ctx, msgID)
	if err != nil {
		// generic fields
		msgLog = database.Assistantlog{
			ID:          msgID,
			FlowID:      aslw.flowID,
			AssistantID: aslw.assistantID,
			CreatedAt:   database.TimeToNullTime(time.Now()),
		}
	}

	newLog := func(msgType database.MsglogType, content, thinking string) database.Assistantlog {
		return database.Assistantlog{
			ID:           msgID,
			Type:         msgType,
			Message:      content,
			Thinking:     database.StringToNullString(thinking),
			Result:       result,
			ResultFormat: resultFormat,
			FlowID:       msgLog.FlowID,
			AssistantID:  msgLog.AssistantID,
			CreatedAt:    msgLog.CreatedAt,
		}
	}

	processChunk := func(chunk *providers.StreamMessageChunk) {
		switch chunk.Type {
		case providers.StreamMessageChunkTypeUpdate:
			thinkingBuf.Reset()
			contentBuf.Reset()
			thinkingBuf.WriteString(chunk.Thinking)
			contentBuf.WriteString(chunk.Content)
			fallthrough // update both thinking and content, send it via publisher

		case providers.StreamMessageChunkTypeFlush:
			content, thinking := contentBuf.String(), thinkingBuf.String()
			msgLog, err = aslw.db.UpdateAssistantLogContent(ctx, database.UpdateAssistantLogContentParams{
				Type:     chunk.MsgType,
				Message:  sanitizeUTF8(content),
				Thinking: database.StringToNullString(sanitizeUTF8(thinking)),
				ID:       msgID,
			})
			if err == nil {
				aslw.pub.AssistantLogUpdated(ctx, msgLog, false)
			}

		case providers.StreamMessageChunkTypeContent:
			contentBuf.WriteString(chunk.Content)
			aslw.pub.AssistantLogUpdated(ctx, newLog(chunk.MsgType, chunk.Content, ""), true)

		case providers.StreamMessageChunkTypeThinking:
			thinkingBuf.WriteString(chunk.Thinking)
			aslw.pub.AssistantLogUpdated(ctx, newLog(chunk.MsgType, "", chunk.Thinking), true)

		case providers.StreamMessageChunkTypeResult:
			result = chunk.Result
			resultFormat = chunk.ResultFormat
			content, thinking := contentBuf.String(), thinkingBuf.String()
			msgLog, err = aslw.db.UpdateAssistantLog(ctx, database.UpdateAssistantLogParams{
				Type:         chunk.MsgType,
				Message:      sanitizeUTF8(content),
				Thinking:     database.StringToNullString(sanitizeUTF8(thinking)),
				Result:       sanitizeUTF8(result),
				ResultFormat: resultFormat,
				ID:           msgID,
			})
			if err == nil {
				aslw.pub.AssistantLogUpdated(ctx, msgLog, false)
			}
		}
	}

	for {
		select {
		case <-timer.C:
			aslw.mx.Lock()
			defer aslw.mx.Unlock()

			for i := 0; i < len(ch); i++ {
				processChunk(<-ch)
			}

			if msgLog, err = aslw.db.GetFlowAssistantLog(ctx, msgID); err == nil {
				content, thinking := contentBuf.String(), thinkingBuf.String()
				_, _ = aslw.db.UpdateAssistantLog(ctx, database.UpdateAssistantLogParams{
					Type:         msgLog.Type,
					Message:      sanitizeUTF8(content),
					Thinking:     database.StringToNullString(sanitizeUTF8(thinking)),
					Result:       msgLog.Result,
					ResultFormat: msgLog.ResultFormat,
					ID:           msgID,
				})
			}

			delete(aslw.results, streamID)
			close(ch)

			return

		case chunk := <-ch:
			timer.Reset(updateMsgTimeout)

			processChunk(chunk)
		}
	}
}
