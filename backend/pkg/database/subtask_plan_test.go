package database

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func subtaskPlanRow(id int64, item CreateSubtaskParams) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "status", "title", "description", "result", "task_id", "created_at", "updated_at", "context"}).
		AddRow(id, item.Status, item.Title, item.Description, "", item.TaskID, time.Now(), time.Now(), "")
}

func TestSaveSubtaskPlanNormalizesTextAndCommits(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	item := CreateSubtaskParams{TaskID: 9, Status: SubtaskStatusCreated, Title: "验\x00证", Description: "证据\x00\xff\n保留"}
	cleaned := item
	cleaned.Title, cleaned.Description = "验证", "证据�\n保留"
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(deleteSubtasks)).WithArgs(pq.Array([]int64{196, 197})).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectQuery(regexp.QuoteMeta(createSubtask)).WithArgs(item.Status, cleaned.Title, cleaned.Description, item.TaskID).WillReturnRows(subtaskPlanRow(201, cleaned))
	mock.ExpectCommit()
	require.NoError(t, SaveSubtaskPlan(context.Background(), New(db), []int64{196, 197}, []CreateSubtaskParams{item}))
	require.Equal(t, "验\x00证", item.Title)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveSubtaskPlanRollsBackPartialWrites(t *testing.T) {
	for _, replace := range []bool{false, true} {
		name := "generation"
		if replace {
			name = "refinement"
		}
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			item := CreateSubtaskParams{TaskID: 9, Status: SubtaskStatusCreated, Title: "first", Description: "description"}
			failure := errors.New("second row rejected")
			mock.ExpectBegin()
			var ids []int64
			if replace {
				ids = []int64{196, 197}
				mock.ExpectExec(regexp.QuoteMeta(deleteSubtasks)).WithArgs(pq.Array(ids)).WillReturnResult(sqlmock.NewResult(0, 2))
			}
			mock.ExpectQuery(regexp.QuoteMeta(createSubtask)).WithArgs(item.Status, item.Title, item.Description, item.TaskID).WillReturnRows(subtaskPlanRow(201, item))
			mock.ExpectQuery(regexp.QuoteMeta(createSubtask)).WillReturnError(failure)
			mock.ExpectRollback()
			require.ErrorIs(t, SaveSubtaskPlan(context.Background(), New(db), ids, []CreateSubtaskParams{item, item}), failure)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSaveSubtaskPlanDeleteFailureRollsBack(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	failure := errors.New("delete rejected")
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(deleteSubtasks)).WillReturnError(failure)
	mock.ExpectRollback()
	require.ErrorIs(t, SaveSubtaskPlan(context.Background(), New(db), []int64{196}, nil), failure)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveSubtaskPlanCommitFailureIsReturned(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	failure := errors.New("commit failed")
	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(failure)
	require.ErrorIs(t, SaveSubtaskPlan(context.Background(), New(db), nil, nil), failure)
	require.NoError(t, mock.ExpectationsWereMet())
}
