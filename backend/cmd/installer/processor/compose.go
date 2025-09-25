package processor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	composeFilePentagi       = "docker-compose.yml"
	composeFileLangfuse      = "docker-compose-langfuse.yml"
	composeFileObservability = "docker-compose-observability.yml"
)

var composeOperationAllStacksOrder = map[ProcessorOperation][]ProductStack{
	ProcessorOperationStart:    {ProductStackObservability, ProductStackLangfuse, ProductStackPentagi},
	ProcessorOperationStop:     {ProductStackPentagi, ProductStackLangfuse, ProductStackObservability},
	ProcessorOperationUpdate:   {ProductStackObservability, ProductStackLangfuse, ProductStackPentagi},
	ProcessorOperationDownload: {ProductStackObservability, ProductStackLangfuse, ProductStackPentagi},
	ProcessorOperationRemove:   {ProductStackObservability, ProductStackLangfuse, ProductStackPentagi},
	ProcessorOperationPurge:    {ProductStackObservability, ProductStackLangfuse, ProductStackPentagi},
}

type composeOperationsImpl struct {
	processor *processor
}

func newComposeOperations(p *processor) composeOperations {
	return &composeOperationsImpl{processor: p}
}

// startStack starts Docker Compose stack with dependency ordering
func (c *composeOperationsImpl) startStack(ctx context.Context, stack ProductStack, state *operationState) error {
	return c.performStackOperation(ctx, stack, state, ProcessorOperationStart, "start")
}

// stopStack stops Docker Compose stack
func (c *composeOperationsImpl) stopStack(ctx context.Context, stack ProductStack, state *operationState) error {
	return c.performStackOperation(ctx, stack, state, ProcessorOperationStop, "stop")
}

// restartStack restarts Docker Compose stack (stop + start to avoid race conditions for dependencies)
func (c *composeOperationsImpl) restartStack(ctx context.Context, stack ProductStack, state *operationState) error {
	if err := c.stopStack(ctx, stack, state); err != nil {
		return err
	}

	// brief pause to ensure clean shutdown
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Second):
	}

	return c.startStack(ctx, stack, state)
}

// updateStack performs rolling update with health checks (also used for install)
func (c *composeOperationsImpl) updateStack(ctx context.Context, stack ProductStack, state *operationState) error {
	return c.performStackOperation(ctx, stack, state, ProcessorOperationUpdate, "up", "-d")
}

func (c *composeOperationsImpl) downloadStack(ctx context.Context, stack ProductStack, state *operationState) error {
	return c.performStackOperation(ctx, stack, state, ProcessorOperationDownload, "pull")
}

func (c *composeOperationsImpl) removeStack(ctx context.Context, stack ProductStack, state *operationState) error {
	return c.performStackOperation(ctx, stack, state, ProcessorOperationRemove, "down")
}

func (c *composeOperationsImpl) purgeStack(ctx context.Context, stack ProductStack, state *operationState) error {
	return c.performStackOperation(ctx, stack, state, ProcessorOperationPurge, "down", "-v")
}

// purgeImagesStack is a stricter purge that also removes images referenced by the compose services
func (c *composeOperationsImpl) purgeImagesStack(ctx context.Context, stack ProductStack, state *operationState) error {
	return c.performStackOperation(ctx, stack, state, ProcessorOperationPurge, "down", "--rmi", "all", "-v")
}

func (c *composeOperationsImpl) performStackOperation(
	ctx context.Context, stack ProductStack, state *operationState, operation ProcessorOperation, args ...string,
) error {
	switch stack {
	case ProductStackPentagi:
		return c.wrapPerformStackCommand(ctx, stack, state, operation, args...)

	case ProductStackLangfuse, ProductStackObservability:
		switch operation {
		// for destructive operations we must always allow compose to run, even if stack is disabled/external now
		case ProcessorOperationRemove, ProcessorOperationPurge, ProcessorOperationStop:
			return c.wrapPerformStackCommand(ctx, stack, state, operation, args...)
		// for non-destructive operations (start/update/download) honor embedded mode only
		default:
			if c.processor.isEmbeddedDeployment(stack) {
				return c.wrapPerformStackCommand(ctx, stack, state, operation, args...)
			}
			return nil
		}

	case ProductStackAll, ProductStackCompose:
		for _, s := range composeOperationAllStacksOrder[operation] {
			if err := c.performStackOperation(ctx, s, state, operation, args...); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("operation %s not applicable for stack %s", operation, stack)
	}
}

func (c *composeOperationsImpl) wrapPerformStackCommand(
	ctx context.Context, stack ProductStack, state *operationState, operation ProcessorOperation, args ...string,
) error {
	msgs := c.getMessages(stack, operation)
	c.processor.appendLog(msgs.Enter, stack, state)

	err := c.performStackCommand(ctx, stack, state, args...)
	if err != nil {
		c.processor.appendLog(fmt.Sprintf("%s: %s\n", msgs.Error, err.Error()), stack, state)
	} else {
		c.processor.appendLog(msgs.Exit+"\n", stack, state)
	}

	return err
}

func (c *composeOperationsImpl) performStackCommand(
	ctx context.Context, stack ProductStack, state *operationState, args ...string,
) error {
	envPath := c.processor.state.GetEnvPath()
	composeFile, err := c.determineComposeFile(stack)
	if err != nil {
		return err
	}
	workingDir := filepath.Dir(envPath)
	composePath := filepath.Join(workingDir, composeFile)

	// check if files exist
	if err := c.processor.isFileExists(composePath); err != nil {
		return err
	}
	if err := c.processor.isFileExists(envPath); err != nil {
		return err
	}

	// build docker compose command
	args = append([]string{"compose", "--env-file", envPath, "-f", composePath}, args...)
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = workingDir
	cmd.Env = os.Environ()
	// stacks are processed one by one, so we can ignore orphans
	// orphans containers are removed by specific stack operations in main logic
	cmd.Env = append(cmd.Env, "COMPOSE_IGNORE_ORPHANS=1")
	// force Python unbuffered output to prevent incomplete data loss
	cmd.Env = append(cmd.Env, "PYTHONUNBUFFERED=1")

	return c.processor.runCommand(cmd, stack, state)
}

func (c *composeOperationsImpl) determineComposeFile(stack ProductStack) (string, error) {
	switch stack {
	case ProductStackPentagi:
		return composeFilePentagi, nil
	case ProductStackLangfuse:
		return composeFileLangfuse, nil
	case ProductStackObservability:
		return composeFileObservability, nil
	default:
		return "", fmt.Errorf("stack %s not supported", stack)
	}
}

func (c *composeOperationsImpl) getMessages(stack ProductStack, operation ProcessorOperation) SubsystemOperationMessage {
	msgs := SubsystemOperationMessages[SubsystemCompose][operation]
	return SubsystemOperationMessage{
		Enter: fmt.Sprintf(msgs.Enter, stack),
		Exit:  fmt.Sprintf(msgs.Exit, stack),
		Error: fmt.Sprintf(msgs.Error, stack),
	}
}
