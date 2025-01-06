package services

import (
	"errors"
	"net/http"
	"slices"
	"strconv"

	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/rdb"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type roles struct {
	Roles []models.RolePrivileges `json:"roles"`
	Total uint64                  `json:"total"`
}

var rolesSQLMappers = map[string]interface{}{
	"id":   "{{table}}.id",
	"name": "{{table}}.name",
	"data": "{{table}}.name",
}

type RoleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{
		db: db,
	}
}

// GetRoles is a function to return roles list
// @Summary Retrieve roles list
// @Tags Roles
// @Produce json
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=roles} "roles list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting roles not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting roles"
// @Router /roles/ [get]
func (s *RoleService) GetRoles(c *gin.Context) {
	var (
		err   error
		query rdb.TableQuery
		resp  roles
		rids  []uint64
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrRolesInvalidRequest, err)
		return
	}

	rid := c.GetUint64("rid")
	privs := c.GetStringSlice("prm")
	scope := func(db *gorm.DB) *gorm.DB {
		if !slices.Contains(privs, "roles.view'") {
			return db.Where("role_id = ?", rid)
		}
		return db
	}

	query.Init("roles", rolesSQLMappers)

	if resp.Total, err = query.Query(s.db, &resp.Roles, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding roles")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for _, role := range resp.Roles {
		rids = append(rids, role.ID)
	}

	var privsObjs []models.Privilege
	if err = s.db.Find(&privsObjs, "role_id IN (?)", rids).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding linked roles")
		response.Error(c, response.ErrInternal, err)
		return
	}

	privsRoles := make(map[uint64][]models.Privilege)
	for i := range privsObjs {
		privsRoles[privsObjs[i].RoleID] = append(privsRoles[privsObjs[i].RoleID], privsObjs[i])
	}

	for i := range resp.Roles {
		resp.Roles[i].Privileges = privsRoles[resp.Roles[i].ID]
	}

	for i := 0; i < len(resp.Roles); i++ {
		if err = resp.Roles[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating role data '%d'", resp.Roles[i].ID)
			response.Error(c, response.ErrRolesInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetRole is a function to return role by id
// @Summary Retrieve role by id
// @Tags Roles
// @Produce json
// @Param id path uint64 true "role id"
// @Success 200 {object} response.successResp{data=models.RolePrivileges} "role received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting role not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting role"
// @Router /roles/{roleID} [get]
func (s *RoleService) GetRole(c *gin.Context) {
	var (
		err    error
		resp   models.RolePrivileges
		roleID uint64
	)

	rid := c.GetUint64("rid")
	privs := c.GetStringSlice("prm")
	scope := func(db *gorm.DB) *gorm.DB {
		if !slices.Contains(privs, "roles.view'") {
			return db.Where("role_id = ?", rid)
		}
		return db
	}

	if roleID, err = strconv.ParseUint(c.Param("roleID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing role id")
		response.Error(c, response.ErrRolesInvalidRequest, err)
		return
	}

	if err := s.db.Scopes(scope).Take(&resp, "id = ?", roleID).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding role by id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrRolesNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}
	if err := s.db.Model(&resp).Association("privileges").Find(&resp.Privileges).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding role privileges by role model")
		response.Error(c, response.ErrInternal, err)
		return
	}
	if err := resp.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating role data '%d'", resp.ID)
		response.Error(c, response.ErrRolesInvalidData, err)
		return
	}

	response.Success(c, http.StatusOK, resp)
}
