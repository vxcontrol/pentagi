package services

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

const OperationTimeout = 30 * time.Second

// unboundedOperations are model runs the operator asked for: taking minutes is
// their answer, not a fault.
var unboundedOperations = map[string]struct{}{
	"testAgent":    {},
	"testProvider": {},
}

func isUnbounded(ctx context.Context) bool {
	oc := graphql.GetOperationContext(ctx)
	if oc.Operation == nil {
		return true
	}

	if oc.Operation.Operation == ast.Subscription {
		return true
	}

	for _, selection := range oc.Operation.SelectionSet {
		field, ok := selection.(*ast.Field)
		if !ok {
			continue
		}

		if _, unbounded := unboundedOperations[field.Name]; unbounded {
			return true
		}
	}

	return false
}

func operationDeadline(timeout time.Duration) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		if isUnbounded(ctx) {
			return next(ctx)
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		responses := next(ctx)

		return func(ctx context.Context) *graphql.Response {
			response := responses(ctx)
			if response == nil {
				cancel()
			}

			return response
		}
	}
}
