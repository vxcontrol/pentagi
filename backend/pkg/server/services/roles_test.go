package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"pentagi/pkg/server/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRoles_ReadsRespectTheCallersRoleScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name     string
		privs    []string
		target   string
		wantCode int
		wantIDs  []uint64
	}{
		{name: "list own role", target: "/roles/?page=1&type=init", wantCode: http.StatusOK, wantIDs: []uint64{2}},
		{name: "read own role", target: "/roles/2", wantCode: http.StatusOK, wantIDs: []uint64{2}},
		{name: "another role is hidden", target: "/roles/1", wantCode: http.StatusNotFound},
		{name: "list all with roles.view", privs: []string{"roles.view"}, target: "/roles/?page=1&type=init", wantCode: http.StatusOK, wantIDs: []uint64{2, 1}},
		{name: "read another with roles.view", privs: []string{"roles.view"}, target: "/roles/1", wantCode: http.StatusOK, wantIDs: []uint64{1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()
			svc := NewRoleService(db)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("rid", uint64(2))
				c.Set("prm", tc.privs)
			})
			router.GET("/roles/", svc.GetRoles)
			router.GET("/roles/:roleID", svc.GetRole)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.target, nil))
			require.Equal(t, tc.wantCode, rec.Code, rec.Body.String())
			if rec.Code != http.StatusOK {
				return
			}
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
			var found []models.RolePrivileges
			if tc.target == "/roles/?page=1&type=init" {
				var list roles
				require.NoError(t, json.Unmarshal(envelope.Data, &list))
				require.Equal(t, uint64(len(tc.wantIDs)), list.Total)
				found = list.Roles
			} else {
				var role models.RolePrivileges
				require.NoError(t, json.Unmarshal(envelope.Data, &role))
				found = append(found, role)
			}
			var ids []uint64
			for _, role := range found {
				ids = append(ids, role.ID)
				require.NotEmpty(t, role.Privileges, "privileges of role "+strconv.FormatUint(role.ID, 10))
				for _, privilege := range role.Privileges {
					require.Equal(t, role.ID, privilege.RoleID)
				}
			}
			require.Equal(t, tc.wantIDs, ids)
		})
	}
}
