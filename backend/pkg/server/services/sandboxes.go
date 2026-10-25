package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/controller"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/response"
	"pentagi/pkg/tools"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// SandboxService provisions headless Docker sandboxes: a flow row and a
// container with no agent loop and no LLM provider behind it, driven entirely
// by whoever holds the caller's API token through exec/file endpoints. It
// never registers with the flow controller, so flows.* endpoints that depend
// on an in-memory FlowWorker (stop/finish/rename/report) correctly see it as
// not found; this service owns its full lifecycle instead.
type SandboxService struct {
	orm    *gorm.DB
	db     database.Querier
	docker docker.DockerClient
	cfg    *config.Config
	subs   subscriptions.SubscriptionsController
}

func NewSandboxService(
	ormDB *gorm.DB,
	db database.Querier,
	dockerClient docker.DockerClient,
	cfg *config.Config,
	subs subscriptions.SubscriptionsController,
) *SandboxService {
	return &SandboxService{
		orm:    ormDB,
		db:     db,
		docker: dockerClient,
		cfg:    cfg,
		subs:   subs,
	}
}

// CreateSandbox is a function to provision a headless sandbox
// @Summary Create a headless Docker sandbox with no agent loop and no LLM provider
// @Tags Sandboxes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param json body models.CreateSandbox true "sandbox model to create"
// @Success 201 {object} response.successResp{data=models.Flow} "sandbox created successful"
// @Failure 400 {object} response.errorResp "invalid sandbox request data"
// @Failure 403 {object} response.errorResp "creating sandbox not permitted"
// @Failure 500 {object} response.errorResp "internal error on creating sandbox"
// @Router /sandboxes/ [post]
func (s *SandboxService) CreateSandbox(c *gin.Context) {
	var req models.CreateSandbox
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrFlowsInvalidRequest, err)
		return
	}

	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "flows.create") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	if err := req.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating sandbox data")
		response.Error(c, response.ErrFlowsInvalidData, err)
		return
	}

	title := req.Title
	if title == "" {
		title = "headless sandbox"
	}
	image := req.Image
	if image == "" {
		image = s.cfg.DockerDefaultImageForPentest
	}

	uid := int64(c.GetUint64("uid"))

	dbFlow, err := s.db.CreateHeadlessFlow(c, database.CreateHeadlessFlowParams{
		Title:  title,
		Status: database.FlowStatusCreated,
		UserID: uid,
	})
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error creating headless flow")
		response.Error(c, response.ErrInternal, err)
		return
	}

	workerConfig, hostConfig := docker.WorkerSpec(s.cfg, image)
	containerName := tools.PrimaryTerminalName(s.cfg.TenantPrefix(), dbFlow.ID)

	if _, err := s.docker.RunContainer(
		c, containerName, database.ContainerTypePrimary, dbFlow.ID, workerConfig, hostConfig,
	); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error launching sandbox container")
		// The flow row stays: it lets the operator see and retry/delete a
		// sandbox whose container failed to come up, instead of losing the id.
		response.Error(c, response.ErrInternal, err)
		return
	}

	if _, err := s.db.UpdateFlowStatus(c, database.UpdateFlowStatusParams{
		Status: database.FlowStatusRunning,
		ID:     dbFlow.ID,
	}); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error marking sandbox running")
		response.Error(c, response.ErrInternal, err)
		return
	}

	var flow models.Flow
	if err := s.orm.Model(&flow).Where("id = ?", dbFlow.ID).Take(&flow).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting sandbox by id")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusCreated, flow)
}

// DeleteSandbox is a function to stop and remove a headless sandbox's container
// @Summary Stop a headless sandbox and remove its container
// @Tags Sandboxes
// @Produce json
// @Security BearerAuth
// @Param sandboxID path int true "sandbox id" minimum(0)
// @Success 200 {object} response.successResp{data=models.Flow} "sandbox stopped successful"
// @Failure 403 {object} response.errorResp "stopping sandbox not permitted"
// @Failure 404 {object} response.errorResp "sandbox not found"
// @Failure 500 {object} response.errorResp "internal error on stopping sandbox"
// @Router /sandboxes/{sandboxID} [delete]
func (s *SandboxService) DeleteSandbox(c *gin.Context) {
	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "flows.delete") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	dbFlow, cnt, ok := s.loadOwnedSandbox(c)
	if !ok {
		return
	}

	if cnt.LocalID.Valid && cnt.LocalID.String != "" {
		if err := s.docker.RemoveContainer(c, cnt.LocalID.String, cnt.ID); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error removing sandbox container")
			response.Error(c, response.ErrInternal, err)
			return
		}
	}

	if _, err := s.db.UpdateFlowStatus(c, database.UpdateFlowStatusParams{
		Status: database.FlowStatusFinished,
		ID:     dbFlow.ID,
	}); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error marking sandbox finished")
		response.Error(c, response.ErrInternal, err)
		return
	}

	var flow models.Flow
	if err := s.orm.Model(&flow).Where("id = ?", dbFlow.ID).Take(&flow).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting sandbox by id")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, flow)
}

// ExecSandbox is a function to run a command inside a sandbox's container
// @Summary Run a command inside a headless sandbox
// @Tags Sandboxes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sandboxID path int true "sandbox id" minimum(0)
// @Param json body models.SandboxExec true "command to run"
// @Success 200 {object} response.successResp{data=string} "command output"
// @Failure 400 {object} response.errorResp "invalid exec request data"
// @Failure 403 {object} response.errorResp "executing in sandbox not permitted"
// @Failure 404 {object} response.errorResp "sandbox not found"
// @Failure 500 {object} response.errorResp "internal error on executing in sandbox"
// @Router /sandboxes/{sandboxID}/exec [post]
func (s *SandboxService) ExecSandbox(c *gin.Context) {
	var req models.SandboxExec
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrFlowsInvalidRequest, err)
		return
	}

	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "flows.edit") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	if err := req.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating exec data")
		response.Error(c, response.ErrFlowsInvalidData, err)
		return
	}

	dbFlow, cnt, ok := s.loadOwnedSandbox(c)
	if !ok {
		return
	}

	term := s.terminalTool(c, dbFlow.ID, cnt)

	action := tools.TerminalAction{
		Input:   req.Input,
		Cwd:     req.Cwd,
		Detach:  tools.Bool(req.Detach),
		Timeout: tools.Int64(req.Timeout),
		Message: "headless sandbox command",
	}
	argsJSON, err := json.Marshal(action)
	if err != nil {
		response.Error(c, response.ErrInternal, err)
		return
	}

	result, err := term.Handle(c, tools.TerminalToolName, argsJSON)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error executing sandbox command")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, result)
}

// GetSandboxFile is a function to read a file from a sandbox's container
// @Summary Read a file from a headless sandbox
// @Tags Sandboxes
// @Produce json
// @Security BearerAuth
// @Param sandboxID path int true "sandbox id" minimum(0)
// @Param path query string true "absolute path inside the container"
// @Success 200 {object} response.successResp{data=string} "file content"
// @Failure 400 {object} response.errorResp "invalid path"
// @Failure 403 {object} response.errorResp "reading sandbox file not permitted"
// @Failure 404 {object} response.errorResp "sandbox not found"
// @Failure 500 {object} response.errorResp "internal error on reading sandbox file"
// @Router /sandboxes/{sandboxID}/file [get]
func (s *SandboxService) GetSandboxFile(c *gin.Context) {
	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "flows.view") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	path := c.Query("path")
	if path == "" {
		response.Error(c, response.ErrFlowsInvalidRequest, nil)
		return
	}

	dbFlow, cnt, ok := s.loadOwnedSandbox(c)
	if !ok {
		return
	}

	term := s.terminalTool(c, dbFlow.ID, cnt)

	action := tools.FileAction{
		Action:  tools.ReadFile,
		Path:    tools.String(path),
		Message: "headless sandbox file read",
	}
	argsJSON, err := json.Marshal(action)
	if err != nil {
		response.Error(c, response.ErrInternal, err)
		return
	}

	result, err := term.Handle(c, tools.FileToolName, argsJSON)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error reading sandbox file")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, result)
}

// PutSandboxFile is a function to write a file into a sandbox's container
// @Summary Write a file into a headless sandbox
// @Tags Sandboxes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sandboxID path int true "sandbox id" minimum(0)
// @Param json body models.SandboxFileWrite true "file to write"
// @Success 200 {object} response.successResp{data=string} "write result"
// @Failure 400 {object} response.errorResp "invalid write request data"
// @Failure 403 {object} response.errorResp "writing sandbox file not permitted"
// @Failure 404 {object} response.errorResp "sandbox not found"
// @Failure 500 {object} response.errorResp "internal error on writing sandbox file"
// @Router /sandboxes/{sandboxID}/file [put]
func (s *SandboxService) PutSandboxFile(c *gin.Context) {
	var req models.SandboxFileWrite
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrFlowsInvalidRequest, err)
		return
	}

	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "flows.edit") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	if err := req.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating write data")
		response.Error(c, response.ErrFlowsInvalidData, err)
		return
	}

	dbFlow, cnt, ok := s.loadOwnedSandbox(c)
	if !ok {
		return
	}

	term := s.terminalTool(c, dbFlow.ID, cnt)

	action := tools.FileAction{
		Action:  tools.WriteFile,
		Path:    tools.String(req.Path),
		Content: req.Content,
		Message: "headless sandbox file write",
	}
	argsJSON, err := json.Marshal(action)
	if err != nil {
		response.Error(c, response.ErrInternal, err)
		return
	}

	result, err := term.Handle(c, tools.FileToolName, argsJSON)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error writing sandbox file")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, result)
}

// loadOwnedSandbox resolves :sandboxID to a headless flow owned by the
// caller and its primary container, writing the HTTP response itself on any
// failure so handlers can just check ok.
func (s *SandboxService) loadOwnedSandbox(c *gin.Context) (database.Flow, database.Container, bool) {
	sandboxID, err := strconv.ParseInt(c.Param("sandboxID"), 10, 64)
	if err != nil {
		response.Error(c, response.ErrFlowsInvalidRequest, err)
		return database.Flow{}, database.Container{}, false
	}

	uid := int64(c.GetUint64("uid"))

	dbFlow, err := s.db.GetUserHeadlessFlow(c, database.GetUserHeadlessFlowParams{
		ID:     sandboxID,
		UserID: uid,
	})
	if err != nil {
		response.ErrorWithLevel(c, response.ErrFlowsNotFound, err, logrus.WarnLevel)
		return database.Flow{}, database.Container{}, false
	}

	// DeleteSandbox marks the flow finished but keeps the row (and the container
	// row, now 'deleted', with a local_id nothing backs any more); reject those
	// explicitly instead of letting exec/file calls fall through to a generic
	// "container runtime is not operational" result.
	if dbFlow.Status == database.FlowStatusFinished || dbFlow.Status == database.FlowStatusFailed {
		response.ErrorWithLevel(c, response.ErrFlowsNotFound, fmt.Errorf("sandbox %d is %s", dbFlow.ID, dbFlow.Status), logrus.WarnLevel)
		return database.Flow{}, database.Container{}, false
	}

	cnt, err := s.db.GetFlowPrimaryContainer(c, dbFlow.ID)
	if err != nil {
		response.Error(c, response.ErrFlowsNotFound, err)
		return database.Flow{}, database.Container{}, false
	}

	return dbFlow, cnt, true
}

// terminalTool reuses the same terminal.Handle exec/file plumbing agents use
// (timeout bounds, output truncation, path validation), publishing through
// the same termlog pipeline so a sandbox's commands show up in its termlogs
// exactly like an agent's would.
func (s *SandboxService) terminalTool(c *gin.Context, flowID int64, cnt database.Container) tools.Tool {
	uid := int64(c.GetUint64("uid"))
	pub := s.subs.NewFlowPublisher(uid, flowID)
	termLog := controller.NewFlowTermLogWorker(s.db, flowID, pub)

	return tools.NewTerminalTool(
		flowID, nil, nil,
		cnt.ID, cnt.LocalID.String,
		s.cfg.TenantPrefix(),
		s.docker,
		termLog,
		time.Duration(s.cfg.TerminalToolTimeout)*time.Second,
	)
}
