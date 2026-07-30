package services

import (
	"encoding/json"
	"net/http"
	"testing"

	"pentagi/pkg/server/models"

	"github.com/stretchr/testify/require"
)

func TestGetFlowsAllowsPendingTraceID(t *testing.T) {
	db := setupFlowFileServiceTestDB(t)
	require.NoError(t, db.Exec(`
		INSERT INTO flows (
			id, user_id, model, model_provider_name, model_provider_type,
			tool_call_id_template, trace_id
		) VALUES (1, 42, 'gpt', 'openai', 'openai', 'tcid', NULL)
	`).Error)

	c, w := newFlowFileTestContext(
		http.MethodGet,
		"/flows/?page=1&pageSize=5&type=init",
		nil,
		[]string{"flows.view"},
		42,
		0,
	)

	NewFlowService(db, nil, nil, nil).GetFlows(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Status string `json:"status"`
		Data   struct {
			Flows []models.Flow `json:"flows"`
			Total uint64        `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "success", resp.Status)
	require.Equal(t, uint64(1), resp.Data.Total)
	require.Len(t, resp.Data.Flows, 1)
	require.Nil(t, resp.Data.Flows[0].TraceID)
}
