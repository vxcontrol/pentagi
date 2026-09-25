package controller

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInputReply_GoesToExactlyOneSide(t *testing.T) {
	refused := errors.New("refused")

	t.Run("delivered before the caller stops waiting", func(t *testing.T) {
		reply := newInputReply()

		assert.True(t, reply.deliver(refused))
		isDelivered, err := reply.abandon()
		assert.True(t, isDelivered)
		assert.ErrorIs(t, err, refused)
	})

	t.Run("abandoned before the worker answers", func(t *testing.T) {
		reply := newInputReply()

		isDelivered, _ := reply.abandon()
		assert.False(t, isDelivered)
		assert.False(t, reply.deliver(refused), "nobody is left to read it, so the worker reports it itself")
	})

	t.Run("racing sides never both claim the answer", func(t *testing.T) {
		for range 1000 {
			reply := newInputReply()
			delivered, taken := make(chan bool, 1), make(chan bool, 1)
			go func() { delivered <- reply.deliver(refused) }()
			go func() { isTaken, _ := reply.abandon(); taken <- isTaken }()

			assert.Equal(t, controllerReceive(t, delivered, "the worker to deliver"),
				controllerReceive(t, taken, "the caller to abandon"))
		}
	})
}
