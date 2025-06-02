package controller

import (
	"context"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

const defaultMaxMessageLength = 2048

type FlowMsgLogWorker interface {
	PutMsg(
		ctx context.Context,
		msgType database.MsglogType,
		taskID, subtaskID *int64,
		streamID int64,
		thinking, msg string,
	) (int64, error)
	PutFlowMsg(
		ctx context.Context,
		msgType database.MsglogType,
		thinking, msg string,
	) (int64, error)
	PutFlowMsgResult(
		ctx context.Context,
		msgType database.MsglogType,
		thinking, msg, result string,
		resultFormat database.MsglogResultFormat,
	) (int64, error)
	PutTaskMsg(
		ctx context.Context,
		msgType database.MsglogType,
		taskID int64,
		thinking, msg string,
	) (int64, error)
	PutTaskMsgResult(
		ctx context.Context,
		msgType database.MsglogType,
		taskID int64,
		thinking, msg, result string,
		resultFormat database.MsglogResultFormat,
	) (int64, error)
	PutSubtaskMsg(
		ctx context.Context,
		msgType database.MsglogType,
		taskID, subtaskID int64,
		thinking, msg string,
	) (int64, error)
	PutSubtaskMsgResult(
		ctx context.Context,
		msgType database.MsglogType,
		taskID, subtaskID int64,
		thinking, msg, result string,
		resultFormat database.MsglogResultFormat,
	) (int64, error)
	UpdateMsgResult(
		ctx context.Context,
		msgID, streamID int64,
		result string,
		resultFormat database.MsglogResultFormat,
	) error
}

type flowMsgLogWorker struct {
	db     database.Querier
	mx     *sync.Mutex
	flowID int64
	pub    subscriptions.FlowPublisher
}

func NewFlowMsgLogWorker(db database.Querier, flowID int64, pub subscriptions.FlowPublisher) FlowMsgLogWorker {
	return &flowMsgLogWorker{
		db:     db,
		mx:     &sync.Mutex{},
		flowID: flowID,
		pub:    pub,
	}
}

func (mlw *flowMsgLogWorker) PutMsg(
	ctx context.Context,
	msgType database.MsglogType,
	taskID, subtaskID *int64,
	streamID int64, // unsupported for now
	thinking, msg string,
) (int64, error) {
	mlw.mx.Lock()
	defer mlw.mx.Unlock()

	return mlw.putMsg(ctx, msgType, taskID, subtaskID, thinking, msg)
}

func (mlw *flowMsgLogWorker) PutFlowMsg(
	ctx context.Context,
	msgType database.MsglogType,
	thinking, msg string,
) (int64, error) {
	mlw.mx.Lock()
	defer mlw.mx.Unlock()

	return mlw.putMsg(ctx, msgType, nil, nil, thinking, msg)
}

func (mlw *flowMsgLogWorker) PutFlowMsgResult(
	ctx context.Context,
	msgType database.MsglogType,
	thinking, msg, result string,
	resultFormat database.MsglogResultFormat,
) (int64, error) {
	mlw.mx.Lock()
	defer mlw.mx.Unlock()

	return mlw.putMsgResult(ctx, msgType, nil, nil, thinking, msg, result, resultFormat)
}

func (mlw *flowMsgLogWorker) PutTaskMsg(
	ctx context.Context,
	msgType database.MsglogType,
	taskID int64,
	thinking, msg string,
) (int64, error) {
	mlw.mx.Lock()
	defer mlw.mx.Unlock()

	return mlw.putMsg(ctx, msgType, &taskID, nil, thinking, msg)
}

func (mlw *flowMsgLogWorker) PutTaskMsgResult(
	ctx context.Context,
	msgType database.MsglogType,
	taskID int64,
	thinking, msg, result string,
	resultFormat database.MsglogResultFormat,
) (int64, error) {
	mlw.mx.Lock()
	defer mlw.mx.Unlock()

	return mlw.putMsgResult(ctx, msgType, &taskID, nil, thinking, msg, result, resultFormat)
}

func (mlw *flowMsgLogWorker) PutSubtaskMsg(
	ctx context.Context,
	msgType database.MsglogType,
	taskID, subtaskID int64,
	thinking, msg string,
) (int64, error) {
	mlw.mx.Lock()
	defer mlw.mx.Unlock()

	return mlw.putMsg(ctx, msgType, &taskID, &subtaskID, thinking, msg)
}

func (mlw *flowMsgLogWorker) PutSubtaskMsgResult(
	ctx context.Context,
	msgType database.MsglogType,
	taskID, subtaskID int64,
	thinking, msg, result string,
	resultFormat database.MsglogResultFormat,
) (int64, error) {
	mlw.mx.Lock()
	defer mlw.mx.Unlock()

	return mlw.putMsgResult(ctx, msgType, &taskID, &subtaskID, thinking, msg, result, resultFormat)
}

func (mlw *flowMsgLogWorker) UpdateMsgResult(
	ctx context.Context,
	msgID int64,
	streamID int64, // unsupported for now
	result string,
	resultFormat database.MsglogResultFormat,
) error {
	mlw.mx.Lock()
	defer mlw.mx.Unlock()

	msgLog, err := mlw.db.UpdateMsgLogResult(ctx, database.UpdateMsgLogResultParams{
		Result:       sanitizeUTF8(result),
		ResultFormat: resultFormat,
		ID:           msgID,
	})
	if err != nil {
		return err
	}

	mlw.pub.MessageLogUpdated(ctx, msgLog)

	return nil
}

func (mlw *flowMsgLogWorker) putMsg(
	ctx context.Context,
	msgType database.MsglogType,
	taskID, subtaskID *int64,
	thinking, msg string,
) (int64, error) {
	if len(msg) > defaultMaxMessageLength {
		msg = msg[:defaultMaxMessageLength] + "..."
	}

	msgLog, err := mlw.db.CreateMsgLog(ctx, database.CreateMsgLogParams{
		Type:      msgType,
		Message:   sanitizeUTF8(msg),
		Thinking:  database.StringToNullString(sanitizeUTF8(thinking)),
		FlowID:    mlw.flowID,
		TaskID:    database.Int64ToNullInt64(taskID),
		SubtaskID: database.Int64ToNullInt64(subtaskID),
	})
	if err != nil {
		return 0, err
	}

	mlw.pub.MessageLogAdded(ctx, msgLog)

	return msgLog.ID, nil
}

func (mlw *flowMsgLogWorker) putMsgResult(
	ctx context.Context,
	msgType database.MsglogType,
	taskID, subtaskID *int64,
	thinking, msg, result string,
	resultFormat database.MsglogResultFormat,
) (int64, error) {
	if len(msg) > defaultMaxMessageLength {
		msg = msg[:defaultMaxMessageLength] + "..."
	}

	msgLog, err := mlw.db.CreateResultMsgLog(ctx, database.CreateResultMsgLogParams{
		Type:         msgType,
		Message:      sanitizeUTF8(msg),
		Thinking:     database.StringToNullString(sanitizeUTF8(thinking)),
		Result:       sanitizeUTF8(result),
		ResultFormat: resultFormat,
		FlowID:       mlw.flowID,
		TaskID:       database.Int64ToNullInt64(taskID),
		SubtaskID:    database.Int64ToNullInt64(subtaskID),
	})
	if err != nil {
		return 0, err
	}

	mlw.pub.MessageLogAdded(ctx, msgLog)

	return msgLog.ID, nil
}
