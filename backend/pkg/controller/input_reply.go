package controller

import "sync/atomic"

const (
	replyWaiting int32 = iota
	replyDelivered
	replyAbandoned
)

type inputReply struct {
	state atomic.Int32
	done  chan error
}

func newInputReply() *inputReply {
	return &inputReply{done: make(chan error, 1)}
}

func (r *inputReply) deliver(err error) bool {
	if !r.state.CompareAndSwap(replyWaiting, replyDelivered) {
		return false
	}
	r.done <- err

	return true
}

func (r *inputReply) abandon() (isDelivered bool, err error) {
	if r.state.CompareAndSwap(replyWaiting, replyAbandoned) {
		return false, nil
	}

	return true, <-r.done
}
