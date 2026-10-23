package providers

import (
	"context"
	"fmt"

	"pentagi/pkg/database"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
)

// maxUnverifiedDoneRejections is how many times a successful done with nothing delegated in its
// subtask is sent back. A second identical done is accepted: a third would trip the repeating tool
// call detector, which ends the chain without running the barrier and fails the subtask.
const maxUnverifiedDoneRejections = 1

const unverifiedDoneReason = "done was not accepted: nothing has been delegated in this subtask yet, " +
	"so no specialist has produced the result you are reporting. Delegate the work to the matching " +
	"specialist first and report only what it returns; never state command output or findings that no " +
	"specialist produced. If this subtask truly needs no execution or research, call done again and " +
	"say so in the result."

// checkDoneEvidence sends a successful done back to the agent when no specialist was delegated to in
// the subtask. A failed result, a repeat after the rejection cap or a failed lookup is let through.
func (fp *flowProvider) checkDoneEvidence(
	ctx context.Context,
	subtaskID int64,
	done tools.Done,
	rejections *int,
	logger *logrus.Entry,
) error {
	if !done.Success.Bool() {
		return nil
	}

	calls, err := fp.db.GetSubtaskToolcalls(ctx, database.Int64ToNullInt64(&subtaskID))
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("failed to get subtask %d tool calls: %w", subtaskID, ctxErr)
		}
		logger.WithError(err).Warn("failed to check the subtask for delegations, accepting done")
		return nil
	}

	for _, call := range calls {
		if tools.GetToolType(call.Name) == tools.AgentToolType {
			return nil
		}
	}

	if *rejections >= maxUnverifiedDoneRejections {
		logger.WithField("rejections", *rejections).
			Warn("done reported success with nothing delegated in the subtask again, accepting it")
		return nil
	}

	*rejections++
	logger.Warn("done reported success with nothing delegated in the subtask, sending it back")

	return &tools.BarrierDeclinedError{Reason: unverifiedDoneReason}
}
