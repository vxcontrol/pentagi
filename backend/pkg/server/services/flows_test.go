package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/controller"
	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/server/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type flowsWorker struct {
	controller.FlowWorker
	touched *bool
}

func (w flowsWorker) Stop(context.Context) error { *w.touched = true; return nil }

func (w flowsWorker) Finish(context.Context) error { *w.touched = true; return nil }

func (w flowsWorker) Rename(context.Context, string) error { *w.touched = true; return nil }

// flowsCtrl answers every flow action with err and records which one it was asked for.
type flowsCtrl struct {
	controller.FlowController
	unloaded      bool
	err           error
	called        string
	flowID        int64
	workerTouched bool
}

func (f *flowsCtrl) GetFlow(context.Context, int64) (controller.FlowWorker, error) {
	if f.unloaded {
		return nil, controller.ErrFlowNotFound
	}

	return flowsWorker{touched: &f.workerTouched}, nil
}

func (f *flowsCtrl) act(name string, flowID int64) error {
	f.called, f.flowID = name, flowID

	return f.err
}

func (f *flowsCtrl) StopFlow(_ context.Context, flowID int64) error {
	return f.act("StopFlow", flowID)
}

func (f *flowsCtrl) FinishFlow(_ context.Context, flowID int64) error {
	return f.act("FinishFlow", flowID)
}

func (f *flowsCtrl) RenameFlow(_ context.Context, flowID int64, _ string) error {
	return f.act("RenameFlow", flowID)
}

func (f *flowsCtrl) ReportFlow(_ context.Context, flowID int64, _ time.Duration) error {
	return f.act("ReportFlow", flowID)
}

func flowsName(name string) *string {
	return &name
}

func TestFlows_GetFlows_ServesThePageItCounts(t *testing.T) {
	for _, tc := range []struct {
		name      string
		rows      string
		target    string
		wantTrace *string
	}{
		{
			name:   "a flow whose trace id is still pending",
			rows:   `(1, 42, 'gpt', 'openai', 'openai', 'tcid', NULL, NULL)`,
			target: "/flows/?page=1&pageSize=5&type=init",
		},
		{
			name:      "a request without a page size",
			rows:      `(1, 42, 'gpt', 'openai', 'openai', 'tcid', 'trace-1', NULL)`,
			target:    "/flows/?page=1&type=init",
			wantTrace: flowsName("trace-1"),
		},
		{
			name: "a soft-deleted flow beside a live one",
			rows: `(1, 42, 'gpt', 'openai', 'openai', 'tcid', 'trace-1', NULL),
				(2, 42, 'gpt', 'openai', 'openai', 'tcid', 'trace-2', '2026-08-01 10:00:00')`,
			target:    "/flows/?page=1&pageSize=5&type=init",
			wantTrace: flowsName("trace-1"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			require.NoError(t, db.Exec(`INSERT INTO flows (id, user_id, model, model_provider_name,
				model_provider_type, tool_call_id_template, trace_id, deleted_at) VALUES `+tc.rows).Error)

			c, w := newFlowFileTestContext(http.MethodGet, tc.target, nil, []string{"flows.view"}, 42, 0)

			NewFlowService(db, nil, nil, nil, nil).GetFlows(c)

			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			var resp struct {
				Status string `json:"status"`
				Data   struct {
					Flows []models.Flow `json:"flows"`
					Total uint64        `json:"total"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			require.Equal(t, "success", resp.Status)
			require.Len(t, resp.Data.Flows, 1,
				"the answer counts %d rows and hands back %d: a page the caller cannot tell from an empty table",
				resp.Data.Total, len(resp.Data.Flows))
			assert.Equal(t, uint64(1), resp.Data.Total, "the total counts only the rows the page serves")
			assert.Equal(t, uint64(1), resp.Data.Flows[0].ID)
			assert.Equal(t, tc.wantTrace, resp.Data.Flows[0].TraceID)
		})
	}
}

func TestFlows_PatchFlow_RoutesStopFinishRenameAndReportThroughTheController(t *testing.T) {
	const (
		finish = `{"action":"finish"}`
		stop   = `{"action":"stop"}`
		rename = `{"action":"rename","name":"renamed"}`
		report = `{"action":"report"}`
	)

	for _, tc := range []struct {
		name     string
		body     string
		unloaded bool
		err      error
		wantCall string
		wantCode int
	}{
		{"finishing a flow", finish, false, nil, "FinishFlow", http.StatusOK},
		{"finishing a flow no worker is loaded for, as a retried finish is", finish, true, nil, "FinishFlow", http.StatusOK},
		{"finishing a flow the controller cannot find", finish, false, controller.ErrFlowNotFound, "FinishFlow", http.StatusNotFound},
		{"stopping a flow", stop, false, nil, "StopFlow", http.StatusOK},
		{"stopping a flow the controller cannot find", stop, false, controller.ErrFlowNotFound, "StopFlow", http.StatusNotFound},
		{"stopping a flow that finished while the caller waited", stop, false, controller.ErrFlowNotLoaded, "StopFlow", http.StatusNotFound},
		{"renaming a flow", rename, false, nil, "RenameFlow", http.StatusOK},
		{"renaming a flow that is no longer loaded", rename, false, controller.ErrFlowNotLoaded, "RenameFlow", http.StatusNotFound},
		{"a report of a flow with nothing open is the caller's mistake", report, false,
			fmt.Errorf("report: %w", controller.ErrNothingToReport), "ReportFlow", http.StatusBadRequest},
		{"a report that could not start is a server fault", report, false,
			errors.New("flow 42 is stopped: context canceled"), "ReportFlow", http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			seedFlow(t, db, 42, 7)
			ctrl := &flowsCtrl{unloaded: tc.unloaded, err: tc.err}

			c, w := newFlowFileTestContext(http.MethodPut, "/flows/42", strings.NewReader(tc.body),
				[]string{"flows.admin"}, 7, 42)
			c.Request.Header.Set("Content-Type", "application/json")

			NewFlowService(db, nil, nil, ctrl, nil).PatchFlow(c)

			require.Equal(t, tc.wantCode, w.Code, w.Body.String())
			assert.Equal(t, tc.wantCall, ctrl.called, "REST and GraphQL must both serialise the action against the flow's other work")
			assert.Equal(t, int64(42), ctrl.flowID)
			assert.False(t, ctrl.workerTouched, "a worker driven behind the controller's back races its own finish or is already gone")
		})
	}
}

type favoritesQuerier struct {
	database.Querier
	dropped *database.DeleteFavoriteFlowParams
}

func (q favoritesQuerier) DeleteFavoriteFlow(
	ctx context.Context, arg database.DeleteFavoriteFlowParams,
) (database.UserPreference, error) {
	if err := ctx.Err(); err != nil {
		return database.UserPreference{}, err
	}
	*q.dropped = arg

	return database.UserPreference{UserID: arg.UserID}, nil
}

type favoritesSubs struct {
	subscriptions.SubscriptionsController
	announcedTo *int64
	announced   *bool
}

type favoritesFlowPublisher struct {
	subscriptions.FlowPublisher
}

func (favoritesFlowPublisher) FlowUpdated(context.Context, database.Flow, []database.Container) {}

func (favoritesFlowPublisher) FlowDeleted(context.Context, database.Flow, []database.Container) {}

func (favoritesSubs) NewFlowPublisher(int64, int64) subscriptions.FlowPublisher {
	return favoritesFlowPublisher{}
}

type favoritesSettingsPublisher struct {
	subscriptions.SettingsPublisher
	announced *bool
}

func (p favoritesSettingsPublisher) SettingsUserUpdated(context.Context, database.UserPreference) {
	*p.announced = true
}

func (s favoritesSubs) NewSettingsPublisher(userID int64) subscriptions.SettingsPublisher {
	*s.announcedTo = userID

	return favoritesSettingsPublisher{announced: s.announced}
}

func TestFlows_DeleteFlow_DropsTheFlowFromTheOwnersFavorites(t *testing.T) {
	db := setupFlowFileServiceTestDB(t)
	require.NoError(t, db.Exec(`
		CREATE TABLE containers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL DEFAULT 'primary',
			name TEXT NOT NULL DEFAULT '',
			image TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'created',
			local_id TEXT,
			local_dir TEXT,
			flow_id INTEGER NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	seedFlow(t, db, 42, 7)

	dropped := database.DeleteFavoriteFlowParams{}
	announced := false
	announcedTo := int64(0)

	c, w := newFlowFileTestContext(http.MethodDelete, "/flows/42", nil, []string{"flows.admin"}, 9, 42)

	NewFlowService(
		db,
		favoritesQuerier{dropped: &dropped},
		nil,
		&flowsCtrl{},
		favoritesSubs{announced: &announced, announcedTo: &announcedTo},
	).DeleteFlow(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Equal(t, database.DeleteFavoriteFlowParams{FlowID: 42, UserID: 7}, dropped,
		"the id of a deleted flow stays starred")
	require.True(t, announced,
		"favorites changed without SettingsUserUpdated, so a connected client keeps the dead id")
	require.Equal(t, int64(7), announcedTo,
		"SettingsUserUpdated announced to the caller instead of the owner")
}
