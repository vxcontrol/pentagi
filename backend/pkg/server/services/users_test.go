package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pentagi/pkg/server/auth"
	"pentagi/pkg/server/models"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser_CreatesUserPreferences(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userCache := auth.NewUserCache(db)
	service := NewUserService(db, userCache)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set up context with admin permissions
	c.Set("uid", uint64(1))
	c.Set("rid", uint64(1))
	c.Set("uhash", "testhash1")
	c.Set("prm", []string{"users.create"})

	// Create request body
	userRequest := models.UserPassword{
		User: models.User{
			Mail:   "newuser@test.com",
			Name:   "New User",
			RoleID: 2,
			Status: models.UserStatusActive,
			Type:   models.UserTypeLocal,
		},
		Password: "SecurePass123!",
	}

	body, err := json.Marshal(userRequest)
	require.NoError(t, err)

	c.Request, _ = http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call the handler
	service.CreateUser(c)

	// Check response status
	assert.Equal(t, http.StatusCreated, w.Code, "Expected HTTP 201 Created")

	// Verify user was created
	var createdUser models.User
	err = db.Where("mail = ?", "newuser@test.com").First(&createdUser).Error
	require.NoError(t, err, "User should be created in database")
	assert.Equal(t, "New User", createdUser.Name)
	assert.Equal(t, uint64(2), createdUser.RoleID)

	// Verify user_preferences was created
	var userPrefs models.UserPreferences
	err = db.Where("user_id = ?", createdUser.ID).First(&userPrefs).Error
	require.NoError(t, err, "User preferences should be created in database")
	assert.Equal(t, createdUser.ID, userPrefs.UserID)
	assert.NotNil(t, userPrefs.Preferences.FavoriteFlows)
	assert.Equal(t, 0, len(userPrefs.Preferences.FavoriteFlows), "FavoriteFlows should be empty array")
}

func TestCreateUser_RollbackOnPreferencesError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Drop user_preferences table to simulate error
	db.Exec("DROP TABLE user_preferences")

	userCache := auth.NewUserCache(db)
	service := NewUserService(db, userCache)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("uid", uint64(1))
	c.Set("rid", uint64(1))
	c.Set("uhash", "testhash1")
	c.Set("prm", []string{"users.create"})

	userRequest := models.UserPassword{
		User: models.User{
			Mail:   "failuser@test.com",
			Name:   "Fail User",
			RoleID: 2,
			Status: models.UserStatusActive,
			Type:   models.UserTypeLocal,
		},
		Password: "SecurePass123!",
	}

	body, err := json.Marshal(userRequest)
	require.NoError(t, err)

	c.Request, _ = http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	service.CreateUser(c)

	// Should return error
	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected HTTP 500 on preferences creation error")

	// Verify user was NOT created (transaction rolled back)
	var user models.User
	err = db.Where("mail = ?", "failuser@test.com").First(&user).Error
	assert.Error(t, err, "User should not exist due to transaction rollback")
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestCreateUser_InvalidPermissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userCache := auth.NewUserCache(db)
	service := NewUserService(db, userCache)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set up context WITHOUT users.create permission
	c.Set("uid", uint64(2))
	c.Set("rid", uint64(2))
	c.Set("uhash", "testhash2")
	c.Set("prm", []string{"flows.view"})

	userRequest := models.UserPassword{
		User: models.User{
			Mail:   "unauthorized@test.com",
			Name:   "Unauthorized User",
			RoleID: 2,
			Status: models.UserStatusActive,
			Type:   models.UserTypeLocal,
		},
		Password: "SecurePass123!",
	}

	body, err := json.Marshal(userRequest)
	require.NoError(t, err)

	c.Request, _ = http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	service.CreateUser(c)

	// Should return forbidden
	assert.Equal(t, http.StatusForbidden, w.Code, "Expected HTTP 403 Forbidden")

	// Verify user was NOT created
	var user models.User
	err = db.Where("mail = ?", "unauthorized@test.com").First(&user).Error
	assert.Error(t, err, "User should not be created")
}

func TestCreateUser_MultipleUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userCache := auth.NewUserCache(db)
	service := NewUserService(db, userCache)

	testCases := []struct {
		name     string
		mail     string
		username string
		roleID   uint64
	}{
		{
			name:     "create first user",
			mail:     "newuser1@test.com",
			username: "User One",
			roleID:   2,
		},
		{
			name:     "create second user",
			mail:     "newuser2@test.com",
			username: "User Two",
			roleID:   2,
		},
		{
			name:     "create third user",
			mail:     "newuser3@test.com",
			username: "User Three",
			roleID:   2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Set("uid", uint64(1))
			c.Set("rid", uint64(1))
			c.Set("uhash", "testhash1")
			c.Set("prm", []string{"users.create"})

			userRequest := models.UserPassword{
				User: models.User{
					Mail:   tc.mail,
					Name:   tc.username,
					RoleID: tc.roleID,
					Status: models.UserStatusActive,
					Type:   models.UserTypeLocal,
				},
				Password: "SecurePass123!",
			}

			body, err := json.Marshal(userRequest)
			require.NoError(t, err)

			c.Request, _ = http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			service.CreateUser(c)

			assert.Equal(t, http.StatusCreated, w.Code, "Expected HTTP 201 Created")

			// Verify both user and preferences were created
			var user models.User
			err = db.Where("mail = ?", tc.mail).First(&user).Error
			require.NoError(t, err)

			var prefs models.UserPreferences
			err = db.Where("user_id = ?", user.ID).First(&prefs).Error
			require.NoError(t, err)
			assert.Equal(t, user.ID, prefs.UserID)
		})
	}

	// Verify all users and preferences exist
	var userCount int
	db.Model(&models.User{}).Where("mail LIKE ?", "newuser%@test.com").Count(&userCount)
	assert.Equal(t, 3, userCount, "Should have 3 newly created users")

	var prefsCount int
	db.Model(&models.UserPreferences{}).Count(&prefsCount)
	assert.Equal(t, 5, prefsCount, "Should have 5 user preferences total (2 initial + 3 created)")
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userCache := auth.NewUserCache(db)
	service := NewUserService(db, userCache)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("uid", uint64(1))
	c.Set("rid", uint64(1))
	c.Set("uhash", "testhash1")
	c.Set("prm", []string{"users.create"})

	// Invalid JSON
	c.Request, _ = http.NewRequest("POST", "/users/", bytes.NewBufferString("{invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	service.CreateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected HTTP 400 Bad Request")
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userCache := auth.NewUserCache(db)
	service := NewUserService(db, userCache)

	// Create first user
	gin.SetMode(gin.TestMode)
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)

	c1.Set("uid", uint64(1))
	c1.Set("rid", uint64(1))
	c1.Set("uhash", "testhash1")
	c1.Set("prm", []string{"users.create"})

	userRequest := models.UserPassword{
		User: models.User{
			Mail:   "duplicate@test.com",
			Name:   "First User",
			RoleID: 2,
			Status: models.UserStatusActive,
			Type:   models.UserTypeLocal,
		},
		Password: "SecurePass123!",
	}

	body, err := json.Marshal(userRequest)
	require.NoError(t, err)

	c1.Request, _ = http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
	c1.Request.Header.Set("Content-Type", "application/json")

	service.CreateUser(c1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Try to create second user with same email
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)

	c2.Set("uid", uint64(1))
	c2.Set("rid", uint64(1))
	c2.Set("uhash", "testhash1")
	c2.Set("prm", []string{"users.create"})

	userRequest2 := models.UserPassword{
		User: models.User{
			Mail:   "duplicate@test.com", // Same email
			Name:   "Second User",
			RoleID: 2,
			Status: models.UserStatusActive,
			Type:   models.UserTypeLocal,
		},
		Password: "AnotherPass456!",
	}

	body2, err := json.Marshal(userRequest2)
	require.NoError(t, err)

	c2.Request, _ = http.NewRequest("POST", "/users/", bytes.NewBuffer(body2))
	c2.Request.Header.Set("Content-Type", "application/json")

	service.CreateUser(c2)

	// Should fail due to unique constraint
	assert.Equal(t, http.StatusInternalServerError, w2.Code, "Expected error on duplicate email")

	// Verify only one user exists
	var count int
	db.Model(&models.User{}).Where("mail = ?", "duplicate@test.com").Count(&count)
	assert.Equal(t, 1, count, "Should have only one user with this email")
}
