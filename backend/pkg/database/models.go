// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package database

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type ContainerStatus string

const (
	ContainerStatusStarting ContainerStatus = "starting"
	ContainerStatusRunning  ContainerStatus = "running"
	ContainerStatusStopped  ContainerStatus = "stopped"
	ContainerStatusDeleted  ContainerStatus = "deleted"
	ContainerStatusFailed   ContainerStatus = "failed"
)

func (e *ContainerStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ContainerStatus(s)
	case string:
		*e = ContainerStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for ContainerStatus: %T", src)
	}
	return nil
}

type NullContainerStatus struct {
	ContainerStatus ContainerStatus `json:"container_status"`
	Valid           bool            `json:"valid"` // Valid is true if ContainerStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullContainerStatus) Scan(value interface{}) error {
	if value == nil {
		ns.ContainerStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ContainerStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullContainerStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ContainerStatus), nil
}

type ContainerType string

const (
	ContainerTypePrimary   ContainerType = "primary"
	ContainerTypeSecondary ContainerType = "secondary"
)

func (e *ContainerType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ContainerType(s)
	case string:
		*e = ContainerType(s)
	default:
		return fmt.Errorf("unsupported scan type for ContainerType: %T", src)
	}
	return nil
}

type NullContainerType struct {
	ContainerType ContainerType `json:"container_type"`
	Valid         bool          `json:"valid"` // Valid is true if ContainerType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullContainerType) Scan(value interface{}) error {
	if value == nil {
		ns.ContainerType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ContainerType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullContainerType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ContainerType), nil
}

type FlowStatus string

const (
	FlowStatusCreated  FlowStatus = "created"
	FlowStatusRunning  FlowStatus = "running"
	FlowStatusWaiting  FlowStatus = "waiting"
	FlowStatusFinished FlowStatus = "finished"
	FlowStatusFailed   FlowStatus = "failed"
)

func (e *FlowStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = FlowStatus(s)
	case string:
		*e = FlowStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for FlowStatus: %T", src)
	}
	return nil
}

type NullFlowStatus struct {
	FlowStatus FlowStatus `json:"flow_status"`
	Valid      bool       `json:"valid"` // Valid is true if FlowStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullFlowStatus) Scan(value interface{}) error {
	if value == nil {
		ns.FlowStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.FlowStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullFlowStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.FlowStatus), nil
}

type MsgchainType string

const (
	MsgchainTypePrimaryAgent  MsgchainType = "primary_agent"
	MsgchainTypeReporter      MsgchainType = "reporter"
	MsgchainTypeGenerator     MsgchainType = "generator"
	MsgchainTypeRefiner       MsgchainType = "refiner"
	MsgchainTypeReflector     MsgchainType = "reflector"
	MsgchainTypeEnricher      MsgchainType = "enricher"
	MsgchainTypeAdviser       MsgchainType = "adviser"
	MsgchainTypeCoder         MsgchainType = "coder"
	MsgchainTypeMemorist      MsgchainType = "memorist"
	MsgchainTypeSearcher      MsgchainType = "searcher"
	MsgchainTypeInstaller     MsgchainType = "installer"
	MsgchainTypePentester     MsgchainType = "pentester"
	MsgchainTypeSummarizer    MsgchainType = "summarizer"
	MsgchainTypeToolCallFixer MsgchainType = "tool_call_fixer"
)

func (e *MsgchainType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = MsgchainType(s)
	case string:
		*e = MsgchainType(s)
	default:
		return fmt.Errorf("unsupported scan type for MsgchainType: %T", src)
	}
	return nil
}

type NullMsgchainType struct {
	MsgchainType MsgchainType `json:"msgchain_type"`
	Valid        bool         `json:"valid"` // Valid is true if MsgchainType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullMsgchainType) Scan(value interface{}) error {
	if value == nil {
		ns.MsgchainType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.MsgchainType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullMsgchainType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.MsgchainType), nil
}

type MsglogResultFormat string

const (
	MsglogResultFormatPlain    MsglogResultFormat = "plain"
	MsglogResultFormatMarkdown MsglogResultFormat = "markdown"
	MsglogResultFormatTerminal MsglogResultFormat = "terminal"
)

func (e *MsglogResultFormat) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = MsglogResultFormat(s)
	case string:
		*e = MsglogResultFormat(s)
	default:
		return fmt.Errorf("unsupported scan type for MsglogResultFormat: %T", src)
	}
	return nil
}

type NullMsglogResultFormat struct {
	MsglogResultFormat MsglogResultFormat `json:"msglog_result_format"`
	Valid              bool               `json:"valid"` // Valid is true if MsglogResultFormat is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullMsglogResultFormat) Scan(value interface{}) error {
	if value == nil {
		ns.MsglogResultFormat, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.MsglogResultFormat.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullMsglogResultFormat) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.MsglogResultFormat), nil
}

type MsglogType string

const (
	MsglogTypeThoughts MsglogType = "thoughts"
	MsglogTypeBrowser  MsglogType = "browser"
	MsglogTypeTerminal MsglogType = "terminal"
	MsglogTypeFile     MsglogType = "file"
	MsglogTypeSearch   MsglogType = "search"
	MsglogTypeAdvice   MsglogType = "advice"
	MsglogTypeAsk      MsglogType = "ask"
	MsglogTypeInput    MsglogType = "input"
	MsglogTypeDone     MsglogType = "done"
)

func (e *MsglogType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = MsglogType(s)
	case string:
		*e = MsglogType(s)
	default:
		return fmt.Errorf("unsupported scan type for MsglogType: %T", src)
	}
	return nil
}

type NullMsglogType struct {
	MsglogType MsglogType `json:"msglog_type"`
	Valid      bool       `json:"valid"` // Valid is true if MsglogType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullMsglogType) Scan(value interface{}) error {
	if value == nil {
		ns.MsglogType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.MsglogType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullMsglogType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.MsglogType), nil
}

type SearchengineType string

const (
	SearchengineTypeGoogle     SearchengineType = "google"
	SearchengineTypeTavily     SearchengineType = "tavily"
	SearchengineTypeTraversaal SearchengineType = "traversaal"
	SearchengineTypeBrowser    SearchengineType = "browser"
)

func (e *SearchengineType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = SearchengineType(s)
	case string:
		*e = SearchengineType(s)
	default:
		return fmt.Errorf("unsupported scan type for SearchengineType: %T", src)
	}
	return nil
}

type NullSearchengineType struct {
	SearchengineType SearchengineType `json:"searchengine_type"`
	Valid            bool             `json:"valid"` // Valid is true if SearchengineType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullSearchengineType) Scan(value interface{}) error {
	if value == nil {
		ns.SearchengineType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.SearchengineType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullSearchengineType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.SearchengineType), nil
}

type SubtaskStatus string

const (
	SubtaskStatusCreated  SubtaskStatus = "created"
	SubtaskStatusRunning  SubtaskStatus = "running"
	SubtaskStatusWaiting  SubtaskStatus = "waiting"
	SubtaskStatusFinished SubtaskStatus = "finished"
	SubtaskStatusFailed   SubtaskStatus = "failed"
)

func (e *SubtaskStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = SubtaskStatus(s)
	case string:
		*e = SubtaskStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for SubtaskStatus: %T", src)
	}
	return nil
}

type NullSubtaskStatus struct {
	SubtaskStatus SubtaskStatus `json:"subtask_status"`
	Valid         bool          `json:"valid"` // Valid is true if SubtaskStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullSubtaskStatus) Scan(value interface{}) error {
	if value == nil {
		ns.SubtaskStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.SubtaskStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullSubtaskStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.SubtaskStatus), nil
}

type TaskStatus string

const (
	TaskStatusCreated  TaskStatus = "created"
	TaskStatusRunning  TaskStatus = "running"
	TaskStatusWaiting  TaskStatus = "waiting"
	TaskStatusFinished TaskStatus = "finished"
	TaskStatusFailed   TaskStatus = "failed"
)

func (e *TaskStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = TaskStatus(s)
	case string:
		*e = TaskStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for TaskStatus: %T", src)
	}
	return nil
}

type NullTaskStatus struct {
	TaskStatus TaskStatus `json:"task_status"`
	Valid      bool       `json:"valid"` // Valid is true if TaskStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullTaskStatus) Scan(value interface{}) error {
	if value == nil {
		ns.TaskStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.TaskStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullTaskStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.TaskStatus), nil
}

type TermlogType string

const (
	TermlogTypeStdin  TermlogType = "stdin"
	TermlogTypeStdout TermlogType = "stdout"
	TermlogTypeStderr TermlogType = "stderr"
)

func (e *TermlogType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = TermlogType(s)
	case string:
		*e = TermlogType(s)
	default:
		return fmt.Errorf("unsupported scan type for TermlogType: %T", src)
	}
	return nil
}

type NullTermlogType struct {
	TermlogType TermlogType `json:"termlog_type"`
	Valid       bool        `json:"valid"` // Valid is true if TermlogType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullTermlogType) Scan(value interface{}) error {
	if value == nil {
		ns.TermlogType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.TermlogType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullTermlogType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.TermlogType), nil
}

type ToolcallStatus string

const (
	ToolcallStatusReceived ToolcallStatus = "received"
	ToolcallStatusRunning  ToolcallStatus = "running"
	ToolcallStatusFinished ToolcallStatus = "finished"
	ToolcallStatusFailed   ToolcallStatus = "failed"
)

func (e *ToolcallStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ToolcallStatus(s)
	case string:
		*e = ToolcallStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for ToolcallStatus: %T", src)
	}
	return nil
}

type NullToolcallStatus struct {
	ToolcallStatus ToolcallStatus `json:"toolcall_status"`
	Valid          bool           `json:"valid"` // Valid is true if ToolcallStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullToolcallStatus) Scan(value interface{}) error {
	if value == nil {
		ns.ToolcallStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ToolcallStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullToolcallStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ToolcallStatus), nil
}

type UserStatus string

const (
	UserStatusCreated UserStatus = "created"
	UserStatusActive  UserStatus = "active"
	UserStatusBlocked UserStatus = "blocked"
)

func (e *UserStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = UserStatus(s)
	case string:
		*e = UserStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for UserStatus: %T", src)
	}
	return nil
}

type NullUserStatus struct {
	UserStatus UserStatus `json:"user_status"`
	Valid      bool       `json:"valid"` // Valid is true if UserStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullUserStatus) Scan(value interface{}) error {
	if value == nil {
		ns.UserStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.UserStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullUserStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.UserStatus), nil
}

type UserType string

const (
	UserTypeLocal UserType = "local"
	UserTypeOauth UserType = "oauth"
)

func (e *UserType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = UserType(s)
	case string:
		*e = UserType(s)
	default:
		return fmt.Errorf("unsupported scan type for UserType: %T", src)
	}
	return nil
}

type NullUserType struct {
	UserType UserType `json:"user_type"`
	Valid    bool     `json:"valid"` // Valid is true if UserType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullUserType) Scan(value interface{}) error {
	if value == nil {
		ns.UserType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.UserType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullUserType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.UserType), nil
}

type VecstoreActionType string

const (
	VecstoreActionTypeRetrieve VecstoreActionType = "retrieve"
	VecstoreActionTypeStore    VecstoreActionType = "store"
)

func (e *VecstoreActionType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = VecstoreActionType(s)
	case string:
		*e = VecstoreActionType(s)
	default:
		return fmt.Errorf("unsupported scan type for VecstoreActionType: %T", src)
	}
	return nil
}

type NullVecstoreActionType struct {
	VecstoreActionType VecstoreActionType `json:"vecstore_action_type"`
	Valid              bool               `json:"valid"` // Valid is true if VecstoreActionType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullVecstoreActionType) Scan(value interface{}) error {
	if value == nil {
		ns.VecstoreActionType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.VecstoreActionType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullVecstoreActionType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.VecstoreActionType), nil
}

type Agentlog struct {
	ID        int64         `json:"id"`
	Initiator MsgchainType  `json:"initiator"`
	Executor  MsgchainType  `json:"executor"`
	Task      string        `json:"task"`
	Result    string        `json:"result"`
	FlowID    int64         `json:"flow_id"`
	TaskID    sql.NullInt64 `json:"task_id"`
	SubtaskID sql.NullInt64 `json:"subtask_id"`
	CreatedAt sql.NullTime  `json:"created_at"`
}

type Container struct {
	ID        int64           `json:"id"`
	Type      ContainerType   `json:"type"`
	Name      string          `json:"name"`
	Image     string          `json:"image"`
	Status    ContainerStatus `json:"status"`
	LocalID   sql.NullString  `json:"local_id"`
	LocalDir  sql.NullString  `json:"local_dir"`
	FlowID    int64           `json:"flow_id"`
	CreatedAt sql.NullTime    `json:"created_at"`
	UpdatedAt sql.NullTime    `json:"updated_at"`
}

type Flow struct {
	ID            int64           `json:"id"`
	Status        FlowStatus      `json:"status"`
	Title         string          `json:"title"`
	Model         string          `json:"model"`
	ModelProvider string          `json:"model_provider"`
	Language      string          `json:"language"`
	Functions     json.RawMessage `json:"functions"`
	Prompts       json.RawMessage `json:"prompts"`
	UserID        int64           `json:"user_id"`
	CreatedAt     sql.NullTime    `json:"created_at"`
	UpdatedAt     sql.NullTime    `json:"updated_at"`
	DeletedAt     sql.NullTime    `json:"deleted_at"`
	TraceID       sql.NullString  `json:"trace_id"`
}

type Msgchain struct {
	ID            int64           `json:"id"`
	Type          MsgchainType    `json:"type"`
	Model         string          `json:"model"`
	ModelProvider string          `json:"model_provider"`
	UsageIn       int64           `json:"usage_in"`
	UsageOut      int64           `json:"usage_out"`
	Chain         json.RawMessage `json:"chain"`
	FlowID        int64           `json:"flow_id"`
	TaskID        sql.NullInt64   `json:"task_id"`
	SubtaskID     sql.NullInt64   `json:"subtask_id"`
	CreatedAt     sql.NullTime    `json:"created_at"`
	UpdatedAt     sql.NullTime    `json:"updated_at"`
}

type Msglog struct {
	ID           int64              `json:"id"`
	Type         MsglogType         `json:"type"`
	Message      string             `json:"message"`
	Result       string             `json:"result"`
	FlowID       int64              `json:"flow_id"`
	TaskID       sql.NullInt64      `json:"task_id"`
	SubtaskID    sql.NullInt64      `json:"subtask_id"`
	CreatedAt    sql.NullTime       `json:"created_at"`
	ResultFormat MsglogResultFormat `json:"result_format"`
}

type Privilege struct {
	ID     int64  `json:"id"`
	RoleID int64  `json:"role_id"`
	Name   string `json:"name"`
}

type Prompt struct {
	ID     int64  `json:"id"`
	Type   string `json:"type"`
	UserID int64  `json:"user_id"`
	Prompt string `json:"prompt"`
}

type Role struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Screenshot struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	Url       string       `json:"url"`
	FlowID    int64        `json:"flow_id"`
	CreatedAt sql.NullTime `json:"created_at"`
}

type Searchlog struct {
	ID        int64            `json:"id"`
	Initiator MsgchainType     `json:"initiator"`
	Executor  MsgchainType     `json:"executor"`
	Engine    SearchengineType `json:"engine"`
	Query     string           `json:"query"`
	Result    string           `json:"result"`
	FlowID    int64            `json:"flow_id"`
	TaskID    sql.NullInt64    `json:"task_id"`
	SubtaskID sql.NullInt64    `json:"subtask_id"`
	CreatedAt sql.NullTime     `json:"created_at"`
}

type Subtask struct {
	ID          int64         `json:"id"`
	Status      SubtaskStatus `json:"status"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Result      string        `json:"result"`
	TaskID      int64         `json:"task_id"`
	CreatedAt   sql.NullTime  `json:"created_at"`
	UpdatedAt   sql.NullTime  `json:"updated_at"`
}

type Task struct {
	ID        int64        `json:"id"`
	Status    TaskStatus   `json:"status"`
	Title     string       `json:"title"`
	Input     string       `json:"input"`
	Result    string       `json:"result"`
	FlowID    int64        `json:"flow_id"`
	CreatedAt sql.NullTime `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
}

type Termlog struct {
	ID          int64        `json:"id"`
	Type        TermlogType  `json:"type"`
	Text        string       `json:"text"`
	ContainerID int64        `json:"container_id"`
	CreatedAt   sql.NullTime `json:"created_at"`
}

type Toolcall struct {
	ID        int64           `json:"id"`
	CallID    string          `json:"call_id"`
	Status    ToolcallStatus  `json:"status"`
	Name      string          `json:"name"`
	Args      json.RawMessage `json:"args"`
	Result    string          `json:"result"`
	FlowID    int64           `json:"flow_id"`
	TaskID    sql.NullInt64   `json:"task_id"`
	SubtaskID sql.NullInt64   `json:"subtask_id"`
	CreatedAt sql.NullTime    `json:"created_at"`
	UpdatedAt sql.NullTime    `json:"updated_at"`
}

type User struct {
	ID                     int64          `json:"id"`
	Hash                   string         `json:"hash"`
	Type                   UserType       `json:"type"`
	Mail                   string         `json:"mail"`
	Name                   string         `json:"name"`
	Password               sql.NullString `json:"password"`
	Status                 UserStatus     `json:"status"`
	RoleID                 int64          `json:"role_id"`
	PasswordChangeRequired bool           `json:"password_change_required"`
	Provider               sql.NullString `json:"provider"`
	CreatedAt              sql.NullTime   `json:"created_at"`
}

type Vecstorelog struct {
	ID        int64              `json:"id"`
	Initiator MsgchainType       `json:"initiator"`
	Executor  MsgchainType       `json:"executor"`
	Filter    json.RawMessage    `json:"filter"`
	Query     string             `json:"query"`
	Action    VecstoreActionType `json:"action"`
	Result    string             `json:"result"`
	FlowID    int64              `json:"flow_id"`
	TaskID    sql.NullInt64      `json:"task_id"`
	SubtaskID sql.NullInt64      `json:"subtask_id"`
	CreatedAt sql.NullTime       `json:"created_at"`
}