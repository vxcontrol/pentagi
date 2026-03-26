package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

// UsageStatsPeriod represents time period enum for analytics
type UsageStatsPeriod string

const (
	UsageStatsPeriodWeek    UsageStatsPeriod = "week"
	UsageStatsPeriodMonth   UsageStatsPeriod = "month"
	UsageStatsPeriodQuarter UsageStatsPeriod = "quarter"
)

func (p UsageStatsPeriod) String() string {
	return string(p)
}

// Valid is function to control input/output data
func (p UsageStatsPeriod) Valid() error {
	switch p {
	case UsageStatsPeriodWeek,
		UsageStatsPeriodMonth,
		UsageStatsPeriodQuarter:
		return nil
	default:
		return fmt.Errorf("invalid UsageStatsPeriod: %s", p)
	}
}

// Validate is function to use callback to control input/output data
func (p UsageStatsPeriod) Validate(db *gorm.DB) {
	if err := p.Valid(); err != nil {
		db.AddError(err)
	}
}

// ==================== Basic Statistics Structures ====================

// UsageStats represents token usage statistics
// nolint:lll
type UsageStats struct {
	TotalUsageIn       int     `json:"total_usage_in" validate:"min=0"`
	TotalUsageOut      int     `json:"total_usage_out" validate:"min=0"`
	TotalUsageCacheIn  int     `json:"total_usage_cache_in" validate:"min=0"`
	TotalUsageCacheOut int     `json:"total_usage_cache_out" validate:"min=0"`
	TotalUsageCostIn   float64 `json:"total_usage_cost_in" validate:"min=0"`
	TotalUsageCostOut  float64 `json:"total_usage_cost_out" validate:"min=0"`
}

// Valid is function to control input/output data
func (u UsageStats) Valid() error {
	return validate.Struct(u)
}

// Validate is function to use callback to control input/output data
func (u UsageStats) Validate(db *gorm.DB) {
	if err := u.Valid(); err != nil {
		db.AddError(err)
	}
}

// ToolcallsStats represents toolcalls statistics
// nolint:lll
type ToolcallsStats struct {
	TotalCount           int     `json:"total_count" validate:"min=0"`
	TotalDurationSeconds float64 `json:"total_duration_seconds" validate:"min=0"`
}

// Valid is function to control input/output data
func (t ToolcallsStats) Valid() error {
	return validate.Struct(t)
}

// Validate is function to use callback to control input/output data
func (t ToolcallsStats) Validate(db *gorm.DB) {
	if err := t.Valid(); err != nil {
		db.AddError(err)
	}
}

// FlowsStats represents flows/tasks/subtasks counts
// nolint:lll
type FlowsStats struct {
	TotalFlowsCount      int `json:"total_flows_count" validate:"min=0"`
	TotalTasksCount      int `json:"total_tasks_count" validate:"min=0"`
	TotalSubtasksCount   int `json:"total_subtasks_count" validate:"min=0"`
	TotalAssistantsCount int `json:"total_assistants_count" validate:"min=0"`
}

// Valid is function to control input/output data
func (f FlowsStats) Valid() error {
	return validate.Struct(f)
}

// Validate is function to use callback to control input/output data
func (f FlowsStats) Validate(db *gorm.DB) {
	if err := f.Valid(); err != nil {
		db.AddError(err)
	}
}

// FlowStats represents single flow statistics
// nolint:lll
type FlowStats struct {
	TotalTasksCount      int `json:"total_tasks_count" validate:"min=0"`
	TotalSubtasksCount   int `json:"total_subtasks_count" validate:"min=0"`
	TotalAssistantsCount int `json:"total_assistants_count" validate:"min=0"`
}

// Valid is function to control input/output data
func (f FlowStats) Valid() error {
	return validate.Struct(f)
}

// Validate is function to use callback to control input/output data
func (f FlowStats) Validate(db *gorm.DB) {
	if err := f.Valid(); err != nil {
		db.AddError(err)
	}
}

// ==================== Time-series Statistics ====================

// DailyUsageStats for time-series usage data
// nolint:lll
type DailyUsageStats struct {
	Date  time.Time   `json:"date" validate:"required"`
	Stats *UsageStats `json:"stats" validate:"required"`
}

// Valid is function to control input/output data
func (d DailyUsageStats) Valid() error {
	if err := validate.Struct(d); err != nil {
		return err
	}
	if d.Stats != nil {
		return d.Stats.Valid()
	}
	return nil
}

// Validate is function to use callback to control input/output data
func (d DailyUsageStats) Validate(db *gorm.DB) {
	if err := d.Valid(); err != nil {
		db.AddError(err)
	}
}

// DailyToolcallsStats for time-series toolcalls data
// nolint:lll
type DailyToolcallsStats struct {
	Date  time.Time       `json:"date" validate:"required"`
	Stats *ToolcallsStats `json:"stats" validate:"required"`
}

// Valid is function to control input/output data
func (d DailyToolcallsStats) Valid() error {
	if err := validate.Struct(d); err != nil {
		return err
	}
	if d.Stats != nil {
		return d.Stats.Valid()
	}
	return nil
}

// Validate is function to use callback to control input/output data
func (d DailyToolcallsStats) Validate(db *gorm.DB) {
	if err := d.Valid(); err != nil {
		db.AddError(err)
	}
}

// DailyFlowsStats for time-series flows data
// nolint:lll
type DailyFlowsStats struct {
	Date  time.Time   `json:"date" validate:"required"`
	Stats *FlowsStats `json:"stats" validate:"required"`
}

// Valid is function to control input/output data
func (d DailyFlowsStats) Valid() error {
	if err := validate.Struct(d); err != nil {
		return err
	}
	if d.Stats != nil {
		return d.Stats.Valid()
	}
	return nil
}

// Validate is function to use callback to control input/output data
func (d DailyFlowsStats) Validate(db *gorm.DB) {
	if err := d.Valid(); err != nil {
		db.AddError(err)
	}
}

// ==================== Grouped Statistics ====================

// ProviderUsageStats for provider-specific usage statistics
// nolint:lll
type ProviderUsageStats struct {
	Provider string      `json:"provider" validate:"required"`
	Stats    *UsageStats `json:"stats" validate:"required"`
}

// Valid is function to control input/output data
func (p ProviderUsageStats) Valid() error {
	if err := validate.Struct(p); err != nil {
		return err
	}
	if p.Stats != nil {
		return p.Stats.Valid()
	}
	return nil
}

// Validate is function to use callback to control input/output data
func (p ProviderUsageStats) Validate(db *gorm.DB) {
	if err := p.Valid(); err != nil {
		db.AddError(err)
	}
}

// ModelUsageStats for model-specific usage statistics
// nolint:lll
type ModelUsageStats struct {
	Model    string      `json:"model" validate:"required"`
	Provider string      `json:"provider" validate:"required"`
	Stats    *UsageStats `json:"stats" validate:"required"`
}

// Valid is function to control input/output data
func (m ModelUsageStats) Valid() error {
	if err := validate.Struct(m); err != nil {
		return err
	}
	if m.Stats != nil {
		return m.Stats.Valid()
	}
	return nil
}

// Validate is function to use callback to control input/output data
func (m ModelUsageStats) Validate(db *gorm.DB) {
	if err := m.Valid(); err != nil {
		db.AddError(err)
	}
}

// AgentTypeUsageStats for agent type usage statistics
// nolint:lll
type AgentTypeUsageStats struct {
	AgentType MsgchainType `json:"agent_type" validate:"valid,required"`
	Stats     *UsageStats  `json:"stats" validate:"required"`
}

// Valid is function to control input/output data
func (a AgentTypeUsageStats) Valid() error {
	if err := validate.Struct(a); err != nil {
		return err
	}
	if a.Stats != nil {
		return a.Stats.Valid()
	}
	return nil
}

// Validate is function to use callback to control input/output data
func (a AgentTypeUsageStats) Validate(db *gorm.DB) {
	if err := a.Valid(); err != nil {
		db.AddError(err)
	}
}

// FunctionToolcallsStats for function-specific toolcalls statistics
// nolint:lll
type FunctionToolcallsStats struct {
	FunctionName         string  `json:"function_name" validate:"required"`
	IsAgent              bool    `json:"is_agent"`
	TotalCount           int     `json:"total_count" validate:"min=0"`
	TotalDurationSeconds float64 `json:"total_duration_seconds" validate:"min=0"`
	AvgDurationSeconds   float64 `json:"avg_duration_seconds" validate:"min=0"`
}

// Valid is function to control input/output data
func (f FunctionToolcallsStats) Valid() error {
	return validate.Struct(f)
}

// Validate is function to use callback to control input/output data
func (f FunctionToolcallsStats) Validate(db *gorm.DB) {
	if err := f.Valid(); err != nil {
		db.AddError(err)
	}
}

// ==================== Execution Statistics ====================

// SubtaskExecutionStats represents execution statistics for a subtask
// nolint:lll
type SubtaskExecutionStats struct {
	SubtaskID            int64   `json:"subtask_id" validate:"min=0"`
	SubtaskTitle         string  `json:"subtask_title" validate:"required"`
	TotalDurationSeconds float64 `json:"total_duration_seconds" validate:"min=0"`
	TotalToolcallsCount  int     `json:"total_toolcalls_count" validate:"min=0"`
}

// Valid is function to control input/output data
func (s SubtaskExecutionStats) Valid() error {
	return validate.Struct(s)
}

// Validate is function to use callback to control input/output data
func (s SubtaskExecutionStats) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// TaskExecutionStats represents execution statistics for a task
// nolint:lll
type TaskExecutionStats struct {
	TaskID               int64                   `json:"task_id" validate:"min=0"`
	TaskTitle            string                  `json:"task_title" validate:"required"`
	TotalDurationSeconds float64                 `json:"total_duration_seconds" validate:"min=0"`
	TotalToolcallsCount  int                     `json:"total_toolcalls_count" validate:"min=0"`
	Subtasks             []SubtaskExecutionStats `json:"subtasks" validate:"omitempty"`
}

// Valid is function to control input/output data
func (t TaskExecutionStats) Valid() error {
	if err := validate.Struct(t); err != nil {
		return err
	}
	for i := range t.Subtasks {
		if err := t.Subtasks[i].Valid(); err != nil {
			return err
		}
	}
	return nil
}

// Validate is function to use callback to control input/output data
func (t TaskExecutionStats) Validate(db *gorm.DB) {
	if err := t.Valid(); err != nil {
		db.AddError(err)
	}
}

// FlowExecutionStats represents execution statistics for a flow
// nolint:lll
type FlowExecutionStats struct {
	FlowID               int64                `json:"flow_id" validate:"min=0"`
	FlowTitle            string               `json:"flow_title" validate:"required"`
	TotalDurationSeconds float64              `json:"total_duration_seconds" validate:"min=0"`
	TotalToolcallsCount  int                  `json:"total_toolcalls_count" validate:"min=0"`
	TotalAssistantsCount int                  `json:"total_assistants_count" validate:"min=0"`
	Tasks                []TaskExecutionStats `json:"tasks" validate:"omitempty"`
}

// Valid is function to control input/output data
func (f FlowExecutionStats) Valid() error {
	if err := validate.Struct(f); err != nil {
		return err
	}
	for i := range f.Tasks {
		if err := f.Tasks[i].Valid(); err != nil {
			return err
		}
	}
	return nil
}

// Validate is function to use callback to control input/output data
func (f FlowExecutionStats) Validate(db *gorm.DB) {
	if err := f.Valid(); err != nil {
		db.AddError(err)
	}
}

// ==================== Aggregated Response Models ====================

// SystemUsageResponse represents system-wide analytics response
// nolint:lll
type SystemUsageResponse struct {
	UsageStatsTotal          *UsageStats              `json:"usage_stats_total" validate:"required"`
	ToolcallsStatsTotal      *ToolcallsStats          `json:"toolcalls_stats_total" validate:"required"`
	FlowsStatsTotal          *FlowsStats              `json:"flows_stats_total" validate:"required"`
	UsageStatsByProvider     []ProviderUsageStats     `json:"usage_stats_by_provider" validate:"omitempty"`
	UsageStatsByModel        []ModelUsageStats        `json:"usage_stats_by_model" validate:"omitempty"`
	UsageStatsByAgentType    []AgentTypeUsageStats    `json:"usage_stats_by_agent_type" validate:"omitempty"`
	ToolcallsStatsByFunction []FunctionToolcallsStats `json:"toolcalls_stats_by_function" validate:"omitempty"`
}

// Valid is function to control input/output data
func (s SystemUsageResponse) Valid() error {
	if err := validate.Struct(s); err != nil {
		return err
	}
	if s.UsageStatsTotal != nil {
		if err := s.UsageStatsTotal.Valid(); err != nil {
			return err
		}
	}
	if s.ToolcallsStatsTotal != nil {
		if err := s.ToolcallsStatsTotal.Valid(); err != nil {
			return err
		}
	}
	if s.FlowsStatsTotal != nil {
		if err := s.FlowsStatsTotal.Valid(); err != nil {
			return err
		}
	}
	for i := range s.UsageStatsByProvider {
		if err := s.UsageStatsByProvider[i].Valid(); err != nil {
			return err
		}
	}
	for i := range s.UsageStatsByModel {
		if err := s.UsageStatsByModel[i].Valid(); err != nil {
			return err
		}
	}
	for i := range s.UsageStatsByAgentType {
		if err := s.UsageStatsByAgentType[i].Valid(); err != nil {
			return err
		}
	}
	for i := range s.ToolcallsStatsByFunction {
		if err := s.ToolcallsStatsByFunction[i].Valid(); err != nil {
			return err
		}
	}
	return nil
}

// Validate is function to use callback to control input/output data
func (s SystemUsageResponse) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// PeriodUsageResponse represents period-based analytics response
// nolint:lll
type PeriodUsageResponse struct {
	Period                      string                `json:"period" validate:"required"`
	UsageStatsByPeriod          []DailyUsageStats     `json:"usage_stats_by_period" validate:"omitempty"`
	ToolcallsStatsByPeriod      []DailyToolcallsStats `json:"toolcalls_stats_by_period" validate:"omitempty"`
	FlowsStatsByPeriod          []DailyFlowsStats     `json:"flows_stats_by_period" validate:"omitempty"`
	FlowsExecutionStatsByPeriod []FlowExecutionStats  `json:"flows_execution_stats_by_period" validate:"omitempty"`
}

// Valid is function to control input/output data
func (p PeriodUsageResponse) Valid() error {
	if err := validate.Struct(p); err != nil {
		return err
	}
	for i := range p.UsageStatsByPeriod {
		if err := p.UsageStatsByPeriod[i].Valid(); err != nil {
			return err
		}
	}
	for i := range p.ToolcallsStatsByPeriod {
		if err := p.ToolcallsStatsByPeriod[i].Valid(); err != nil {
			return err
		}
	}
	for i := range p.FlowsStatsByPeriod {
		if err := p.FlowsStatsByPeriod[i].Valid(); err != nil {
			return err
		}
	}
	for i := range p.FlowsExecutionStatsByPeriod {
		if err := p.FlowsExecutionStatsByPeriod[i].Valid(); err != nil {
			return err
		}
	}
	return nil
}

// Validate is function to use callback to control input/output data
func (p PeriodUsageResponse) Validate(db *gorm.DB) {
	if err := p.Valid(); err != nil {
		db.AddError(err)
	}
}

// FlowUsageResponse represents flow-specific analytics response
// nolint:lll
type FlowUsageResponse struct {
	FlowID                          int64                    `json:"flow_id" validate:"min=0"`
	UsageStatsByFlow                *UsageStats              `json:"usage_stats_by_flow" validate:"required"`
	UsageStatsByAgentTypeForFlow    []AgentTypeUsageStats    `json:"usage_stats_by_agent_type_for_flow" validate:"omitempty"`
	ToolcallsStatsByFlow            *ToolcallsStats          `json:"toolcalls_stats_by_flow" validate:"required"`
	ToolcallsStatsByFunctionForFlow []FunctionToolcallsStats `json:"toolcalls_stats_by_function_for_flow" validate:"omitempty"`
	FlowStatsByFlow                 *FlowStats               `json:"flow_stats_by_flow" validate:"required"`
}

// Valid is function to control input/output data
func (f FlowUsageResponse) Valid() error {
	if err := validate.Struct(f); err != nil {
		return err
	}
	if f.UsageStatsByFlow != nil {
		if err := f.UsageStatsByFlow.Valid(); err != nil {
			return err
		}
	}
	for i := range f.UsageStatsByAgentTypeForFlow {
		if err := f.UsageStatsByAgentTypeForFlow[i].Valid(); err != nil {
			return err
		}
	}
	if f.ToolcallsStatsByFlow != nil {
		if err := f.ToolcallsStatsByFlow.Valid(); err != nil {
			return err
		}
	}
	for i := range f.ToolcallsStatsByFunctionForFlow {
		if err := f.ToolcallsStatsByFunctionForFlow[i].Valid(); err != nil {
			return err
		}
	}
	if f.FlowStatsByFlow != nil {
		if err := f.FlowStatsByFlow.Valid(); err != nil {
			return err
		}
	}
	return nil
}

// Validate is function to use callback to control input/output data
func (f FlowUsageResponse) Validate(db *gorm.DB) {
	if err := f.Valid(); err != nil {
		db.AddError(err)
	}
}
