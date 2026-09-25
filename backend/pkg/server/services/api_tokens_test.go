package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/server/auth"
	"pentagi/pkg/server/models"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingTokenSubs struct {
	subscriptions.SubscriptionsController
	created []database.ApiToken
	updated []database.ApiToken
	deleted []database.ApiToken
}

type recordingTokenPublisher struct {
	subscriptions.APITokenPublisher
	subs *recordingTokenSubs
}

func (p recordingTokenPublisher) APITokenCreated(_ context.Context, token database.APITokenWithSecret) {
	p.subs.created = append(p.subs.created, token.ApiToken)
}

func (p recordingTokenPublisher) APITokenUpdated(_ context.Context, token database.ApiToken) {
	p.subs.updated = append(p.subs.updated, token)
}

func (p recordingTokenPublisher) APITokenDeleted(_ context.Context, token database.ApiToken) {
	p.subs.deleted = append(p.subs.deleted, token)
}

func (s *recordingTokenSubs) NewAPITokenPublisher(int64) subscriptions.APITokenPublisher {
	return recordingTokenPublisher{subs: s}
}

// apiTokensCall runs one token handler for the given user the way the router hands it a request.
func apiTokensCall(handler gin.HandlerFunc, uid uint64, privs []string, tokenID, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("uid", uid)
	c.Set("rid", uint64(2))
	c.Set("uhash", "testhash")
	c.Set("prm", privs)
	c.Params = gin.Params{{Key: "tokenID", Value: tokenID}}
	c.Request = httptest.NewRequest(http.MethodPost, "/tokens/"+tokenID, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler(c)

	return w
}

func apiTokensPublished(events []database.ApiToken, tokenID string) []database.ApiToken {
	var sent []database.ApiToken
	for _, event := range events {
		if event.TokenID == tokenID {
			sent = append(sent, event)
		}
	}

	return sent
}

// apiTokensStoredRow is every column of the token's row as text, whatever the model maps.
func apiTokensStoredRow(t *testing.T, db *gorm.DB, tokenID string) string {
	t.Helper()

	rows, err := db.Raw("SELECT * FROM api_tokens WHERE token_id = ?", tokenID).Rows()
	require.NoError(t, err)
	defer rows.Close()

	columns, err := rows.Columns()
	require.NoError(t, err)
	require.True(t, rows.Next(), "token %s was not stored", tokenID)

	values := make([]any, len(columns))
	pointers := make([]any, len(columns))
	for i := range values {
		pointers[i] = &values[i]
	}
	require.NoError(t, rows.Scan(pointers...))

	var row strings.Builder
	for i, value := range values {
		fmt.Fprintf(&row, "%s=%s;", columns[i], value)
	}

	return row.String()
}

func apiTokensName(name string) *string {
	return &name
}

func TestAPITokens_CreateToken_IssuesASignedTokenOnlyForAValidRequest(t *testing.T) {
	for _, tc := range []struct {
		name     string
		salt     string
		seed     []models.APIToken
		body     string
		wantCode int
		wantBody string
		wantName *string
	}{
		{
			name:     "a named token",
			salt:     "custom_salt",
			body:     `{"ttl": 3600, "name": "Test Token"}`,
			wantCode: http.StatusCreated,
			wantName: apiTokensName("Test Token"),
		},
		{
			name:     "a token without a name",
			salt:     "custom_salt",
			body:     `{"ttl": 7200}`,
			wantCode: http.StatusCreated,
		},
		{
			name:     "an empty name beside another unnamed token",
			salt:     "custom_salt",
			seed:     []models.APIToken{{TokenID: "unnamed1", UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive}},
			body:     `{"ttl": 3600, "name": ""}`,
			wantCode: http.StatusCreated,
		},
		{
			name: "a name only another user gave a token",
			salt: "custom_salt",
			seed: []models.APIToken{{TokenID: "others1", UserID: 2, RoleID: 2, TTL: 3600,
				Status: models.TokenStatusActive, Name: apiTokensName("Duplicate Name")}},
			body:     `{"ttl": 3600, "name": "Duplicate Name"}`,
			wantCode: http.StatusCreated,
			wantName: apiTokensName("Duplicate Name"),
		},
		{
			name: "a name the user already gave a token",
			salt: "custom_salt",
			seed: []models.APIToken{{TokenID: "mine1", UserID: 1, RoleID: 2, TTL: 3600,
				Status: models.TokenStatusActive, Name: apiTokensName("Duplicate Name")}},
			body:     `{"ttl": 3600, "name": "Duplicate Name"}`,
			wantCode: http.StatusBadRequest,
			wantBody: "already exists",
		},
		{
			name:     "the default salt",
			salt:     "salt",
			body:     `{"ttl": 3600}`,
			wantCode: http.StatusBadRequest,
			wantBody: `"code":"Token.CreationDisabled"`,
		},
		{
			name:     "an empty salt",
			salt:     "",
			body:     `{"ttl": 3600}`,
			wantCode: http.StatusBadRequest,
			wantBody: `"code":"Token.CreationDisabled"`,
		},
		{
			name:     "a ttl under a minute",
			salt:     "custom_salt",
			body:     `{"ttl": 30}`,
			wantCode: http.StatusBadRequest,
			wantBody: `"code":"Token.InvalidRequest"`,
		},
		{
			name:     "a ttl over three years",
			salt:     "custom_salt",
			body:     `{"ttl": 100000000}`,
			wantCode: http.StatusBadRequest,
			wantBody: `"code":"Token.InvalidRequest"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()
			for _, token := range tc.seed {
				require.NoError(t, db.Create(&token).Error)
			}

			subs := &recordingTokenSubs{}
			service := NewTokenService(db, tc.salt, auth.NewTokenCache(db), subs)

			w := apiTokensCall(service.CreateToken, 1, []string{"settings.tokens.create"}, "", tc.body)

			require.Equal(t, tc.wantCode, w.Code, w.Body.String())
			if tc.wantCode != http.StatusCreated {
				assert.Contains(t, w.Body.String(), tc.wantBody)
				assert.Empty(t, subs.created, "a refused token must not reach subscribers")
				return
			}

			var resp struct {
				Status string `json:"status"`
				Data   struct {
					Token   string `json:"token"`
					TokenID string `json:"token_id"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, "success", resp.Status)
			assert.Len(t, resp.Data.TokenID, 10)

			claims, err := auth.ValidateAPIToken(resp.Data.Token, tc.salt)
			require.NoError(t, err, "the token must verify under the instance salt")
			assert.Equal(t, resp.Data.TokenID, claims.TokenID)

			require.Len(t, subs.created, 1, "a created token must reach subscribers")
			assert.Equal(t, resp.Data.TokenID, subs.created[0].TokenID)

			var stored models.APIToken
			require.NoError(t, db.Where("token_id = ?", resp.Data.TokenID).First(&stored).Error)
			assert.Equal(t, tc.wantName, stored.Name, "an empty name must reach the column as no name")
			assert.NotContains(t, apiTokensStoredRow(t, db, resp.Data.TokenID), resp.Data.Token,
				"only the token's metadata is stored, never the secret")
		})
	}
}

func TestAPITokens_ListTokens_ShowsOnlyOwnTokensUnlessAdmin(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	for _, token := range []models.APIToken{
		{TokenID: "token1", UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive},
		{TokenID: "token2", UserID: 1, RoleID: 2, TTL: 7200, Status: models.TokenStatusActive},
		{TokenID: "token3", UserID: 2, RoleID: 2, TTL: 3600, Status: models.TokenStatusActive},
		{TokenID: "token4", UserID: 1, RoleID: 2, TTL: 3600, Status: models.TokenStatusRevoked},
	} {
		require.NoError(t, db.Create(&token).Error)
	}

	service := NewTokenService(db, "custom_salt", auth.NewTokenCache(db), nil)

	for _, tc := range []struct {
		name          string
		uid           uint64
		permissions   []string
		expectedCount int
	}{
		{"a user sees their own tokens, revoked ones included", 1, []string{"settings.tokens.view"}, 3},
		{"an admin sees every token", 1, []string{"settings.tokens.view", "settings.tokens.admin"}, 4},
		{"another user sees only their own token", 2, []string{"settings.tokens.view"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := apiTokensCall(service.ListTokens, tc.uid, tc.permissions, "", "")

			assert.Equal(t, http.StatusOK, w.Code)

			var response struct {
				Status string `json:"status"`
				Data   struct {
					Tokens []models.APIToken `json:"tokens"`
					Total  uint64            `json:"total"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
			assert.Equal(t, "success", response.Status)
			assert.Equal(t, tc.expectedCount, len(response.Data.Tokens))
			assert.Equal(t, uint64(tc.expectedCount), response.Data.Total)
		})
	}
}

func TestAPITokens_GetToken_ServesOnlyTheOwnerOrAnAdmin(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	require.NoError(t, db.Create(&models.APIToken{TokenID: "usertoken1", UserID: 1, RoleID: 2, TTL: 3600,
		Status: models.TokenStatusActive}).Error)
	require.NoError(t, db.Create(&models.APIToken{TokenID: "usertoken2", UserID: 2, RoleID: 2, TTL: 3600,
		Status: models.TokenStatusActive}).Error)

	service := NewTokenService(db, "custom_salt", auth.NewTokenCache(db), nil)

	for _, tc := range []struct {
		name         string
		tokenID      string
		permissions  []string
		expectedCode int
	}{
		{"the owner gets their token", "usertoken1", []string{"settings.tokens.view"}, http.StatusOK},
		{"another user's token is refused", "usertoken2", []string{"settings.tokens.view"}, http.StatusForbidden},
		{"an admin gets another user's token", "usertoken2", []string{"settings.tokens.view", "settings.tokens.admin"}, http.StatusOK},
		{"a token that does not exist", "nonexistent", []string{"settings.tokens.view", "settings.tokens.admin"}, http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := apiTokensCall(service.GetToken, 1, tc.permissions, tc.tokenID, "")

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.expectedCode == http.StatusOK {
				var response struct {
					Status string          `json:"status"`
					Data   models.APIToken `json:"data"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
				assert.Equal(t, "success", response.Status)
				assert.Equal(t, tc.tokenID, response.Data.TokenID)
			}
		})
	}
}

func TestAPITokens_UpdateToken_ChangesNameAndStatusForTheOwnerOrAnAdmin(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	subs := &recordingTokenSubs{}
	service := NewTokenService(db, "custom_salt", auth.NewTokenCache(db), subs)

	require.NoError(t, db.Create(&models.APIToken{TokenID: "updatetest1", UserID: 1, RoleID: 2, TTL: 3600,
		Status: models.TokenStatusActive}).Error)
	require.NoError(t, db.Create(&models.APIToken{TokenID: "updatetest2", UserID: 1, RoleID: 2, TTL: 3600,
		Status: models.TokenStatusActive}).Error)
	expired := models.APIToken{TokenID: "updateexpired", UserID: 1, RoleID: 2, TTL: 1, Status: models.TokenStatusActive}
	require.NoError(t, db.Create(&expired).Error)
	require.NoError(t, db.Model(&expired).Update("created_at", time.Now().Add(-time.Hour)).Error)

	cached, _, err := service.tokenCache.GetStatus("updatetest1")
	require.NoError(t, err)
	require.Equal(t, models.TokenStatusActive, cached, "the cache has to hold the old status for the rows below to mean anything")

	stored := func(t *testing.T, tokenID string) models.APIToken {
		t.Helper()
		var token models.APIToken
		require.NoError(t, db.Where("token_id = ?", tokenID).First(&token).Error)
		return token
	}
	cachedStatus := func(t *testing.T, tokenID string) (models.TokenStatus, []string) {
		t.Helper()
		status, privileges, err := service.tokenCache.GetStatus(tokenID)
		require.NoError(t, err)
		return status, privileges
	}

	for _, tc := range []struct {
		name         string
		tokenID      string
		uid          uint64
		permissions  []string
		requestBody  string
		expectedCode int
		wantBody     string
		checkResult  func(t *testing.T)
	}{
		{
			name:         "renaming a token",
			tokenID:      "updatetest1",
			uid:          1,
			permissions:  []string{"settings.tokens.edit"},
			requestBody:  `{"name": "Updated Name"}`,
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T) {
				assert.Equal(t, apiTokensName("Updated Name"), stored(t, "updatetest1").Name)
			},
		},
		{
			name:         "a name another token of the user already has",
			tokenID:      "updatetest2",
			uid:          1,
			permissions:  []string{"settings.tokens.edit"},
			requestBody:  `{"name": "Updated Name"}`,
			expectedCode: http.StatusBadRequest,
			wantBody:     "already exists",
			checkResult: func(t *testing.T) {
				assert.Nil(t, stored(t, "updatetest2").Name, "a refused rename must not reach the row")
			},
		},
		{
			name:         "echoing back the derived expired status keeps the stored one",
			tokenID:      "updateexpired",
			uid:          1,
			permissions:  []string{"settings.tokens.edit"},
			requestBody:  `{"name": "Renamed While Expired", "status": "expired"}`,
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T) {
				token := stored(t, "updateexpired")
				assert.Equal(t, models.TokenStatusActive, token.Status)
				assert.Equal(t, apiTokensName("Renamed While Expired"), token.Name)
			},
		},
		{
			name:         "revoking a token",
			tokenID:      "updatetest1",
			uid:          1,
			permissions:  []string{"settings.tokens.edit"},
			requestBody:  `{"status": "revoked"}`,
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T) {
				assert.Equal(t, models.TokenStatusRevoked, stored(t, "updatetest1").Status)
				status, privileges := cachedStatus(t, "updatetest1")
				assert.Equal(t, models.TokenStatusRevoked, status, "a revoked token must not stay active in the cache")
				assert.Contains(t, privileges, "pentagi.automation", "the reloaded entry keeps what every token is granted")
			},
		},
		{
			name:         "reactivating a token",
			tokenID:      "updatetest1",
			uid:          1,
			permissions:  []string{"settings.tokens.edit"},
			requestBody:  `{"status": "active"}`,
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T) {
				assert.Equal(t, models.TokenStatusActive, stored(t, "updatetest1").Status)
				status, _ := cachedStatus(t, "updatetest1")
				assert.Equal(t, models.TokenStatusActive, status)
			},
		},
		{
			name:         "another user's token without the admin privilege",
			tokenID:      "updatetest1",
			uid:          2,
			permissions:  []string{"settings.tokens.edit"},
			requestBody:  `{"name": "Hacked"}`,
			expectedCode: http.StatusForbidden,
			checkResult: func(t *testing.T) {
				assert.Equal(t, apiTokensName("Updated Name"), stored(t, "updatetest1").Name)
			},
		},
		{
			name:         "another user's token with the admin privilege",
			tokenID:      "updatetest1",
			uid:          2,
			permissions:  []string{"settings.tokens.edit", "settings.tokens.admin"},
			requestBody:  `{"status": "revoked"}`,
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T) {
				assert.Equal(t, models.TokenStatusRevoked, stored(t, "updatetest1").Status)
			},
		},
		{
			name:         "revoking a token that never got a name",
			tokenID:      "updatetest2",
			uid:          1,
			permissions:  []string{"settings.tokens.edit"},
			requestBody:  `{"status": "revoked"}`,
			expectedCode: http.StatusOK,
			checkResult: func(t *testing.T) {
				token := stored(t, "updatetest2")
				require.Nil(t, token.Name, "the fixture must still be nameless for this case to mean anything")
				assert.Equal(t, models.TokenStatusRevoked, token.Status)

				published := apiTokensPublished(subs.updated, "updatetest2")
				require.Len(t, published, 1, "a nameless token must still reach subscribers")
				assert.False(t, published[0].Name.Valid)
				assert.Equal(t, database.TokenStatusRevoked, published[0].Status)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := apiTokensCall(service.UpdateToken, tc.uid, tc.permissions, tc.tokenID, tc.requestBody)

			assert.Equal(t, tc.expectedCode, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), tc.wantBody)
			tc.checkResult(t)
		})
	}
}

func TestAPITokens_DeleteToken_SoftDeletesForTheOwnerOrAnAdmin(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	subs := &recordingTokenSubs{}
	service := NewTokenService(db, "custom_salt", auth.NewTokenCache(db), subs)

	for _, tc := range []struct {
		name         string
		ownerID      uint64
		tokenID      string
		permissions  []string
		expectedCode int
	}{
		{"the owner deletes their token", 1, "deltest1", []string{"settings.tokens.delete"}, http.StatusOK},
		{"another user's token is refused", 2, "deltest2", []string{"settings.tokens.delete"}, http.StatusForbidden},
		{"an admin deletes another user's token", 2, "deltest3", []string{"settings.tokens.delete", "settings.tokens.admin"}, http.StatusOK},
		{"a token that does not exist", 0, "nonexistent", []string{"settings.tokens.delete", "settings.tokens.admin"}, http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.ownerID != 0 {
				require.NoError(t, db.Create(&models.APIToken{TokenID: tc.tokenID, UserID: tc.ownerID, RoleID: 2,
					TTL: 3600, Status: models.TokenStatusActive}).Error)
				_, _, err := service.tokenCache.GetStatus(tc.tokenID)
				require.NoError(t, err, "the cache has to hold the token for the case to mean anything")
			}

			w := apiTokensCall(service.DeleteToken, 1, tc.permissions, tc.tokenID, "")

			require.Equal(t, tc.expectedCode, w.Code, w.Body.String())
			if tc.ownerID == 0 {
				return
			}

			var token models.APIToken
			require.NoError(t, db.Unscoped().Where("token_id = ?", tc.tokenID).First(&token).Error)
			_, _, cacheErr := service.tokenCache.GetStatus(tc.tokenID)
			if tc.expectedCode != http.StatusOK {
				assert.Nil(t, token.DeletedAt, "a refused delete must leave the token in place")
				assert.NoError(t, cacheErr)
				assert.Empty(t, apiTokensPublished(subs.deleted, tc.tokenID))
				return
			}

			assert.NotNil(t, token.DeletedAt, "the delete is soft")
			assert.ErrorIs(t, cacheErr, gorm.ErrRecordNotFound, "a deleted token must not stay valid in the cache")
			assert.Len(t, apiTokensPublished(subs.deleted, tc.tokenID), 1, "a deleted token must reach subscribers")
		})
	}
}
