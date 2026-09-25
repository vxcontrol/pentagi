package services

import (
	"context"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestGraphqlDeadline_OperationDeadline_BoundsOnlyOrdinaryOperations(t *testing.T) {
	for _, tc := range []struct {
		name         string
		operation    ast.Operation
		field        string
		wantDeadline bool
	}{
		{"an ordinary mutation, which could otherwise hold a caller forever", ast.Mutation, "createFlow", true},
		{"a provider test, a model run whose answer may take minutes", ast.Mutation, "testProvider", false},
		{"an agent test, a model run whose answer may take minutes", ast.Mutation, "testAgent", false},
		{"a subscription, which lives as long as the client keeps it", ast.Subscription, "flowUpdated", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := graphql.WithOperationContext(context.Background(), &graphql.OperationContext{
				Operation: &ast.OperationDefinition{
					Operation:    tc.operation,
					SelectionSet: ast.SelectionSet{&ast.Field{Name: tc.field}},
				},
			})

			var (
				inner       context.Context
				deadline    time.Time
				hasDeadline bool
			)
			responses := operationDeadline(time.Minute)(ctx, func(ctx context.Context) graphql.ResponseHandler {
				inner = ctx

				return func(ctx context.Context) *graphql.Response {
					deadline, hasDeadline = ctx.Deadline()

					return &graphql.Response{}
				}
			})
			responses(inner)

			assert.Equal(t, tc.wantDeadline, hasDeadline)
			if tc.wantDeadline {
				assert.WithinDuration(t, time.Now().Add(time.Minute), deadline, 5*time.Second)
			}
		})
	}
}
