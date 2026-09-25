package services

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net"
	"strconv"

	"pentagi/pkg/controller"
	"pentagi/pkg/database/knowledge"
	"pentagi/pkg/graph"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/version"

	"github.com/99designs/gqlgen/graphql"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

const (
	gqlCodeNotFound         = "NOT_FOUND"
	gqlCodeAlreadyExists    = "ALREADY_EXISTS"
	gqlCodeInvalidReference = "INVALID_REFERENCE"
	gqlCodeInternal         = "INTERNAL"
	gqlCodeUnauthenticated  = "UNAUTHENTICATED"
	gqlCodeForbidden        = "FORBIDDEN"
	gqlCodeBadUserInput     = "BAD_USER_INPUT"
)

func redactGqlError(err error) (msg string, code string, ok bool) {
	if errors.Is(err, graph.ErrUnauthenticated) {
		return "authentication required", gqlCodeUnauthenticated, true
	}

	if errors.Is(err, graph.ErrForbidden) {
		return "not permitted", gqlCodeForbidden, true
	}

	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, controller.ErrFlowNotFound) {
		return "not found", gqlCodeNotFound, true
	}

	if errors.Is(err, knowledge.ErrInvalidDocument) {
		return err.Error(), gqlCodeBadUserInput, true
	}

	var numErr *strconv.NumError
	if errors.As(err, &numErr) {
		return "argument is not a valid number", gqlCodeBadUserInput, true
	}

	var netErr net.Error
	if errors.As(err, &netErr) || errors.Is(err, driver.ErrBadConn) {
		return "a required service is unavailable", gqlCodeInternal, true
	}

	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return "already exists", gqlCodeAlreadyExists, true
		case pgForeignKeyViolation:
			return "referenced record does not exist", gqlCodeInvalidReference, true
		}
		return "internal server error", gqlCodeInternal, true
	}

	return "", "", false
}

func newGqlErrorPresenter(entry *logrus.Entry) graphql.ErrorPresenterFunc {
	return func(ctx context.Context, e error) *gqlerror.Error {
		presented := graphql.DefaultErrorPresenter(ctx, e)

		msg, code, ok := redactGqlError(e)
		if !ok {
			return presented
		}

		level, label := logger.LevelForGqlCode(code), "graphql internal error"
		switch code {
		case gqlCodeUnauthenticated, gqlCodeForbidden:
			label = "graphql request refused"
		case gqlCodeNotFound:
			label = "graphql record not found"
		case gqlCodeAlreadyExists:
			label = "graphql constraint violation"
		case gqlCodeInvalidReference:
			label = "graphql invalid reference"
		case gqlCodeBadUserInput:
			label = "graphql invalid argument"
		}

		entry.WithError(e).WithField("gql.code", code).Log(level, label)

		original := e.Error()

		presented.Message = msg
		if presented.Extensions == nil {
			presented.Extensions = map[string]any{}
		}
		presented.Extensions["code"] = code
		if version.IsDevelopMode() {
			presented.Extensions["error"] = original
		}

		return presented
	}
}
