package services

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/response"
	"pentagi/pkg/tools"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type AnalyticsService struct {
	db *gorm.DB
}

func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{
		db: db,
	}
}

// GetSystemUsage is a function to return system-wide analytics
// @Summary Retrieve system-wide analytics
// @Description Get comprehensive analytics for all user's flows including usage, toolcalls, and structural stats
// @Tags Usage
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.successResp{data=models.SystemUsageResponse} "analytics received successful"
// @Failure 403 {object} response.errorResp "getting analytics not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting analytics"
// @Router /usage [get]
func (s *AnalyticsService) GetSystemUsage(c *gin.Context) {
	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")

	if !slices.Contains(privs, "usage.view") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	var resp models.SystemUsageResponse

	// 1. Get total usage stats from msgchains
	var usageStats struct {
		TotalUsageIn       int64
		TotalUsageOut      int64
		TotalUsageCacheIn  int64
		TotalUsageCacheOut int64
		TotalUsageCostIn   float64
		TotalUsageCostOut  float64
	}

	err := s.db.Raw(`
		SELECT
			COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
			COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
			COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
			COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
			COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
			COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
		FROM msgchains mc
		LEFT JOIN subtasks s ON mc.subtask_id = s.id
		LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
		INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
		WHERE f.deleted_at IS NULL AND f.user_id = ?
	`, uid).Scan(&usageStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting total usage stats")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.UsageStatsTotal = &models.UsageStats{
		TotalUsageIn:       int(usageStats.TotalUsageIn),
		TotalUsageOut:      int(usageStats.TotalUsageOut),
		TotalUsageCacheIn:  int(usageStats.TotalUsageCacheIn),
		TotalUsageCacheOut: int(usageStats.TotalUsageCacheOut),
		TotalUsageCostIn:   usageStats.TotalUsageCostIn,
		TotalUsageCostOut:  usageStats.TotalUsageCostOut,
	}

	// 2. Get total toolcalls stats
	var toolcallsStats struct {
		TotalCount           int64
		TotalDurationSeconds float64
	}

	err = s.db.Raw(`
		SELECT
			COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
			COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds
		FROM toolcalls tc
		LEFT JOIN subtasks s ON tc.subtask_id = s.id
		LEFT JOIN tasks t ON s.task_id = t.id OR tc.task_id = t.id
		INNER JOIN flows f ON (tc.flow_id = f.id OR t.flow_id = f.id)
		WHERE f.deleted_at IS NULL AND f.user_id = ?
			AND (tc.task_id IS NULL OR t.id IS NOT NULL)
			AND (tc.subtask_id IS NULL OR s.id IS NOT NULL)
	`, uid).Scan(&toolcallsStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting total toolcalls stats")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.ToolcallsStatsTotal = &models.ToolcallsStats{
		TotalCount:           int(toolcallsStats.TotalCount),
		TotalDurationSeconds: toolcallsStats.TotalDurationSeconds,
	}

	// 3. Get flows stats
	var flowsStats struct {
		TotalFlowsCount      int64
		TotalTasksCount      int64
		TotalSubtasksCount   int64
		TotalAssistantsCount int64
	}

	err = s.db.Raw(`
		SELECT
			COALESCE(COUNT(DISTINCT f.id), 0)::bigint AS total_flows_count,
			COALESCE(COUNT(DISTINCT t.id), 0)::bigint AS total_tasks_count,
			COALESCE(COUNT(DISTINCT s.id), 0)::bigint AS total_subtasks_count,
			COALESCE(COUNT(DISTINCT a.id), 0)::bigint AS total_assistants_count
		FROM flows f
		LEFT JOIN tasks t ON f.id = t.flow_id
		LEFT JOIN subtasks s ON t.id = s.task_id
		LEFT JOIN assistants a ON f.id = a.flow_id AND a.deleted_at IS NULL
		WHERE f.user_id = ? AND f.deleted_at IS NULL
	`, uid).Scan(&flowsStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting flows stats")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.FlowsStatsTotal = &models.FlowsStats{
		TotalFlowsCount:      int(flowsStats.TotalFlowsCount),
		TotalTasksCount:      int(flowsStats.TotalTasksCount),
		TotalSubtasksCount:   int(flowsStats.TotalSubtasksCount),
		TotalAssistantsCount: int(flowsStats.TotalAssistantsCount),
	}

	// 4. Get usage stats by provider
	var providerStats []struct {
		ModelProvider      string
		TotalUsageIn       int64
		TotalUsageOut      int64
		TotalUsageCacheIn  int64
		TotalUsageCacheOut int64
		TotalUsageCostIn   float64
		TotalUsageCostOut  float64
	}

	err = s.db.Raw(`
		SELECT
			mc.model_provider,
			COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
			COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
			COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
			COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
			COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
			COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
		FROM msgchains mc
		LEFT JOIN subtasks s ON mc.subtask_id = s.id
		LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
		INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
		WHERE f.deleted_at IS NULL AND f.user_id = ?
		GROUP BY mc.model_provider
		ORDER BY mc.model_provider
	`, uid).Scan(&providerStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting usage stats by provider")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.UsageStatsByProvider = make([]models.ProviderUsageStats, 0, len(providerStats))
	for _, stat := range providerStats {
		resp.UsageStatsByProvider = append(resp.UsageStatsByProvider, models.ProviderUsageStats{
			Provider: stat.ModelProvider,
			Stats: &models.UsageStats{
				TotalUsageIn:       int(stat.TotalUsageIn),
				TotalUsageOut:      int(stat.TotalUsageOut),
				TotalUsageCacheIn:  int(stat.TotalUsageCacheIn),
				TotalUsageCacheOut: int(stat.TotalUsageCacheOut),
				TotalUsageCostIn:   stat.TotalUsageCostIn,
				TotalUsageCostOut:  stat.TotalUsageCostOut,
			},
		})
	}

	// 5. Get usage stats by model
	var modelStats []struct {
		Model              string
		ModelProvider      string
		TotalUsageIn       int64
		TotalUsageOut      int64
		TotalUsageCacheIn  int64
		TotalUsageCacheOut int64
		TotalUsageCostIn   float64
		TotalUsageCostOut  float64
	}

	err = s.db.Raw(`
		SELECT
			mc.model,
			mc.model_provider,
			COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
			COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
			COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
			COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
			COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
			COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
		FROM msgchains mc
		LEFT JOIN subtasks s ON mc.subtask_id = s.id
		LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
		INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
		WHERE f.deleted_at IS NULL AND f.user_id = ?
		GROUP BY mc.model, mc.model_provider
		ORDER BY mc.model, mc.model_provider
	`, uid).Scan(&modelStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting usage stats by model")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.UsageStatsByModel = make([]models.ModelUsageStats, 0, len(modelStats))
	for _, stat := range modelStats {
		resp.UsageStatsByModel = append(resp.UsageStatsByModel, models.ModelUsageStats{
			Model:    stat.Model,
			Provider: stat.ModelProvider,
			Stats: &models.UsageStats{
				TotalUsageIn:       int(stat.TotalUsageIn),
				TotalUsageOut:      int(stat.TotalUsageOut),
				TotalUsageCacheIn:  int(stat.TotalUsageCacheIn),
				TotalUsageCacheOut: int(stat.TotalUsageCacheOut),
				TotalUsageCostIn:   stat.TotalUsageCostIn,
				TotalUsageCostOut:  stat.TotalUsageCostOut,
			},
		})
	}

	// 6. Get usage stats by agent type
	var agentTypeStats []struct {
		Type               string
		TotalUsageIn       int64
		TotalUsageOut      int64
		TotalUsageCacheIn  int64
		TotalUsageCacheOut int64
		TotalUsageCostIn   float64
		TotalUsageCostOut  float64
	}

	err = s.db.Raw(`
		SELECT
			mc.type,
			COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
			COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
			COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
			COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
			COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
			COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
		FROM msgchains mc
		LEFT JOIN subtasks s ON mc.subtask_id = s.id
		LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
		INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
		WHERE f.deleted_at IS NULL AND f.user_id = ?
		GROUP BY mc.type
		ORDER BY mc.type
	`, uid).Scan(&agentTypeStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting usage stats by agent type")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.UsageStatsByAgentType = make([]models.AgentTypeUsageStats, 0, len(agentTypeStats))
	for _, stat := range agentTypeStats {
		resp.UsageStatsByAgentType = append(resp.UsageStatsByAgentType, models.AgentTypeUsageStats{
			AgentType: models.MsgchainType(stat.Type),
			Stats: &models.UsageStats{
				TotalUsageIn:       int(stat.TotalUsageIn),
				TotalUsageOut:      int(stat.TotalUsageOut),
				TotalUsageCacheIn:  int(stat.TotalUsageCacheIn),
				TotalUsageCacheOut: int(stat.TotalUsageCacheOut),
				TotalUsageCostIn:   stat.TotalUsageCostIn,
				TotalUsageCostOut:  stat.TotalUsageCostOut,
			},
		})
	}

	// 7. Get toolcalls stats by function
	var functionStats []struct {
		FunctionName         string
		TotalCount           int64
		TotalDurationSeconds float64
		AvgDurationSeconds   float64
	}

	err = s.db.Raw(`
		SELECT
			tc.name AS function_name,
			COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
			COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds,
			COALESCE(AVG(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE NULL END), 0.0)::double precision AS avg_duration_seconds
		FROM toolcalls tc
		LEFT JOIN subtasks s ON tc.subtask_id = s.id
		LEFT JOIN tasks t ON s.task_id = t.id OR tc.task_id = t.id
		INNER JOIN flows f ON (tc.flow_id = f.id OR t.flow_id = f.id)
		WHERE f.deleted_at IS NULL AND f.user_id = ?
		GROUP BY tc.name
		ORDER BY total_duration_seconds DESC
	`, uid).Scan(&functionStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting toolcalls stats by function")
		response.Error(c, response.ErrInternal, err)
		return
	}

	// Determine isAgent flag using tools package
	toolTypeMapping := tools.GetToolTypeMapping()
	resp.ToolcallsStatsByFunction = make([]models.FunctionToolcallsStats, 0, len(functionStats))
	for _, stat := range functionStats {
		isAgent := false
		if toolType, exists := toolTypeMapping[stat.FunctionName]; exists {
			isAgent = toolType == tools.AgentToolType || toolType == tools.StoreAgentResultToolType
		}

		resp.ToolcallsStatsByFunction = append(resp.ToolcallsStatsByFunction, models.FunctionToolcallsStats{
			FunctionName:         stat.FunctionName,
			IsAgent:              isAgent,
			TotalCount:           int(stat.TotalCount),
			TotalDurationSeconds: stat.TotalDurationSeconds,
			AvgDurationSeconds:   stat.AvgDurationSeconds,
		})
	}

	response.Success(c, http.StatusOK, resp)
}

// GetPeriodUsage is a function to return analytics for time period
// @Summary Retrieve analytics for specific time period
// @Description Get time-series analytics data for week, month, or quarter
// @Tags Usage
// @Produce json
// @Security BearerAuth
// @Param period path string true "period" Enums(week, month, quarter)
// @Success 200 {object} response.successResp{data=models.PeriodUsageResponse} "period analytics received successful"
// @Failure 400 {object} response.errorResp "invalid period parameter"
// @Failure 403 {object} response.errorResp "getting analytics not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting analytics"
// @Router /usage/{period} [get]
func (s *AnalyticsService) GetPeriodUsage(c *gin.Context) {
	period := c.Param("period")

	// Validate period
	var validPeriod models.UsageStatsPeriod
	var intervalDays int
	switch period {
	case "week":
		validPeriod = models.UsageStatsPeriodWeek
		intervalDays = 7
	case "month":
		validPeriod = models.UsageStatsPeriodMonth
		intervalDays = 30
	case "quarter":
		validPeriod = models.UsageStatsPeriodQuarter
		intervalDays = 90
	default:
		logger.FromContext(c).Errorf("invalid period parameter: %s", period)
		response.Error(c, response.ErrFlowsInvalidRequest, nil)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")

	if !slices.Contains(privs, "usage.view") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	var resp models.PeriodUsageResponse
	resp.Period = string(validPeriod)

	// 1. Get daily usage stats
	var dailyUsageStats []struct {
		Date               time.Time
		TotalUsageIn       int64
		TotalUsageOut      int64
		TotalUsageCacheIn  int64
		TotalUsageCacheOut int64
		TotalUsageCostIn   float64
		TotalUsageCostOut  float64
	}

	intervalSQL := fmt.Sprintf("NOW() - INTERVAL '%d days'", intervalDays)
	err := s.db.Raw(fmt.Sprintf(`
		SELECT
			DATE(mc.created_at) AS date,
			COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
			COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
			COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
			COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
			COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
			COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
		FROM msgchains mc
		LEFT JOIN subtasks s ON mc.subtask_id = s.id
		LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
		INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
		WHERE mc.created_at >= %s AND f.deleted_at IS NULL AND f.user_id = ?
		GROUP BY DATE(mc.created_at)
		ORDER BY date DESC
	`, intervalSQL), uid).Scan(&dailyUsageStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting daily usage stats")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.UsageStatsByPeriod = make([]models.DailyUsageStats, 0, len(dailyUsageStats))
	for _, stat := range dailyUsageStats {
		resp.UsageStatsByPeriod = append(resp.UsageStatsByPeriod, models.DailyUsageStats{
			Date: stat.Date,
			Stats: &models.UsageStats{
				TotalUsageIn:       int(stat.TotalUsageIn),
				TotalUsageOut:      int(stat.TotalUsageOut),
				TotalUsageCacheIn:  int(stat.TotalUsageCacheIn),
				TotalUsageCacheOut: int(stat.TotalUsageCacheOut),
				TotalUsageCostIn:   stat.TotalUsageCostIn,
				TotalUsageCostOut:  stat.TotalUsageCostOut,
			},
		})
	}

	// 2. Get daily toolcalls stats
	var dailyToolcallsStats []struct {
		Date                 time.Time
		TotalCount           int64
		TotalDurationSeconds float64
	}

	err = s.db.Raw(fmt.Sprintf(`
		SELECT
			DATE(tc.created_at) AS date,
			COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
			COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds
		FROM toolcalls tc
		LEFT JOIN subtasks s ON tc.subtask_id = s.id
		LEFT JOIN tasks t ON s.task_id = t.id OR tc.task_id = t.id
		INNER JOIN flows f ON (tc.flow_id = f.id OR t.flow_id = f.id)
		WHERE tc.created_at >= %s AND f.deleted_at IS NULL AND f.user_id = ?
		GROUP BY DATE(tc.created_at)
		ORDER BY date DESC
	`, intervalSQL), uid).Scan(&dailyToolcallsStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting daily toolcalls stats")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.ToolcallsStatsByPeriod = make([]models.DailyToolcallsStats, 0, len(dailyToolcallsStats))
	for _, stat := range dailyToolcallsStats {
		resp.ToolcallsStatsByPeriod = append(resp.ToolcallsStatsByPeriod, models.DailyToolcallsStats{
			Date: stat.Date,
			Stats: &models.ToolcallsStats{
				TotalCount:           int(stat.TotalCount),
				TotalDurationSeconds: stat.TotalDurationSeconds,
			},
		})
	}

	// 3. Get daily flows stats
	var dailyFlowsStats []struct {
		Date                 time.Time
		TotalFlowsCount      int64
		TotalTasksCount      int64
		TotalSubtasksCount   int64
		TotalAssistantsCount int64
	}

	err = s.db.Raw(fmt.Sprintf(`
		SELECT
			DATE(f.created_at) AS date,
			COALESCE(COUNT(DISTINCT f.id), 0)::bigint AS total_flows_count,
			COALESCE(COUNT(DISTINCT t.id), 0)::bigint AS total_tasks_count,
			COALESCE(COUNT(DISTINCT s.id), 0)::bigint AS total_subtasks_count,
			COALESCE(COUNT(DISTINCT a.id), 0)::bigint AS total_assistants_count
		FROM flows f
		LEFT JOIN tasks t ON f.id = t.flow_id
		LEFT JOIN subtasks s ON t.id = s.task_id
		LEFT JOIN assistants a ON f.id = a.flow_id AND a.deleted_at IS NULL
		WHERE f.created_at >= %s AND f.deleted_at IS NULL AND f.user_id = ?
		GROUP BY DATE(f.created_at)
		ORDER BY date DESC
	`, intervalSQL), uid).Scan(&dailyFlowsStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting daily flows stats")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.FlowsStatsByPeriod = make([]models.DailyFlowsStats, 0, len(dailyFlowsStats))
	for _, stat := range dailyFlowsStats {
		resp.FlowsStatsByPeriod = append(resp.FlowsStatsByPeriod, models.DailyFlowsStats{
			Date: stat.Date,
			Stats: &models.FlowsStats{
				TotalFlowsCount:      int(stat.TotalFlowsCount),
				TotalTasksCount:      int(stat.TotalTasksCount),
				TotalSubtasksCount:   int(stat.TotalSubtasksCount),
				TotalAssistantsCount: int(stat.TotalAssistantsCount),
			},
		})
	}

	// 4. Get flows execution stats for the period
	// This is complex and requires using converter logic from GraphQL resolvers
	// We'll get flows for the period and then build execution stats for each
	var flowsForPeriod []struct {
		ID    int64
		Title string
	}

	err = s.db.Raw(fmt.Sprintf(`
		SELECT id, title
		FROM flows
		WHERE created_at >= %s AND deleted_at IS NULL AND user_id = ?
		ORDER BY created_at DESC
	`, intervalSQL), uid).Scan(&flowsForPeriod).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting flows for period")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.FlowsExecutionStatsByPeriod = make([]models.FlowExecutionStats, 0, len(flowsForPeriod))

	// For each flow, build full execution stats with tasks/subtasks hierarchy
	for _, flow := range flowsForPeriod {
		// Get tasks for this flow
		var tasks []struct {
			ID        int64
			Title     string
			CreatedAt time.Time
			UpdatedAt time.Time
		}

		err = s.db.Raw(`
			SELECT id, title, created_at, updated_at
			FROM tasks
			WHERE flow_id = ?
			ORDER BY id ASC
		`, flow.ID).Scan(&tasks).Error

		if err != nil {
			logger.FromContext(c).WithError(err).Errorf("error getting tasks for flow %d", flow.ID)
			continue
		}

		// Collect task IDs
		taskIDs := make([]int64, len(tasks))
		for i, task := range tasks {
			taskIDs[i] = task.ID
		}

		// Get subtasks for all tasks
		var subtasks []struct {
			ID        int64
			TaskID    int64
			Title     string
			Status    string
			CreatedAt time.Time
			UpdatedAt time.Time
		}

		if len(taskIDs) > 0 {
			// PostgreSQL array parameter requires special handling with pq.Array
			// Using IN clause instead for GORM compatibility
			err = s.db.Raw(`
				SELECT id, task_id, title, status, created_at, updated_at
				FROM subtasks
				WHERE task_id IN (?)
				ORDER BY id ASC
			`, taskIDs).Scan(&subtasks).Error

			if err != nil {
				logger.FromContext(c).WithError(err).Errorf("error getting subtasks for flow %d", flow.ID)
				continue
			}
		}

		// Get toolcalls for the flow
		var toolcalls []struct {
			ID              int64
			Status          string
			FlowID          int64
			TaskID          *int64
			SubtaskID       *int64
			DurationSeconds float64
			CreatedAt       time.Time
			UpdatedAt       time.Time
		}

		err = s.db.Raw(`
			SELECT tc.id, tc.status, tc.flow_id, tc.task_id, tc.subtask_id, tc.duration_seconds, tc.created_at, tc.updated_at
			FROM toolcalls tc
			LEFT JOIN tasks t ON tc.task_id = t.id
			LEFT JOIN subtasks s ON tc.subtask_id = s.id
			INNER JOIN flows f ON tc.flow_id = f.id
			WHERE tc.flow_id = ? AND f.deleted_at IS NULL
				AND (tc.task_id IS NULL OR t.id IS NOT NULL)
				AND (tc.subtask_id IS NULL OR s.id IS NOT NULL)
			ORDER BY tc.created_at ASC
		`, flow.ID).Scan(&toolcalls).Error

		if err != nil {
			logger.FromContext(c).WithError(err).Errorf("error getting toolcalls for flow %d", flow.ID)
			continue
		}

		// Get assistants count
		var assistantsCountResult struct {
			TotalAssistantsCount int64
		}
		err = s.db.Raw(`
			SELECT COALESCE(COUNT(id), 0)::bigint AS total_assistants_count
			FROM assistants
			WHERE flow_id = ? AND deleted_at IS NULL
		`, flow.ID).Scan(&assistantsCountResult).Error

		if err != nil {
			logger.FromContext(c).WithError(err).Errorf("error getting assistants count for flow %d", flow.ID)
			continue
		}
		assistantsCount := assistantsCountResult.TotalAssistantsCount

		// Build task execution stats
		taskStats := make([]models.TaskExecutionStats, 0, len(tasks))
		for _, task := range tasks {
			// Get subtasks for this task
			var taskSubtasks []struct {
				ID        int64
				Title     string
				Status    string
				CreatedAt time.Time
				UpdatedAt time.Time
			}

			for _, st := range subtasks {
				if st.TaskID == task.ID {
					taskSubtasks = append(taskSubtasks, struct {
						ID        int64
						Title     string
						Status    string
						CreatedAt time.Time
						UpdatedAt time.Time
					}{
						ID:        st.ID,
						Title:     st.Title,
						Status:    st.Status,
						CreatedAt: st.CreatedAt,
						UpdatedAt: st.UpdatedAt,
					})
				}
			}

			// Build subtask execution stats
			subtaskStats := make([]models.SubtaskExecutionStats, 0, len(taskSubtasks))
			var totalTaskDuration float64
			var totalTaskToolcalls int

			for _, subtask := range taskSubtasks {
				// Calculate subtask duration (linear time for executed subtasks)
				var subtaskDuration float64
				if subtask.Status != "created" && subtask.Status != "waiting" {
					if subtask.Status == "running" {
						subtaskDuration = time.Since(subtask.CreatedAt).Seconds()
					} else {
						subtaskDuration = subtask.UpdatedAt.Sub(subtask.CreatedAt).Seconds()
					}
				}

				// Count finished toolcalls for this subtask
				var subtaskToolcallsCount int
				for _, tc := range toolcalls {
					if tc.SubtaskID != nil && *tc.SubtaskID == subtask.ID {
						if tc.Status == "finished" || tc.Status == "failed" {
							subtaskToolcallsCount++
						}
					}
				}

				subtaskStats = append(subtaskStats, models.SubtaskExecutionStats{
					SubtaskID:            subtask.ID,
					SubtaskTitle:         subtask.Title,
					TotalDurationSeconds: subtaskDuration,
					TotalToolcallsCount:  subtaskToolcallsCount,
				})

				totalTaskDuration += subtaskDuration
				totalTaskToolcalls += subtaskToolcallsCount
			}

			// Count task-level toolcalls
			for _, tc := range toolcalls {
				if tc.TaskID != nil && *tc.TaskID == task.ID && tc.SubtaskID == nil {
					if tc.Status == "finished" || tc.Status == "failed" {
						totalTaskToolcalls++
					}
				}
			}

			taskStats = append(taskStats, models.TaskExecutionStats{
				TaskID:               task.ID,
				TaskTitle:            task.Title,
				TotalDurationSeconds: totalTaskDuration,
				TotalToolcallsCount:  totalTaskToolcalls,
				Subtasks:             subtaskStats,
			})
		}

		// Calculate total flow duration and toolcalls
		var totalFlowDuration float64
		var totalFlowToolcalls int
		for _, ts := range taskStats {
			totalFlowDuration += ts.TotalDurationSeconds
			totalFlowToolcalls += ts.TotalToolcallsCount
		}

		// Add flow-level toolcalls (without task binding)
		for _, tc := range toolcalls {
			if tc.TaskID == nil && tc.SubtaskID == nil {
				if tc.Status == "finished" || tc.Status == "failed" {
					totalFlowToolcalls++
				}
			}
		}

		resp.FlowsExecutionStatsByPeriod = append(resp.FlowsExecutionStatsByPeriod, models.FlowExecutionStats{
			FlowID:               flow.ID,
			FlowTitle:            flow.Title,
			TotalDurationSeconds: totalFlowDuration,
			TotalToolcallsCount:  totalFlowToolcalls,
			TotalAssistantsCount: int(assistantsCount),
			Tasks:                taskStats,
		})
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowUsage is a function to return analytics for specific flow
// @Summary Retrieve analytics for specific flow
// @Description Get comprehensive analytics for a single flow including all breakdowns
// @Tags Flows, Usage
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Success 200 {object} response.successResp{data=models.FlowUsageResponse} "flow analytics received successful"
// @Failure 400 {object} response.errorResp "invalid flow id"
// @Failure 403 {object} response.errorResp "getting flow analytics not permitted"
// @Failure 404 {object} response.errorResp "flow not found"
// @Failure 500 {object} response.errorResp "internal error on getting flow analytics"
// @Router /flows/{flowID}/usage [get]
func (s *AnalyticsService) GetFlowUsage(c *gin.Context) {
	flowID, err := strconv.ParseUint(c.Param("flowID"), 10, 64)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrFlowsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")

	// Check permissions
	var hasPermission bool
	if slices.Contains(privs, "usage.admin") {
		hasPermission = true
	} else if slices.Contains(privs, "usage.view") {
		hasPermission = true
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	// Check flow ownership without loading the full object (to avoid JSON field scanning issues)
	var count int64
	var checkQuery *gorm.DB
	if slices.Contains(privs, "usage.admin") {
		checkQuery = s.db.Model(&models.Flow{}).Where("id = ? AND deleted_at IS NULL", flowID)
	} else {
		checkQuery = s.db.Model(&models.Flow{}).Where("id = ? AND user_id = ? AND deleted_at IS NULL", flowID, uid)
	}

	if err := checkQuery.Count(&count).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error checking flow ownership: %d", flowID)
		response.Error(c, response.ErrInternal, err)
		return
	}

	if count == 0 {
		logger.FromContext(c).Errorf("flow not found or access denied: %d", flowID)
		response.Error(c, response.ErrFlowsNotFound, nil)
		return
	}

	if !hasPermission {
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	var resp models.FlowUsageResponse
	resp.FlowID = int64(flowID)

	// 1. Get usage stats for this flow
	var usageStats struct {
		TotalUsageIn       int64
		TotalUsageOut      int64
		TotalUsageCacheIn  int64
		TotalUsageCacheOut int64
		TotalUsageCostIn   float64
		TotalUsageCostOut  float64
	}

	err = s.db.Raw(`
		SELECT
			COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
			COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
			COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
			COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
			COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
			COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
		FROM msgchains mc
		LEFT JOIN subtasks s ON mc.subtask_id = s.id
		LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
		INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
		WHERE (mc.flow_id = ? OR t.flow_id = ?) AND f.deleted_at IS NULL
	`, flowID, flowID).Scan(&usageStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting usage stats for flow")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.UsageStatsByFlow = &models.UsageStats{
		TotalUsageIn:       int(usageStats.TotalUsageIn),
		TotalUsageOut:      int(usageStats.TotalUsageOut),
		TotalUsageCacheIn:  int(usageStats.TotalUsageCacheIn),
		TotalUsageCacheOut: int(usageStats.TotalUsageCacheOut),
		TotalUsageCostIn:   usageStats.TotalUsageCostIn,
		TotalUsageCostOut:  usageStats.TotalUsageCostOut,
	}

	// 2. Get usage stats by agent type for this flow
	var agentTypeStats []struct {
		Type               string
		TotalUsageIn       int64
		TotalUsageOut      int64
		TotalUsageCacheIn  int64
		TotalUsageCacheOut int64
		TotalUsageCostIn   float64
		TotalUsageCostOut  float64
	}

	err = s.db.Raw(`
		SELECT
			mc.type,
			COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
			COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
			COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
			COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
			COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
			COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
		FROM msgchains mc
		LEFT JOIN subtasks s ON mc.subtask_id = s.id
		LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
		INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
		WHERE (mc.flow_id = ? OR t.flow_id = ?) AND f.deleted_at IS NULL
		GROUP BY mc.type
		ORDER BY mc.type
	`, flowID, flowID).Scan(&agentTypeStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting usage stats by agent type for flow")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.UsageStatsByAgentTypeForFlow = make([]models.AgentTypeUsageStats, 0, len(agentTypeStats))
	for _, stat := range agentTypeStats {
		resp.UsageStatsByAgentTypeForFlow = append(resp.UsageStatsByAgentTypeForFlow, models.AgentTypeUsageStats{
			AgentType: models.MsgchainType(stat.Type),
			Stats: &models.UsageStats{
				TotalUsageIn:       int(stat.TotalUsageIn),
				TotalUsageOut:      int(stat.TotalUsageOut),
				TotalUsageCacheIn:  int(stat.TotalUsageCacheIn),
				TotalUsageCacheOut: int(stat.TotalUsageCacheOut),
				TotalUsageCostIn:   stat.TotalUsageCostIn,
				TotalUsageCostOut:  stat.TotalUsageCostOut,
			},
		})
	}

	// 3. Get toolcalls stats for this flow
	var toolcallsStats struct {
		TotalCount           int64
		TotalDurationSeconds float64
	}

	err = s.db.Raw(`
		SELECT
			COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
			COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds
		FROM toolcalls tc
		LEFT JOIN tasks t ON tc.task_id = t.id
		LEFT JOIN subtasks s ON tc.subtask_id = s.id
		INNER JOIN flows f ON tc.flow_id = f.id
		WHERE tc.flow_id = ? AND f.deleted_at IS NULL
			AND (tc.task_id IS NULL OR t.id IS NOT NULL)
			AND (tc.subtask_id IS NULL OR s.id IS NOT NULL)
	`, flowID).Scan(&toolcallsStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting toolcalls stats for flow")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.ToolcallsStatsByFlow = &models.ToolcallsStats{
		TotalCount:           int(toolcallsStats.TotalCount),
		TotalDurationSeconds: toolcallsStats.TotalDurationSeconds,
	}

	// 4. Get toolcalls stats by function for this flow
	var functionStats []struct {
		FunctionName         string
		TotalCount           int64
		TotalDurationSeconds float64
		AvgDurationSeconds   float64
	}

	err = s.db.Raw(`
		SELECT
			tc.name AS function_name,
			COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
			COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds,
			COALESCE(AVG(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE NULL END), 0.0)::double precision AS avg_duration_seconds
		FROM toolcalls tc
		LEFT JOIN subtasks s ON tc.subtask_id = s.id
		LEFT JOIN tasks t ON s.task_id = t.id OR tc.task_id = t.id
		INNER JOIN flows f ON (tc.flow_id = f.id OR t.flow_id = f.id)
		WHERE (tc.flow_id = ? OR t.flow_id = ?) AND f.deleted_at IS NULL
		GROUP BY tc.name
		ORDER BY total_duration_seconds DESC
	`, flowID, flowID).Scan(&functionStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting toolcalls stats by function for flow")
		response.Error(c, response.ErrInternal, err)
		return
	}

	toolTypeMapping := tools.GetToolTypeMapping()
	resp.ToolcallsStatsByFunctionForFlow = make([]models.FunctionToolcallsStats, 0, len(functionStats))
	for _, stat := range functionStats {
		isAgent := false
		if toolType, exists := toolTypeMapping[stat.FunctionName]; exists {
			isAgent = toolType == tools.AgentToolType || toolType == tools.StoreAgentResultToolType
		}

		resp.ToolcallsStatsByFunctionForFlow = append(resp.ToolcallsStatsByFunctionForFlow, models.FunctionToolcallsStats{
			FunctionName:         stat.FunctionName,
			IsAgent:              isAgent,
			TotalCount:           int(stat.TotalCount),
			TotalDurationSeconds: stat.TotalDurationSeconds,
			AvgDurationSeconds:   stat.AvgDurationSeconds,
		})
	}

	// 5. Get flow structure stats
	var flowStats struct {
		TotalTasksCount      int64
		TotalSubtasksCount   int64
		TotalAssistantsCount int64
	}

	err = s.db.Raw(`
		SELECT
			COALESCE(COUNT(DISTINCT t.id), 0)::bigint AS total_tasks_count,
			COALESCE(COUNT(DISTINCT s.id), 0)::bigint AS total_subtasks_count,
			COALESCE(COUNT(DISTINCT a.id), 0)::bigint AS total_assistants_count
		FROM flows f
		LEFT JOIN tasks t ON f.id = t.flow_id
		LEFT JOIN subtasks s ON t.id = s.task_id
		LEFT JOIN assistants a ON f.id = a.flow_id AND a.deleted_at IS NULL
		WHERE f.id = ? AND f.deleted_at IS NULL
	`, flowID).Scan(&flowStats).Error

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting flow stats")
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.FlowStatsByFlow = &models.FlowStats{
		TotalTasksCount:      int(flowStats.TotalTasksCount),
		TotalSubtasksCount:   int(flowStats.TotalSubtasksCount),
		TotalAssistantsCount: int(flowStats.TotalAssistantsCount),
	}

	response.Success(c, http.StatusOK, resp)
}
