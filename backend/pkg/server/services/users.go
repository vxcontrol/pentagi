package services

import (
	"errors"
	"net/http"
	"slices"

	"pentagi/pkg/server/auth"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/rdb"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type users struct {
	Users []models.UserRole `json:"users"`
	Total uint64            `json:"total"`
}

type usersGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var usersSQLMappers = map[string]any{
	"id":         "{{table}}.id",
	"hash":       "{{table}}.hash",
	"type":       "{{table}}.type",
	"mail":       "{{table}}.mail",
	"name":       "{{table}}.name",
	"role_id":    "{{table}}.role_id",
	"status":     "{{table}}.status",
	"created_at": "{{table}}.created_at",
	"data":       "({{table}}.hash || ' ' || {{table}}.mail || ' ' || {{table}}.name || ' ' || {{table}}.status)",
}

type UserService struct {
	db        *gorm.DB
	userCache *auth.UserCache
}

func NewUserService(db *gorm.DB, userCache *auth.UserCache) *UserService {
	return &UserService{
		db:        db,
		userCache: userCache,
	}
}

// GetCurrentUser is a function to return account information
// @Summary Retrieve current user information
// @Tags Users
// @Produce json
// @Success 200 {object} response.successResp{data=models.UserRolePrivileges} "user info received successful"
// @Failure 403 {object} response.errorResp "getting current user not permitted"
// @Failure 404 {object} response.errorResp "current user not found"
// @Failure 500 {object} response.errorResp "internal error on getting current user"
// @Router /user/ [get]
func (s *UserService) GetCurrentUser(c *gin.Context) {
	var (
		err  error
		resp models.UserRolePrivileges
	)

	uid := c.GetUint64("uid")

	if err = s.db.Take(&resp.User, "id = ?", uid).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding current user")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrUsersNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	if err = s.db.Take(&resp.Role, "id = ?", resp.User.RoleID).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding role by role id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrGetUserModelsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	if err = s.db.Model(&resp.Role).Association("privileges").Find(&resp.Role.Privileges).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding privileges by role id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrGetUserModelsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	if err = resp.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user data '%s'", resp.Hash)
		response.Error(c, response.ErrUsersInvalidData, err)
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// ChangePasswordCurrentUser is a function to update account password
// @Summary Update password for current user (account)
// @Tags Users
// @Accept json
// @Produce json
// @Param json body models.Password true "container to validate and update account password"
// @Success 200 {object} response.successResp "account password updated successful"
// @Failure 400 {object} response.errorResp "invalid account password form data"
// @Failure 403 {object} response.errorResp "updating account password not permitted"
// @Failure 404 {object} response.errorResp "current user not found"
// @Failure 500 {object} response.errorResp "internal error on updating account password"
// @Router /user/password [put]
func (s *UserService) ChangePasswordCurrentUser(c *gin.Context) {
	var (
		encPass []byte
		err     error
		form    models.Password
		user    models.UserPassword
	)

	if err = c.ShouldBindJSON(&form); err != nil || form.Valid() != nil {
		if err == nil {
			err = form.Valid()
		}
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrChangePasswordCurrentUserInvalidPassword, err)
		return
	}

	uid := c.GetUint64("uid")
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", uid)
	}

	if err = s.db.Scopes(scope).Take(&user).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding current user")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrUsersNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	} else if err = user.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user data '%s'", user.Hash)
		response.Error(c, response.ErrUsersInvalidData, err)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.CurrentPassword)); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error checking password for current user")
		response.Error(c, response.ErrChangePasswordCurrentUserInvalidCurrentPassword, err)
		return
	}

	if encPass, err = rdb.EncryptPassword(form.Password); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error making new password for current user")
		response.Error(c, response.ErrChangePasswordCurrentUserInvalidNewPassword, err)
		return
	}

	// Use map to update fields to avoid GORM ignoring zero values (false for bool)
	updates := map[string]any{
		"password":                 string(encPass),
		"password_change_required": false,
	}

	if err = s.db.Model(&user).Scopes(scope).Updates(updates).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error updating password for current user")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, struct{}{})
}

// GetUsers returns users list
// @Summary Retrieve users list by filters
// @Tags Users
// @Produce json
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=users} "users list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting users not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting users"
// @Router /users/ [get]
func (s *UserService) GetUsers(c *gin.Context) {
	var (
		err   error
		query rdb.TableQuery
		resp  users
		rids  []uint64
		roles []models.Role
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrUsersInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	scope := func(db *gorm.DB) *gorm.DB {
		if !slices.Contains(privs, "users.view") {
			return db.Where("id = ?", uid)
		}
		return db
	}

	query.Init("users", usersSQLMappers)

	if query.Group != "" {
		if _, ok := usersSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding users grouped: group field not found")
			response.Error(c, response.ErrUsersInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped usersGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding users grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.Users, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding users")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for _, user := range resp.Users {
		rids = append(rids, user.RoleID)
	}

	if err = s.db.Find(&roles, "id IN (?)", rids).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding linked roles")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := range resp.Users {
		roleID := resp.Users[i].RoleID
		for _, role := range roles {
			if roleID == role.ID {
				resp.Users[i].Role = role
				break
			}
		}
	}

	for i := 0; i < len(resp.Users); i++ {
		if err = resp.Users[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating user data '%s'", resp.Users[i].Hash)
			response.Error(c, response.ErrUsersInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetUser is a function to return user by hash
// @Summary Retrieve user by hash
// @Tags Users
// @Produce json
// @Param hash path string true "hash in hex format (md5)" minlength(32) maxlength(32)
// @Success 200 {object} response.successResp{data=models.UserRolePrivileges} "user received successful"
// @Failure 403 {object} response.errorResp "getting user not permitted"
// @Failure 404 {object} response.errorResp "user not found"
// @Failure 500 {object} response.errorResp "internal error on getting user"
// @Router /users/{hash} [get]
func (s *UserService) GetUser(c *gin.Context) {
	var (
		err  error
		hash string = c.Param("hash")
		resp models.UserRolePrivileges
	)

	uhash := c.GetString("uhash")
	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "users.view") && uhash != hash {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	if err = s.db.Take(&resp.User, "hash = ?", hash).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding user by hash")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrUsersNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	if err = s.db.Take(&resp.Role, "id = ?", resp.User.RoleID).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding role by role id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrGetUserModelsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	if err = s.db.Model(&resp.Role).Association("privileges").Find(&resp.Role.Privileges).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding privileges by role id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrGetUserModelsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}
	if err = resp.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user data '%s'", resp.Hash)
		response.Error(c, response.ErrUsersInvalidData, err)
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// CreateUser is a function to create new user
// @Summary Create new user
// @Tags Users
// @Accept json
// @Produce json
// @Param json body models.UserPassword true "user model to create from"
// @Success 201 {object} response.successResp{data=models.UserRole} "user created successful"
// @Failure 400 {object} response.errorResp "invalid user request data"
// @Failure 403 {object} response.errorResp "creating user not permitted"
// @Failure 500 {object} response.errorResp "internal error on creating user"
// @Router /users/ [post]
func (s *UserService) CreateUser(c *gin.Context) {
	var (
		encPassword []byte
		err         error
		resp        models.UserRole
		user        models.UserPassword
	)

	if err = c.ShouldBindJSON(&user); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrUsersInvalidRequest, err)
		return
	}

	rid := c.GetUint64("rid")
	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "users.create") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	privsCurrentUser, err := s.GetUserPrivileges(c, rid)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting current user privileges")
		response.Error(c, response.ErrInternal, err)
		return
	}

	privsNewUser, err := s.GetUserPrivileges(c, user.RoleID)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting new user privileges")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if !s.CheckPrivilege(c, privsCurrentUser, privsNewUser) {
		logger.FromContext(c).Errorf("error checking new user privileges")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	user.ID = 0
	user.Hash = rdb.MakeUserHash(user.Name)
	if err = user.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user")
		response.Error(c, response.ErrCreateUserInvalidUser, err)
		return
	}

	if encPassword, err = rdb.EncryptPassword(user.Password); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error encoding password")
		response.Error(c, response.ErrInternal, err)
		return
	} else {
		user.Password = string(encPassword)
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		logger.FromContext(c).WithError(tx.Error).Errorf("error starting transaction")
		response.Error(c, response.ErrInternal, tx.Error)
		return
	}

	if err = tx.Create(&user).Error; err != nil {
		tx.Rollback()
		logger.FromContext(c).WithError(err).Errorf("error creating user")
		response.Error(c, response.ErrInternal, err)
		return
	}

	preferences := models.NewUserPreferences(user.ID)
	if err = tx.Create(preferences).Error; err != nil {
		tx.Rollback()
		logger.FromContext(c).WithError(err).Errorf("error creating user preferences")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if err = tx.Commit().Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error committing transaction")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if err = s.db.Take(&resp.User, "hash = ?", user.Hash).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding user by hash")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if err = s.db.Take(&resp.Role, "id = ?", resp.User.RoleID).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding role by role id")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if err = resp.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user data '%s'", resp.Hash)
		response.Error(c, response.ErrUsersInvalidData, err)
		return
	}

	s.userCache.Invalidate(resp.User.ID)

	response.Success(c, http.StatusCreated, resp)
}

// PatchUser is a function to update user by hash
// @Summary Update user
// @Tags Users
// @Accept json
// @Produce json
// @Param hash path string true "user hash in hex format (md5)" minlength(32) maxlength(32)
// @Param json body models.UserPassword true "user model to update"
// @Success 200 {object} response.successResp{data=models.UserRole} "user updated successful"
// @Failure 400 {object} response.errorResp "invalid user request data"
// @Failure 403 {object} response.errorResp "updating user not permitted"
// @Failure 404 {object} response.errorResp "user not found"
// @Failure 500 {object} response.errorResp "internal error on updating user"
// @Router /users/{hash} [put]
func (s *UserService) PatchUser(c *gin.Context) {
	var (
		err  error
		hash = c.Param("hash")
		resp models.UserRole
		user models.UserPassword
	)

	if err = c.ShouldBindJSON(&user); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrUsersInvalidRequest, err)
		return
	} else if hash != user.Hash {
		logger.FromContext(c).Errorf("mismatch user hash to requested one")
		response.Error(c, response.ErrUsersInvalidRequest, nil)
		return
	} else if err = user.User.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user JSON")
		response.Error(c, response.ErrUsersInvalidRequest, err)
		return
	} else if err = user.Valid(); user.Password != "" && err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user password")
		response.Error(c, response.ErrUsersInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	uhash := c.GetString("uhash")
	privs := c.GetStringSlice("prm")
	scope := func(db *gorm.DB) *gorm.DB {
		if slices.Contains(privs, "users.edit") {
			return db.Where("hash = ?", hash)
		} else {
			return db.Where("hash = ? AND id = ?", hash, uid)
		}
	}
	if !slices.Contains(privs, "users.edit") && uhash != hash {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	// Check if user exists before updating
	var existingUser models.User
	if err = s.db.Scopes(scope).Take(&existingUser).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding user by hash")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrUsersNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	if user.Password != "" {
		var encPassword []byte
		encPassword, err = rdb.EncryptPassword(user.Password)
		if err != nil {
			logger.FromContext(c).WithError(err).Errorf("error encoding password")
			response.Error(c, response.ErrInternal, err)
			return
		}
		// Use map to update fields to avoid GORM ignoring zero values (false for bool)
		updates := map[string]any{
			"name":                     user.Name,
			"status":                   user.Status,
			"password":                 string(encPassword),
			"password_change_required": false,
		}
		err = s.db.Model(&existingUser).Updates(updates).Error
	} else {
		updates := map[string]any{
			"name":   user.Name,
			"status": user.Status,
		}
		err = s.db.Model(&existingUser).Updates(updates).Error
	}

	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error updating user by hash '%s'", hash)
		response.Error(c, response.ErrInternal, err)
		return
	}

	if err = s.db.Scopes(scope).Take(&resp.User).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding user by hash")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrUsersNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	if err = s.db.Take(&resp.Role, "id = ?", resp.User.RoleID).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding role by role id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrPatchUserModelsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}
	if err = resp.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user data '%s'", resp.Hash)
		response.Error(c, response.ErrInternal, err)
		return
	}

	s.userCache.Invalidate(resp.User.ID)

	response.Success(c, http.StatusOK, resp)
}

// DeleteUser is a function to delete user by hash
// @Summary Delete user by hash
// @Tags Users
// @Produce json
// @Param hash path string true "hash in hex format (md5)" minlength(32) maxlength(32)
// @Success 200 {object} response.successResp "user deleted successful"
// @Failure 403 {object} response.errorResp "deleting user not permitted"
// @Failure 404 {object} response.errorResp "user not found"
// @Failure 500 {object} response.errorResp "internal error on deleting user"
// @Router /users/{hash} [delete]
func (s *UserService) DeleteUser(c *gin.Context) {
	var (
		err  error
		hash string = c.Param("hash")
		user models.UserRole
	)

	uid := c.GetUint64("uid")
	uhash := c.GetString("uhash")
	privs := c.GetStringSlice("prm")
	scope := func(db *gorm.DB) *gorm.DB {
		if slices.Contains(privs, "users.delete") {
			return db.Where("hash = ?", hash)
		} else {
			return db.Where("hash = ? AND id = ?", hash, uid)
		}
	}
	if !slices.Contains(privs, "users.delete") && uhash != hash {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	if err = s.db.Scopes(scope).Take(&user.User).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding user by hash")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrUsersNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	if err = s.db.Take(&user.Role, "id = ?", user.User.RoleID).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding role by role id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrDeleteUserModelsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}
	if err = user.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user data '%s'", user.Hash)
		response.Error(c, response.ErrUsersInvalidData, err)
		return
	}

	if err = s.db.Delete(&user.User).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error deleting user by hash '%s'", hash)
		response.Error(c, response.ErrInternal, err)
		return
	}

	s.userCache.Invalidate(user.ID)

	response.Success(c, http.StatusOK, struct{}{})
}

// GetUserPrivileges is a function to return user privileges
func (s *UserService) GetUserPrivileges(c *gin.Context, rid uint64) ([]string, error) {
	var (
		err   error
		privs []string
		resp  []models.Privilege
	)

	if err = s.db.Model(&models.Privilege{}).Where("role_id = ?", rid).Find(&resp).Error; err != nil {
		return nil, err
	}
	for _, p := range resp {
		if err = p.Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating privilege data '%s'", p.Name)
			return nil, err
		}
		privs = append(privs, p.Name)
	}

	return privs, nil
}

// CheckPrivilege is a function to check if user has privilege
func (s *UserService) CheckPrivilege(c *gin.Context, privsCurrentUser, privsNewUser []string) bool {
	for _, priv := range privsNewUser {
		if !slices.Contains(privsCurrentUser, priv) {
			return false
		}
	}
	return true
}
