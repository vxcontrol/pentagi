package services

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	graphmodel "pentagi/pkg/graph/model"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/resources"
	"pentagi/pkg/server/models"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupResourceServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	db.LogMode(false)

	require.NoError(t, db.Exec(`
		CREATE TABLE user_resources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			hash TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			size INTEGER NOT NULL DEFAULT 0,
			is_dir BOOLEAN NOT NULL DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, path)
		)
	`).Error)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

func seedResource(t *testing.T, db *gorm.DB, rec models.UserResource) models.UserResource {
	t.Helper()

	require.NoError(t, db.Create(&rec).Error)
	return rec
}

func resourcePaths(entries []models.ResourceEntry) []string {
	paths := make([]string, len(entries))
	for i, entry := range entries {
		paths[i] = entry.Path
	}
	return paths
}

func allResourcePaths(t *testing.T, db *gorm.DB) []string {
	t.Helper()

	var recs []models.UserResource
	require.NoError(t, db.Order("path ASC").Find(&recs).Error)
	paths := make([]string, len(recs))
	for i, rec := range recs {
		paths[i] = rec.Path
	}
	return paths
}

func newResourceTestContext(method, target string, body *bytes.Buffer, privs []string) (*gin.Context, *httptest.ResponseRecorder) {
	return newResourceTestContextWithUID(method, target, body, privs, 1)
}

func newResourceTestContextWithUID(
	method, target string,
	body *bytes.Buffer,
	privs []string,
	uid uint64,
) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("uid", uid)
	c.Set("prm", privs)
	if body == nil {
		body = bytes.NewBuffer(nil)
	}
	c.Request = httptest.NewRequest(method, target, body)
	return c, w
}

func decodeResourceListResponse(t *testing.T, w *httptest.ResponseRecorder) models.ResourceList {
	t.Helper()

	var resp struct {
		Status string              `json:"status"`
		Data   models.ResourceList `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "success", resp.Status)
	return resp.Data
}

type resourceEvent struct {
	action string
	path   string
}

type uploadTestFile struct {
	name    string
	content string
}

func multipartUploadBody(t *testing.T, files []uploadTestFile) (*bytes.Buffer, string) {
	return multipartUploadBodyWithField(t, files, "files")
}

func multipartUploadBodyWithField(
	t *testing.T,
	files []uploadTestFile,
	fieldName string,
) (*bytes.Buffer, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, file := range files {
		part, err := writer.CreateFormFile(fieldName, file.name)
		require.NoError(t, err)
		_, err = part.Write([]byte(file.content))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	return &body, writer.FormDataContentType()
}

// countResourceBlobs returns the number of *.blob files in the resources dir.
func countResourceBlobs(t *testing.T, dataDir string) int {
	t.Helper()

	dir := resources.ResourcesDir(dataDir)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return 0
	}
	require.NoError(t, err)
	count := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".blob" {
			count++
		}
	}
	return count
}

type captureSubscriptions struct {
	events []resourceEvent
}

func (s *captureSubscriptions) NewFlowSubscriber(int64, int64) subscriptions.FlowSubscriber {
	return nil
}
func (s *captureSubscriptions) NewFlowPublisher(int64, int64) subscriptions.FlowPublisher { return nil }
func (s *captureSubscriptions) NewResourceSubscriber(int64) subscriptions.ResourceSubscriber {
	return nil
}
func (s *captureSubscriptions) NewProviderSubscriber(int64) subscriptions.ProviderSubscriber {
	return nil
}
func (s *captureSubscriptions) NewProviderPublisher(int64) subscriptions.ProviderPublisher {
	return nil
}
func (s *captureSubscriptions) NewAPITokenSubscriber(int64) subscriptions.APITokenSubscriber {
	return nil
}
func (s *captureSubscriptions) NewAPITokenPublisher(int64) subscriptions.APITokenPublisher {
	return nil
}
func (s *captureSubscriptions) NewSettingsSubscriber(int64) subscriptions.SettingsSubscriber {
	return nil
}
func (s *captureSubscriptions) NewSettingsPublisher(int64) subscriptions.SettingsPublisher {
	return nil
}
func (s *captureSubscriptions) NewFlowTemplateSubscriber(int64) subscriptions.FlowTemplateSubscriber {
	return nil
}
func (s *captureSubscriptions) NewFlowTemplatePublisher(int64) subscriptions.FlowTemplatePublisher {
	return nil
}
func (s *captureSubscriptions) NewResourcePublisher(int64) subscriptions.ResourcePublisher {
	return &captureResourcePublisher{events: &s.events}
}

type captureResourcePublisher struct {
	userID int64
	events *[]resourceEvent
}

func (p *captureResourcePublisher) GetUserID() int64 { return p.userID }
func (p *captureResourcePublisher) SetUserID(userID int64) {
	p.userID = userID
}
func (p *captureResourcePublisher) ResourceAdded(_ context.Context, resource *graphmodel.UserResource) {
	*p.events = append(*p.events, resourceEvent{action: "added", path: resource.Path})
}
func (p *captureResourcePublisher) ResourceUpdated(_ context.Context, resource *graphmodel.UserResource) {
	*p.events = append(*p.events, resourceEvent{action: "updated", path: resource.Path})
}
func (p *captureResourcePublisher) ResourceDeleted(_ context.Context, resource *graphmodel.UserResource) {
	*p.events = append(*p.events, resourceEvent{action: "deleted", path: resource.Path})
}

func TestResourceService_ListResourcesScenarios(t *testing.T) {
	type seed struct {
		userID  uint64
		path    string
		isDir   bool
		content string
	}

	tests := []struct {
		name              string
		seeds             []seed
		path              string
		recursive         bool
		privs             []string
		uid               uint64
		wantStatus        int
		wantResponsePaths []string
	}{
		{
			name: "list root non-recursive returns top-level entries only",
			seeds: []seed{
				{path: "root.txt", content: "r"},
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
			},
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"root.txt", "docs"},
		},
		{
			name: "list root recursive returns full tree",
			seeds: []seed{
				{path: "root.txt", content: "r"},
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/b.txt", content: "b"},
			},
			recursive:         true,
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"root.txt", "docs", "docs/a.txt", "docs/sub", "docs/sub/b.txt"},
		},
		{
			name: "list directory non-recursive returns dir and direct children only",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/b.txt", content: "b"},
			},
			path:              "docs",
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"docs", "docs/a.txt", "docs/sub"},
		},
		{
			name: "list directory recursive returns whole subtree",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/b.txt", content: "b"},
			},
			path:              "docs",
			recursive:         true,
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"docs", "docs/a.txt", "docs/sub", "docs/sub/b.txt"},
		},
		{
			name: "non-admin does not see other user's resources at root",
			seeds: []seed{
				{userID: 1, path: "own.txt", content: "own"},
				{userID: 2, path: "alien.txt", content: "alien"},
			},
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"own.txt"},
		},
		{
			name: "admin sees other user's resources at root",
			seeds: []seed{
				{userID: 1, path: "own.txt", content: "own"},
				{userID: 2, path: "alien.txt", content: "alien"},
			},
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"own.txt", "alien.txt"},
		},
		{
			name:              "missing privilege returns forbidden",
			seeds:             []seed{{path: "a.txt", content: "a"}},
			privs:             []string{"resources.upload"},
			wantStatus:        http.StatusForbidden,
			wantResponsePaths: nil,
		},
		{
			name:              "invalid path returns bad request",
			path:              "../escape",
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusBadRequest,
			wantResponsePaths: nil,
		},
		{
			name:              "absolute path returns bad request",
			path:              "/abs/path",
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusBadRequest,
			wantResponsePaths: nil,
		},
		{
			name:              "list non-existent directory returns empty list",
			path:              "missing",
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{},
		},
		{
			name: "view privilege does not bypass uid filter when querying subtree",
			seeds: []seed{
				{userID: 2, path: "docs", isDir: true},
				{userID: 2, path: "docs/a.txt", content: "a"},
			},
			path:              "docs",
			recursive:         true,
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupResourceServiceTestDB(t)
			svc := NewResourceService(db, t.TempDir(), nil)
			for _, rec := range tt.seeds {
				userID := rec.userID
				if userID == 0 {
					userID = 1
				}
				seeded := models.UserResource{
					UserID: userID,
					Name:   filepath.Base(rec.path),
					Path:   rec.path,
					IsDir:  rec.isDir,
				}
				if !rec.isDir {
					seeded.Hash = md5HexForService(rec.content)
					seeded.Size = int64(len(rec.content))
				}
				seedResource(t, db, seeded)
			}

			target := "/resources/"
			query := []string{}
			if tt.path != "" {
				query = append(query, "path="+tt.path)
			}
			if tt.recursive {
				query = append(query, "recursive=true")
			}
			if len(query) > 0 {
				target += "?" + strings.Join(query, "&")
			}

			uid := tt.uid
			if uid == 0 {
				uid = 1
			}
			c, w := newResourceTestContextWithUID(http.MethodGet, target, nil, tt.privs, uid)

			svc.ListResources(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				list := decodeResourceListResponse(t, w)
				assert.ElementsMatch(t, tt.wantResponsePaths, resourcePaths(list.Items))
				assert.Equal(t, uint64(len(tt.wantResponsePaths)), list.Total)
			}
		})
	}
}

func TestResourceService_MkdirResourceScenarios(t *testing.T) {
	type seed struct {
		path    string
		isDir   bool
		content string
	}

	tests := []struct {
		name             string
		seeds            []seed
		path             string
		rawBody          string
		privs            []string
		wantStatus       int
		wantPaths        []string
		wantResponsePath string
		wantResponseDir  bool
		wantEvents       []resourceEvent
	}{
		{
			name:             "create new directory at root",
			path:             "docs",
			privs:            []string{"resources.edit"},
			wantStatus:       http.StatusOK,
			wantPaths:        []string{"docs"},
			wantResponsePath: "docs",
			wantResponseDir:  true,
			wantEvents:       []resourceEvent{{action: "added", path: "docs"}},
		},
		{
			name:             "admin can create directory",
			path:             "secret",
			privs:            []string{"resources.admin"},
			wantStatus:       http.StatusOK,
			wantPaths:        []string{"secret"},
			wantResponsePath: "secret",
			wantResponseDir:  true,
			wantEvents:       []resourceEvent{{action: "added", path: "secret"}},
		},
		{
			name:             "deeply nested mkdir creates only the leaf entry",
			path:             "a/b/c",
			privs:            []string{"resources.edit"},
			wantStatus:       http.StatusOK,
			wantPaths:        []string{"a/b/c"},
			wantResponsePath: "a/b/c",
			wantResponseDir:  true,
			wantEvents:       []resourceEvent{{action: "added", path: "a/b/c"}},
		},
		{
			name:             "idempotent existing directory returns existing record without event",
			seeds:            []seed{{path: "docs", isDir: true}},
			path:             "docs",
			privs:            []string{"resources.edit"},
			wantStatus:       http.StatusOK,
			wantPaths:        []string{"docs"},
			wantResponsePath: "docs",
			wantResponseDir:  true,
			wantEvents:       nil,
		},
		{
			name:       "conflict when path occupied by file",
			seeds:      []seed{{path: "docs", content: "data"}},
			path:       "docs",
			privs:      []string{"resources.edit"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"docs"},
			wantEvents: nil,
		},
		{
			name:       "missing privilege returns forbidden",
			path:       "docs",
			privs:      []string{"resources.view"},
			wantStatus: http.StatusForbidden,
			wantPaths:  []string{},
			wantEvents: nil,
		},
		{
			name:       "invalid path returns bad request",
			path:       "../escape",
			privs:      []string{"resources.edit"},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{},
			wantEvents: nil,
		},
		{
			name:       "absolute path returns bad request",
			path:       "/abs",
			privs:      []string{"resources.edit"},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{},
			wantEvents: nil,
		},
		{
			name:       "missing path field returns bad request",
			rawBody:    `{}`,
			privs:      []string{"resources.edit"},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{},
			wantEvents: nil,
		},
		{
			name:       "malformed JSON returns bad request",
			rawBody:    `{not json}`,
			privs:      []string{"resources.edit"},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{},
			wantEvents: nil,
		},
		{
			name:       "empty body returns bad request",
			rawBody:    "",
			privs:      []string{"resources.edit"},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{},
			wantEvents: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupResourceServiceTestDB(t)
			ss := &captureSubscriptions{}
			svc := NewResourceService(db, t.TempDir(), ss)
			for _, rec := range tt.seeds {
				seeded := models.UserResource{
					UserID: 1,
					Name:   filepath.Base(rec.path),
					Path:   rec.path,
					IsDir:  rec.isDir,
				}
				if !rec.isDir {
					seeded.Hash = md5HexForService(rec.content)
					seeded.Size = int64(len(rec.content))
				}
				seedResource(t, db, seeded)
			}

			var body *bytes.Buffer
			if tt.rawBody != "" {
				body = bytes.NewBufferString(tt.rawBody)
			} else if tt.path != "" {
				payload, err := json.Marshal(map[string]string{"path": tt.path})
				require.NoError(t, err)
				body = bytes.NewBuffer(payload)
			} else {
				body = bytes.NewBuffer(nil)
			}

			c, w := newResourceTestContext(http.MethodPost, "/resources/mkdir", body, tt.privs)
			c.Request.Header.Set("Content-Type", "application/json")

			svc.MkdirResource(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantPaths != nil {
				assert.ElementsMatch(t, tt.wantPaths, allResourcePaths(t, db))
			}
			assert.Equal(t, tt.wantEvents, ss.events)
			if tt.wantStatus == http.StatusOK {
				var resp struct {
					Status string                `json:"status"`
					Data   models.ResourceEntry `json:"data"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				require.Equal(t, "success", resp.Status)
				assert.Equal(t, tt.wantResponsePath, resp.Data.Path)
				assert.Equal(t, tt.wantResponseDir, resp.Data.IsDir)
			}
		})
	}
}

func TestResourceService_UploadResourcesScenarios(t *testing.T) {
	type seed struct {
		path    string
		isDir   bool
		content string
	}

	tests := []struct {
		name              string
		dir               string
		files             []uploadTestFile
		fieldName         string
		privs             []string
		seeds             []seed
		wantStatus        int
		wantPaths         []string
		wantResponsePaths []string
		wantEvents        []resourceEvent
		wantMissingBlobs  []string
		wantPresentBlobs  []string
	}{
		{
			name:              "upload without dir creates file in root",
			files:             []uploadTestFile{{name: "report.txt", content: "payload"}},
			privs:             []string{"resources.upload"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"report.txt"},
			wantResponsePaths: []string{"report.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "report.txt"}},
			wantPresentBlobs:  []string{"payload"},
		},
		{
			name:              "admin can upload",
			files:             []uploadTestFile{{name: "report.txt", content: "payload"}},
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"report.txt"},
			wantResponsePaths: []string{"report.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "report.txt"}},
			wantPresentBlobs:  []string{"payload"},
		},
		{
			name:              "single 'file' field accepted as fallback",
			files:             []uploadTestFile{{name: "report.txt", content: "payload"}},
			fieldName:         "file",
			privs:             []string{"resources.upload"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"report.txt"},
			wantResponsePaths: []string{"report.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "report.txt"}},
			wantPresentBlobs:  []string{"payload"},
		},
		{
			name:       "empty multipart form returns bad request",
			files:      []uploadTestFile{},
			privs:      []string{"resources.upload"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:              "upload into existing dir creates file only",
			dir:               "docs",
			files:             []uploadTestFile{{name: "report.txt", content: "payload"}},
			privs:             []string{"resources.upload"},
			seeds:             []seed{{path: "docs", isDir: true}},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"docs", "docs/report.txt"},
			wantResponsePaths: []string{"docs/report.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "docs/report.txt"}},
			wantPresentBlobs:  []string{"payload"},
		},
		{
			name:              "upload into missing dir creates dir and file",
			dir:               "docs",
			files:             []uploadTestFile{{name: "report.txt", content: "payload"}},
			privs:             []string{"resources.upload"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"docs", "docs/report.txt"},
			wantResponsePaths: []string{"docs/report.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "docs"}, {action: "added", path: "docs/report.txt"}},
			wantPresentBlobs:  []string{"payload"},
		},
		{
			name:              "upload into missing three-level nested dir creates parents and files",
			dir:               "docs/sub/deep",
			files:             []uploadTestFile{{name: "a.txt", content: "a"}, {name: "b.txt", content: "b"}},
			privs:             []string{"resources.upload"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"docs", "docs/sub", "docs/sub/deep", "docs/sub/deep/a.txt", "docs/sub/deep/b.txt"},
			wantResponsePaths: []string{"docs/sub/deep/a.txt", "docs/sub/deep/b.txt"},
			wantEvents: []resourceEvent{
				{action: "added", path: "docs"},
				{action: "added", path: "docs/sub"},
				{action: "added", path: "docs/sub/deep"},
				{action: "added", path: "docs/sub/deep/a.txt"},
				{action: "added", path: "docs/sub/deep/b.txt"},
			},
			wantPresentBlobs: []string{"a", "b"},
		},
		{
			name:             "upload target dir path occupied by file conflicts",
			dir:              "docs",
			files:            []uploadTestFile{{name: "report.txt", content: "payload"}},
			privs:            []string{"resources.upload"},
			seeds:            []seed{{path: "docs", content: "file"}},
			wantStatus:       http.StatusConflict,
			wantPaths:        []string{"docs"},
			wantMissingBlobs: []string{"payload"},
			wantPresentBlobs: []string{"file"},
		},
		{
			name:             "upload existing file path conflicts before writing",
			files:            []uploadTestFile{{name: "report.txt", content: "new"}},
			privs:            []string{"resources.upload"},
			seeds:            []seed{{path: "report.txt", content: "old"}},
			wantStatus:       http.StatusConflict,
			wantPaths:        []string{"report.txt"},
			wantMissingBlobs: []string{"new"},
			wantPresentBlobs: []string{"old"},
		},
		{
			name:             "upload duplicate filenames in same request conflicts",
			files:            []uploadTestFile{{name: "report.txt", content: "first"}, {name: "./report.txt", content: "second"}},
			privs:            []string{"resources.upload"},
			wantStatus:       http.StatusConflict,
			wantMissingBlobs: []string{"first", "second"},
		},
		{
			name:             "upload invalid dir rejected",
			dir:              "../docs",
			files:            []uploadTestFile{{name: "report.txt", content: "payload"}},
			privs:            []string{"resources.upload"},
			wantStatus:       http.StatusBadRequest,
			wantMissingBlobs: []string{"payload"},
		},
		{
			name:             "upload invalid filename rejected",
			files:            []uploadTestFile{{name: "bad\nname.txt", content: "payload"}},
			privs:            []string{"resources.upload"},
			wantStatus:       http.StatusBadRequest,
			wantMissingBlobs: []string{"payload"},
		},
		{
			name:             "upload without privilege forbidden",
			files:            []uploadTestFile{{name: "report.txt", content: "payload"}},
			privs:            []string{"resources.view"},
			wantStatus:       http.StatusForbidden,
			wantMissingBlobs: []string{"payload"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupResourceServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &captureSubscriptions{}
			svc := NewResourceService(db, dataDir, ss)
			hashes := map[string]string{}
			for _, rec := range tt.seeds {
				seeded := models.UserResource{
					UserID: 1,
					Name:   filepath.Base(rec.path),
					Path:   rec.path,
					IsDir:  rec.isDir,
				}
				if !rec.isDir {
					seeded.Hash = md5HexForService(rec.content)
					seeded.Size = int64(len(rec.content))
					writeResourceBlob(t, dataDir, seeded.Hash, rec.content)
					hashes[rec.content] = seeded.Hash
				}
				seedResource(t, db, seeded)
			}
			for _, file := range tt.files {
				hashes[file.content] = md5HexForService(file.content)
			}

			fieldName := tt.fieldName
			if fieldName == "" {
				fieldName = "files"
			}
			body, contentType := multipartUploadBodyWithField(t, tt.files, fieldName)
			target := "/resources/"
			if tt.dir != "" {
				target += "?dir=" + tt.dir
			}
			c, w := newResourceTestContext(http.MethodPost, target, body, tt.privs)
			c.Request.Header.Set("Content-Type", contentType)

			svc.UploadResources(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantPaths != nil {
				assert.ElementsMatch(t, tt.wantPaths, allResourcePaths(t, db))
			}
			assert.Equal(t, tt.wantEvents, ss.events)
			if tt.wantStatus == http.StatusOK {
				list := decodeResourceListResponse(t, w)
				assert.ElementsMatch(t, tt.wantResponsePaths, resourcePaths(list.Items))
			}
			for _, key := range tt.wantMissingBlobs {
				_, err := os.Lstat(resources.BlobPath(dataDir, hashes[key]))
				assert.True(t, os.IsNotExist(err), "blob for %q should not exist", key)
			}
			for _, key := range tt.wantPresentBlobs {
				_, err := os.Lstat(resources.BlobPath(dataDir, hashes[key]))
				assert.NoError(t, err, "blob for %q should exist", key)
			}
		})
	}
}

func TestResourceService_UploadResourcesNonMultipartBodyReturnsBadRequest(t *testing.T) {
	db := setupResourceServiceTestDB(t)
	dataDir := t.TempDir()
	ss := &captureSubscriptions{}
	svc := NewResourceService(db, dataDir, ss)

	body := bytes.NewBufferString(`{"hello":"world"}`)
	c, w := newResourceTestContext(http.MethodPost, "/resources/", body, []string{"resources.upload"})
	c.Request.Header.Set("Content-Type", "application/json")

	svc.UploadResources(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, ss.events)
	assert.Equal(t, 0, countResourceBlobs(t, dataDir))
}

func TestResourceService_UploadResourcesDeduplicatesBlobs(t *testing.T) {
	db := setupResourceServiceTestDB(t)
	dataDir := t.TempDir()
	ss := &captureSubscriptions{}
	svc := NewResourceService(db, dataDir, ss)

	body, contentType := multipartUploadBody(t, []uploadTestFile{
		{name: "a.txt", content: "shared"},
		{name: "b.txt", content: "shared"},
	})
	c, w := newResourceTestContext(http.MethodPost, "/resources/", body, []string{"resources.upload"})
	c.Request.Header.Set("Content-Type", contentType)

	svc.UploadResources(c)

	require.Equal(t, http.StatusOK, w.Code)
	list := decodeResourceListResponse(t, w)
	assert.ElementsMatch(t, []string{"a.txt", "b.txt"}, resourcePaths(list.Items))

	var rows []models.UserResource
	require.NoError(t, db.Order("path ASC").Find(&rows).Error)
	require.Len(t, rows, 2)
	expectedHash := md5HexForService("shared")
	for _, row := range rows {
		assert.Equal(t, expectedHash, row.Hash, "row %q must reference shared blob hash", row.Path)
	}
	assert.Equal(t, 1, countResourceBlobs(t, dataDir), "duplicate-content uploads must reuse a single blob file")

	assert.ElementsMatch(t,
		[]resourceEvent{{action: "added", path: "a.txt"}, {action: "added", path: "b.txt"}},
		ss.events)
}

func TestResourceService_UploadResourcesPreservesExistingBlobOnDBFailure(t *testing.T) {
	db := setupResourceServiceTestDB(t)
	dataDir := t.TempDir()
	ss := &captureSubscriptions{}
	svc := NewResourceService(db, dataDir, ss)
	hash := md5HexForService("shared-existing")
	writeResourceBlob(t, dataDir, hash, "shared-existing")
	seedResource(t, db, models.UserResource{
		UserID: 1,
		Hash:   hash,
		Name:   "existing.txt",
		Path:   "existing.txt",
		Size:   int64(len("shared-existing")),
	})

	body, contentType := multipartUploadBody(t, []uploadTestFile{
		{name: "existing.txt", content: "different"},
	})
	c, w := newResourceTestContext(http.MethodPost, "/resources/", body, []string{"resources.upload"})
	c.Request.Header.Set("Content-Type", contentType)

	svc.UploadResources(c)

	require.Equal(t, http.StatusConflict, w.Code)
	_, err := os.Lstat(resources.BlobPath(dataDir, hash))
	assert.NoError(t, err, "existing blob must not be removed by orphan cleanup on conflict")
	assert.Empty(t, ss.events)
}

func TestResourceService_DownloadResourceScenarios(t *testing.T) {
	type seed struct {
		userID  uint64
		path    string
		isDir   bool
		content string
		// skipBlobWrite leaves the DB row intact but does not write the blob file.
		skipBlobWrite bool
	}

	tests := []struct {
		name             string
		seeds            []seed
		path             string
		privs            []string
		uid              uint64
		wantStatus       int
		wantBody         string
		wantContentType  string
		wantDispContains string
		wantZipEntries   map[string]string
	}{
		{
			name:             "download single file with download privilege",
			seeds:            []seed{{path: "report.txt", content: "payload"}},
			path:             "report.txt",
			privs:            []string{"resources.download"},
			wantStatus:       http.StatusOK,
			wantBody:         "payload",
			wantDispContains: "report.txt",
		},
		{
			name:             "admin can download other user's file",
			seeds:            []seed{{userID: 2, path: "report.txt", content: "alien"}},
			path:             "report.txt",
			privs:            []string{"resources.admin"},
			wantStatus:       http.StatusOK,
			wantBody:         "alien",
			wantDispContains: "report.txt",
		},
		{
			name:       "non-admin cannot download another user's file",
			seeds:      []seed{{userID: 2, path: "report.txt", content: "alien"}},
			path:       "report.txt",
			privs:      []string{"resources.download"},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "download directory returns zip archive with relative paths",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a-data"},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/b.txt", content: "b-data"},
			},
			path:             "docs",
			privs:            []string{"resources.download"},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "docs.zip",
			wantZipEntries: map[string]string{
				"a.txt":     "a-data",
				"sub/b.txt": "b-data",
			},
		},
		{
			name:             "download empty directory returns empty zip archive",
			seeds:            []seed{{path: "docs", isDir: true}},
			path:             "docs",
			privs:            []string{"resources.download"},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "docs.zip",
			wantZipEntries:   map[string]string{},
		},
		{
			name: "download directory containing only sub-directories yields empty zip",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/empty", isDir: true},
			},
			path:             "docs",
			privs:            []string{"resources.download"},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "docs.zip",
			wantZipEntries:   map[string]string{},
		},
		{
			name:       "missing privilege returns forbidden",
			seeds:      []seed{{path: "report.txt", content: "payload"}},
			path:       "report.txt",
			privs:      []string{"resources.view"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "empty path returns bad request",
			seeds:      []seed{{path: "report.txt", content: "payload"}},
			path:       "",
			privs:      []string{"resources.download"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid path returns bad request",
			seeds:      []seed{{path: "report.txt", content: "payload"}},
			path:       "../escape",
			privs:      []string{"resources.download"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing resource returns not found",
			path:       "missing.txt",
			privs:      []string{"resources.download"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "blob missing on disk returns not found",
			seeds:      []seed{{path: "ghost.txt", content: "body", skipBlobWrite: true}},
			path:       "ghost.txt",
			privs:      []string{"resources.download"},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupResourceServiceTestDB(t)
			dataDir := t.TempDir()
			svc := NewResourceService(db, dataDir, nil)
			for _, rec := range tt.seeds {
				userID := rec.userID
				if userID == 0 {
					userID = 1
				}
				seeded := models.UserResource{
					UserID: userID,
					Name:   filepath.Base(rec.path),
					Path:   rec.path,
					IsDir:  rec.isDir,
				}
				if !rec.isDir {
					seeded.Hash = md5HexForService(rec.content)
					seeded.Size = int64(len(rec.content))
					if !rec.skipBlobWrite {
						writeResourceBlob(t, dataDir, seeded.Hash, rec.content)
					}
				}
				seedResource(t, db, seeded)
			}

			target := "/resources/download"
			if tt.path != "" {
				target += "?path=" + tt.path
			}
			uid := tt.uid
			if uid == 0 {
				uid = 1
			}
			c, w := newResourceTestContextWithUID(http.MethodGet, target, nil, tt.privs, uid)

			svc.DownloadResource(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus != http.StatusOK {
				return
			}
			if tt.wantContentType != "" {
				assert.Equal(t, tt.wantContentType, w.Header().Get("Content-Type"))
			}
			if tt.wantDispContains != "" {
				assert.Contains(t, w.Header().Get("Content-Disposition"), tt.wantDispContains)
			}
			if tt.wantZipEntries != nil {
				zr, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
				require.NoError(t, err)
				got := map[string]string{}
				for _, f := range zr.File {
					rc, err := f.Open()
					require.NoError(t, err)
					data, err := io.ReadAll(rc)
					rc.Close()
					require.NoError(t, err)
					got[f.Name] = string(data)
				}
				assert.Equal(t, tt.wantZipEntries, got)
			} else if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestResourceService_DeleteResourceScenarios(t *testing.T) {
	type seed struct {
		userID  uint64
		path    string
		isDir   bool
		content string
		hash    string
	}

	tests := []struct {
		name              string
		seeds             []seed
		targetPath        string
		privs             []string
		wantStatus        int
		wantPaths         []string
		wantResponsePaths []string
		wantEvents        []resourceEvent
		wantMissingBlobs  []string
		wantPresentBlobs  []string
	}{
		{
			name:              "delete file removes row and orphan blob",
			seeds:             []seed{{path: "report.txt", content: "payload"}},
			targetPath:        "report.txt",
			privs:             []string{"resources.delete"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"report.txt"},
			wantEvents:        []resourceEvent{{action: "deleted", path: "report.txt"}},
			wantMissingBlobs:  []string{"payload"},
		},
		{
			name: "delete directory removes recursive tree and orphan blobs",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/b.txt", content: "b"},
				{path: "other.txt", content: "other"},
			},
			targetPath:        "docs",
			privs:             []string{"resources.delete"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"other.txt"},
			wantResponsePaths: []string{"docs", "docs/a.txt", "docs/sub", "docs/sub/b.txt"},
			wantEvents: []resourceEvent{
				{action: "deleted", path: "docs"},
				{action: "deleted", path: "docs/a.txt"},
				{action: "deleted", path: "docs/sub"},
				{action: "deleted", path: "docs/sub/b.txt"},
			},
			wantMissingBlobs: []string{"a", "b"},
			wantPresentBlobs: []string{"other"},
		},
		{
			name: "delete file keeps shared blob",
			seeds: []seed{
				{path: "a.txt", content: "same", hash: "shared"},
				{path: "b.txt", content: "same", hash: "shared"},
			},
			targetPath:        "a.txt",
			privs:             []string{"resources.delete"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"b.txt"},
			wantResponsePaths: []string{"a.txt"},
			wantEvents:        []resourceEvent{{action: "deleted", path: "a.txt"}},
			wantPresentBlobs:  []string{"shared"},
		},
		{
			name: "admin delete is still scoped to current user writes",
			seeds: []seed{
				{userID: 1, path: "own.txt", content: "own"},
				{userID: 2, path: "own.txt", content: "other"},
			},
			targetPath:        "own.txt",
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"own.txt"},
			wantResponsePaths: []string{"own.txt"},
			wantEvents:        []resourceEvent{{action: "deleted", path: "own.txt"}},
			wantMissingBlobs:  []string{"own"},
			wantPresentBlobs:  []string{"other"},
		},
		{
			name:              "delete empty directory removes only directory record",
			seeds:             []seed{{path: "docs", isDir: true}},
			targetPath:        "docs",
			privs:             []string{"resources.delete"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{},
			wantResponsePaths: []string{"docs"},
			wantEvents:        []resourceEvent{{action: "deleted", path: "docs"}},
		},
		{
			name: "delete directory keeps shared blob referenced from outside",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "shared", hash: "shared"},
				{path: "outside.txt", content: "shared", hash: "shared"},
			},
			targetPath:        "docs",
			privs:             []string{"resources.delete"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"outside.txt"},
			wantResponsePaths: []string{"docs", "docs/a.txt"},
			wantEvents: []resourceEvent{
				{action: "deleted", path: "docs"},
				{action: "deleted", path: "docs/a.txt"},
			},
			wantPresentBlobs: []string{"shared"},
		},
		{
			name:       "missing path returns not found",
			targetPath: "missing.txt",
			privs:      []string{"resources.delete"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unsafe path returns bad request",
			targetPath: "../evil.txt",
			privs:      []string{"resources.delete"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "absolute path returns bad request",
			targetPath: "/abs/file.txt",
			privs:      []string{"resources.delete"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty path returns bad request",
			targetPath: "",
			privs:      []string{"resources.delete"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:             "missing privilege returns forbidden",
			seeds:            []seed{{path: "report.txt", content: "payload"}},
			targetPath:       "report.txt",
			privs:            []string{"resources.view"},
			wantStatus:       http.StatusForbidden,
			wantPaths:        []string{"report.txt"},
			wantPresentBlobs: []string{"payload"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupResourceServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &captureSubscriptions{}
			svc := NewResourceService(db, dataDir, ss)
			hashes := map[string]string{}
			for _, rec := range tt.seeds {
				userID := rec.userID
				if userID == 0 {
					userID = 1
				}
				seeded := models.UserResource{
					UserID: userID,
					Name:   filepath.Base(rec.path),
					Path:   rec.path,
					IsDir:  rec.isDir,
				}
				if !rec.isDir {
					hashKey := rec.content
					if rec.hash != "" {
						hashKey = rec.hash
					}
					seeded.Hash = md5HexForService(hashKey)
					seeded.Size = int64(len(rec.content))
					writeResourceBlob(t, dataDir, seeded.Hash, rec.content)
					hashes[hashKey] = seeded.Hash
				}
				seedResource(t, db, seeded)
			}

			c, w := newResourceTestContext(http.MethodDelete, "/resources/?path="+tt.targetPath, nil, tt.privs)
			svc.DeleteResource(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantPaths != nil {
				assert.ElementsMatch(t, tt.wantPaths, allResourcePaths(t, db))
			}
			assert.Equal(t, tt.wantEvents, ss.events)
			if tt.wantStatus == http.StatusOK {
				list := decodeResourceListResponse(t, w)
				assert.ElementsMatch(t, tt.wantResponsePaths, resourcePaths(list.Items))
			}
			for _, key := range tt.wantMissingBlobs {
				_, err := os.Lstat(resources.BlobPath(dataDir, hashes[key]))
				assert.True(t, os.IsNotExist(err), "blob for %q should be removed", key)
			}
			for _, key := range tt.wantPresentBlobs {
				_, err := os.Lstat(resources.BlobPath(dataDir, hashes[key]))
				assert.NoError(t, err, "blob for %q should still exist", key)
			}
		})
	}
}

func TestResourceService_CopyResourceScenarios(t *testing.T) {
	type seed struct {
		path    string
		isDir   bool
		content string
	}
	type copyRequest struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
		Force       bool   `json:"force,omitempty"`
	}

	tests := []struct {
		name                 string
		seeds                []seed
		req                  copyRequest
		privs                []string
		wantStatus           int
		wantPaths            []string
		wantResponsePaths    []string
		wantEvents           []resourceEvent
		wantDeletedBlobTexts []string
	}{
		{
			name:              "file to absent path adds copy",
			seeds:             []seed{{path: "a.txt", content: "src"}},
			req:               copyRequest{Source: "a.txt", Destination: "copies/a.txt"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"a.txt", "copies", "copies/a.txt"},
			wantResponsePaths: []string{"copies", "copies/a.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "copies"}, {action: "added", path: "copies/a.txt"}},
		},
		{
			name:              "admin can copy file",
			seeds:             []seed{{path: "a.txt", content: "src"}},
			req:               copyRequest{Source: "a.txt", Destination: "b.txt"},
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"a.txt", "b.txt"},
			wantResponsePaths: []string{"b.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "b.txt"}},
		},
		{
			name:              "trailing slash on absent destination treats target as directory",
			seeds:             []seed{{path: "a.txt", content: "src"}},
			req:               copyRequest{Source: "a.txt", Destination: "docs/"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"a.txt", "docs", "docs/a.txt"},
			wantResponsePaths: []string{"docs", "docs/a.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "docs"}, {action: "added", path: "docs/a.txt"}},
		},
		{
			name:              "trailing slash with existing directory copies inside",
			seeds:             []seed{{path: "a.txt", content: "src"}, {path: "docs", isDir: true}},
			req:               copyRequest{Source: "a.txt", Destination: "docs/"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"a.txt", "docs", "docs/a.txt"},
			wantResponsePaths: []string{"docs/a.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "docs/a.txt"}},
		},
		{
			name:       "trailing slash with existing file conflicts without force",
			seeds:      []seed{{path: "a.txt", content: "src"}, {path: "docs", content: "blocking"}},
			req:        copyRequest{Source: "a.txt", Destination: "docs/"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "docs"},
		},
		{
			name:                 "trailing slash with existing file replaces it with directory under force",
			seeds:                []seed{{path: "a.txt", content: "src"}, {path: "docs", content: "blocking"}},
			req:                  copyRequest{Source: "a.txt", Destination: "docs/", Force: true},
			wantStatus:           http.StatusOK,
			wantPaths:            []string{"a.txt", "docs", "docs/a.txt"},
			wantResponsePaths:    []string{"docs", "docs/a.txt"},
			wantEvents:           []resourceEvent{{action: "deleted", path: "docs"}, {action: "added", path: "docs"}, {action: "added", path: "docs/a.txt"}},
			wantDeletedBlobTexts: []string{"blocking"},
		},
		{
			name:       "missing privilege returns forbidden",
			seeds:      []seed{{path: "a.txt", content: "src"}},
			req:        copyRequest{Source: "a.txt", Destination: "b.txt"},
			privs:      []string{"resources.view"},
			wantStatus: http.StatusForbidden,
			wantPaths:  []string{"a.txt"},
		},
		{
			name:       "invalid source path returns bad request",
			seeds:      []seed{{path: "a.txt", content: "src"}},
			req:        copyRequest{Source: "../a.txt", Destination: "b.txt"},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"a.txt"},
		},
		{
			name:       "invalid destination path returns bad request",
			seeds:      []seed{{path: "a.txt", content: "src"}},
			req:        copyRequest{Source: "a.txt", Destination: "/abs"},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"a.txt"},
		},
		{
			name:       "missing source returns not found",
			req:        copyRequest{Source: "ghost.txt", Destination: "copy.txt"},
			wantStatus: http.StatusNotFound,
			wantPaths:  []string{},
		},
		{
			name:              "file to absent three-level nested path creates parents",
			seeds:             []seed{{path: "a.txt", content: "src"}},
			req:               copyRequest{Source: "a.txt", Destination: "one/two/three/a.txt"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"a.txt", "one", "one/two", "one/two/three", "one/two/three/a.txt"},
			wantResponsePaths: []string{"one", "one/two", "one/two/three", "one/two/three/a.txt"},
			wantEvents: []resourceEvent{
				{action: "added", path: "one"},
				{action: "added", path: "one/two"},
				{action: "added", path: "one/two/three"},
				{action: "added", path: "one/two/three/a.txt"},
			},
		},
		{
			name:       "file to existing file without force conflicts",
			seeds:      []seed{{path: "a.txt", content: "src"}, {path: "b.txt", content: "dst"}},
			req:        copyRequest{Source: "a.txt", Destination: "b.txt"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "b.txt"},
		},
		{
			name:                 "file to existing file with force overwrites destination",
			seeds:                []seed{{path: "a.txt", content: "src"}, {path: "b.txt", content: "dst"}},
			req:                  copyRequest{Source: "a.txt", Destination: "b.txt", Force: true},
			wantStatus:           http.StatusOK,
			wantPaths:            []string{"a.txt", "b.txt"},
			wantResponsePaths:    []string{"b.txt"},
			wantEvents:           []resourceEvent{{action: "deleted", path: "b.txt"}, {action: "updated", path: "b.txt"}},
			wantDeletedBlobTexts: []string{"dst"},
		},
		{
			name:              "file to existing directory copies inside directory",
			seeds:             []seed{{path: "a.txt", content: "src"}, {path: "docs", isDir: true}},
			req:               copyRequest{Source: "a.txt", Destination: "docs"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"a.txt", "docs", "docs/a.txt"},
			wantResponsePaths: []string{"docs/a.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "docs/a.txt"}},
		},
		{
			name:       "file to existing directory child without force conflicts",
			seeds:      []seed{{path: "a.txt", content: "src"}, {path: "docs", isDir: true}, {path: "docs/a.txt", content: "dst"}},
			req:        copyRequest{Source: "a.txt", Destination: "docs"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "docs", "docs/a.txt"},
		},
		{
			name:                 "file to existing directory child with force overwrites child",
			seeds:                []seed{{path: "a.txt", content: "src"}, {path: "docs", isDir: true}, {path: "docs/a.txt", content: "dst"}},
			req:                  copyRequest{Source: "a.txt", Destination: "docs", Force: true},
			wantStatus:           http.StatusOK,
			wantPaths:            []string{"a.txt", "docs", "docs/a.txt"},
			wantResponsePaths:    []string{"docs/a.txt"},
			wantEvents:           []resourceEvent{{action: "deleted", path: "docs/a.txt"}, {action: "updated", path: "docs/a.txt"}},
			wantDeletedBlobTexts: []string{"dst"},
		},
		{
			name:       "file to existing directory child directory conflicts",
			seeds:      []seed{{path: "a.txt", content: "src"}, {path: "docs", isDir: true}, {path: "docs/a.txt", isDir: true}},
			req:        copyRequest{Source: "a.txt", Destination: "docs", Force: true},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "docs", "docs/a.txt"},
		},
		{
			name: "directory to absent path adds whole tree",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/b.txt", content: "b"},
			},
			req:               copyRequest{Source: "docs", Destination: "copies/docs"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"docs", "docs/a.txt", "docs/sub", "docs/sub/b.txt", "copies", "copies/docs", "copies/docs/a.txt", "copies/docs/sub", "copies/docs/sub/b.txt"},
			wantResponsePaths: []string{"copies", "copies/docs", "copies/docs/a.txt", "copies/docs/sub", "copies/docs/sub/b.txt"},
			wantEvents: []resourceEvent{
				{action: "added", path: "copies"},
				{action: "added", path: "copies/docs"},
				{action: "added", path: "copies/docs/a.txt"},
				{action: "added", path: "copies/docs/sub"},
				{action: "added", path: "copies/docs/sub/b.txt"},
			},
		},
		{
			name:                 "directory to existing file with force replaces file with directory",
			seeds:                []seed{{path: "docs", isDir: true}, {path: "target", content: "file"}},
			req:                  copyRequest{Source: "docs", Destination: "target", Force: true},
			wantStatus:           http.StatusOK,
			wantPaths:            []string{"docs", "target"},
			wantResponsePaths:    []string{"target"},
			wantEvents:           []resourceEvent{{action: "deleted", path: "target"}, {action: "added", path: "target"}},
			wantDeletedBlobTexts: []string{"file"},
		},
		{
			name:       "directory to existing directory without force conflicts",
			seeds:      []seed{{path: "docs", isDir: true}, {path: "archive", isDir: true}},
			req:        copyRequest{Source: "docs", Destination: "archive"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"archive", "docs"},
		},
		{
			name: "directory to existing directory with force merges",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/b.txt", content: "b"},
				{path: "archive", isDir: true},
			},
			req:               copyRequest{Source: "docs", Destination: "archive", Force: true},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"docs", "docs/a.txt", "docs/sub", "docs/sub/b.txt", "archive", "archive/a.txt", "archive/sub", "archive/sub/b.txt"},
			wantResponsePaths: []string{"archive/a.txt", "archive/sub", "archive/sub/b.txt"},
			wantEvents: []resourceEvent{
				{action: "added", path: "archive/a.txt"},
				{action: "added", path: "archive/sub"},
				{action: "added", path: "archive/sub/b.txt"},
			},
		},
		{
			name: "directory merge with existing file overwrites file",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "src-a"},
				{path: "archive", isDir: true},
				{path: "archive/a.txt", content: "dst-a"},
			},
			req:                  copyRequest{Source: "docs", Destination: "archive", Force: true},
			wantStatus:           http.StatusOK,
			wantPaths:            []string{"docs", "docs/a.txt", "archive", "archive/a.txt"},
			wantResponsePaths:    []string{"archive/a.txt"},
			wantEvents:           []resourceEvent{{action: "deleted", path: "archive/a.txt"}, {action: "updated", path: "archive/a.txt"}},
			wantDeletedBlobTexts: []string{"dst-a"},
		},
		{
			name: "directory merge with existing subdirectory keeps destination directory",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/a.txt", content: "a"},
				{path: "archive", isDir: true},
				{path: "archive/sub", isDir: true},
			},
			req:               copyRequest{Source: "docs", Destination: "archive", Force: true},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"docs", "docs/sub", "docs/sub/a.txt", "archive", "archive/sub", "archive/sub/a.txt"},
			wantResponsePaths: []string{"archive/sub", "archive/sub/a.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "archive/sub/a.txt"}, {action: "updated", path: "archive/sub"}},
		},
		{
			name:       "directory merge file over existing directory conflicts",
			seeds:      []seed{{path: "docs", isDir: true}, {path: "docs/a.txt", content: "a"}, {path: "archive", isDir: true}, {path: "archive/a.txt", isDir: true}},
			req:        copyRequest{Source: "docs", Destination: "archive", Force: true},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"archive", "archive/a.txt", "docs", "docs/a.txt"},
		},
		{
			name:       "directory into itself is invalid",
			seeds:      []seed{{path: "docs", isDir: true}},
			req:        copyRequest{Source: "docs", Destination: "docs/archive", Force: true},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"docs"},
		},
		{
			name:       "same source and destination is invalid",
			seeds:      []seed{{path: "a.txt", content: "a"}},
			req:        copyRequest{Source: "a.txt", Destination: "a.txt", Force: true},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"a.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupResourceServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &captureSubscriptions{}
			svc := NewResourceService(db, dataDir, ss)
			deletedHashes := map[string]string{}
			for _, rec := range tt.seeds {
				seeded := models.UserResource{
					UserID: 1,
					Name:   filepath.Base(rec.path),
					Path:   rec.path,
					IsDir:  rec.isDir,
				}
				if !rec.isDir {
					seeded.Hash = md5HexForService(rec.content)
					seeded.Size = int64(len(rec.content))
					writeResourceBlob(t, dataDir, seeded.Hash, rec.content)
					deletedHashes[rec.content] = seeded.Hash
				}
				seedResource(t, db, seeded)
			}

			bodyBytes, err := json.Marshal(tt.req)
			require.NoError(t, err)
			privs := tt.privs
			if privs == nil {
				privs = []string{"resources.edit"}
			}
			c, w := newResourceTestContext(http.MethodPost, "/resources/copy", bytes.NewBuffer(bodyBytes), privs)
			c.Request.Header.Set("Content-Type", "application/json")

			svc.CopyResource(c)

			require.Equal(t, tt.wantStatus, w.Code)
			assert.ElementsMatch(t, tt.wantPaths, allResourcePaths(t, db))
			assert.Equal(t, tt.wantEvents, ss.events)
			if tt.wantStatus == http.StatusOK {
				list := decodeResourceListResponse(t, w)
				assert.ElementsMatch(t, tt.wantResponsePaths, resourcePaths(list.Items))
			}
			for _, content := range tt.wantDeletedBlobTexts {
				_, err := os.Lstat(resources.BlobPath(dataDir, deletedHashes[content]))
				assert.True(t, os.IsNotExist(err), "blob for %q should be removed", content)
			}
		})
	}
}

func TestResourceService_CopyResourceMalformedJSONReturnsBadRequest(t *testing.T) {
	db := setupResourceServiceTestDB(t)
	dataDir := t.TempDir()
	ss := &captureSubscriptions{}
	svc := NewResourceService(db, dataDir, ss)
	seedResource(t, db, models.UserResource{UserID: 1, Name: "a.txt", Path: "a.txt", Hash: md5HexForService("src"), Size: 3})

	c, w := newResourceTestContext(
		http.MethodPost,
		"/resources/copy",
		bytes.NewBufferString(`{not valid json`),
		[]string{"resources.edit"},
	)
	c.Request.Header.Set("Content-Type", "application/json")

	svc.CopyResource(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, ss.events)
	assert.ElementsMatch(t, []string{"a.txt"}, allResourcePaths(t, db))
}

func TestResourceService_CopyResourceFileReusesBlob(t *testing.T) {
	db := setupResourceServiceTestDB(t)
	dataDir := t.TempDir()
	ss := &captureSubscriptions{}
	svc := NewResourceService(db, dataDir, ss)
	hash := md5HexForService("payload")
	writeResourceBlob(t, dataDir, hash, "payload")
	seedResource(t, db, models.UserResource{UserID: 1, Name: "src.txt", Path: "src.txt", Hash: hash, Size: 7})

	body, err := json.Marshal(map[string]any{"source": "src.txt", "destination": "dst.txt"})
	require.NoError(t, err)
	c, w := newResourceTestContext(http.MethodPost, "/resources/copy", bytes.NewBuffer(body), []string{"resources.edit"})
	c.Request.Header.Set("Content-Type", "application/json")

	svc.CopyResource(c)

	require.Equal(t, http.StatusOK, w.Code)

	var rows []models.UserResource
	require.NoError(t, db.Order("path ASC").Find(&rows).Error)
	require.Len(t, rows, 2)
	for _, row := range rows {
		assert.Equal(t, hash, row.Hash, "row %q must reuse source blob hash", row.Path)
	}
	assert.Equal(t, 1, countResourceBlobs(t, dataDir), "copying a file must not create a new blob on disk")
}

func TestResourceService_MoveResourceScenarios(t *testing.T) {
	type seed struct {
		path    string
		isDir   bool
		content string
	}
	type moveRequest struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
		Force       bool   `json:"force,omitempty"`
	}

	tests := []struct {
		name                 string
		seeds                []seed
		req                  moveRequest
		privs                []string
		wantStatus           int
		wantPaths            []string
		wantResponsePaths    []string
		wantEvents           []resourceEvent
		wantDeletedBlobTexts []string
	}{
		{
			name:              "file to absent path updates source",
			seeds:             []seed{{path: "a.txt", content: "src"}},
			req:               moveRequest{Source: "a.txt", Destination: "b.txt"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"b.txt"},
			wantResponsePaths: []string{"b.txt"},
			wantEvents:        []resourceEvent{{action: "updated", path: "b.txt"}},
		},
		{
			name:              "admin can move file",
			seeds:             []seed{{path: "a.txt", content: "src"}},
			req:               moveRequest{Source: "a.txt", Destination: "b.txt"},
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"b.txt"},
			wantResponsePaths: []string{"b.txt"},
			wantEvents:        []resourceEvent{{action: "updated", path: "b.txt"}},
		},
		{
			name:              "trailing slash on absent destination treats target as directory",
			seeds:             []seed{{path: "a.txt", content: "src"}},
			req:               moveRequest{Source: "a.txt", Destination: "docs/"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"docs", "docs/a.txt"},
			wantResponsePaths: []string{"docs/a.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "docs"}, {action: "updated", path: "docs/a.txt"}},
		},
		{
			name:              "trailing slash with existing directory moves inside",
			seeds:             []seed{{path: "a.txt", content: "src"}, {path: "docs", isDir: true}},
			req:               moveRequest{Source: "a.txt", Destination: "docs/"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"docs", "docs/a.txt"},
			wantResponsePaths: []string{"docs/a.txt"},
			wantEvents:        []resourceEvent{{action: "updated", path: "docs/a.txt"}},
		},
		{
			name:       "trailing slash with existing file conflicts without force",
			seeds:      []seed{{path: "a.txt", content: "src"}, {path: "docs", content: "blocking"}},
			req:        moveRequest{Source: "a.txt", Destination: "docs/"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "docs"},
		},
		{
			name:                 "trailing slash with existing file replaces it with directory under force",
			seeds:                []seed{{path: "a.txt", content: "src"}, {path: "docs", content: "blocking"}},
			req:                  moveRequest{Source: "a.txt", Destination: "docs/", Force: true},
			wantStatus:           http.StatusOK,
			wantPaths:            []string{"docs", "docs/a.txt"},
			wantResponsePaths:    []string{"docs/a.txt"},
			wantEvents:           []resourceEvent{{action: "deleted", path: "docs"}, {action: "added", path: "docs"}, {action: "updated", path: "docs/a.txt"}},
			wantDeletedBlobTexts: []string{"blocking"},
		},
		{
			name:       "missing privilege returns forbidden",
			seeds:      []seed{{path: "a.txt", content: "src"}},
			req:        moveRequest{Source: "a.txt", Destination: "b.txt"},
			privs:      []string{"resources.view"},
			wantStatus: http.StatusForbidden,
			wantPaths:  []string{"a.txt"},
		},
		{
			name:       "invalid source path returns bad request",
			seeds:      []seed{{path: "a.txt", content: "src"}},
			req:        moveRequest{Source: "../a.txt", Destination: "b.txt"},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"a.txt"},
		},
		{
			name:       "invalid destination path returns bad request",
			seeds:      []seed{{path: "a.txt", content: "src"}},
			req:        moveRequest{Source: "a.txt", Destination: "/abs"},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"a.txt"},
		},
		{
			name:       "missing source returns not found",
			req:        moveRequest{Source: "ghost.txt", Destination: "copy.txt"},
			wantStatus: http.StatusNotFound,
			wantPaths:  []string{},
		},
		{
			name:              "file to absent three-level nested path creates parents",
			seeds:             []seed{{path: "a.txt", content: "src"}},
			req:               moveRequest{Source: "a.txt", Destination: "one/two/three/a.txt"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"one", "one/two", "one/two/three", "one/two/three/a.txt"},
			wantResponsePaths: []string{"one/two/three/a.txt"},
			wantEvents: []resourceEvent{
				{action: "added", path: "one"},
				{action: "added", path: "one/two"},
				{action: "added", path: "one/two/three"},
				{action: "updated", path: "one/two/three/a.txt"},
			},
		},
		{
			name:       "file to existing file without force conflicts",
			seeds:      []seed{{path: "a.txt", content: "src"}, {path: "b.txt", content: "dst"}},
			req:        moveRequest{Source: "a.txt", Destination: "b.txt"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "b.txt"},
		},
		{
			name:                 "file to existing file with force overwrites destination",
			seeds:                []seed{{path: "a.txt", content: "src"}, {path: "b.txt", content: "dst"}},
			req:                  moveRequest{Source: "a.txt", Destination: "b.txt", Force: true},
			wantStatus:           http.StatusOK,
			wantPaths:            []string{"b.txt"},
			wantResponsePaths:    []string{"b.txt"},
			wantEvents:           []resourceEvent{{action: "deleted", path: "b.txt"}, {action: "updated", path: "b.txt"}},
			wantDeletedBlobTexts: []string{"dst"},
		},
		{
			name:              "file to existing directory moves inside directory",
			seeds:             []seed{{path: "a.txt", content: "src"}, {path: "docs", isDir: true}},
			req:               moveRequest{Source: "a.txt", Destination: "docs"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"docs", "docs/a.txt"},
			wantResponsePaths: []string{"docs/a.txt"},
			wantEvents:        []resourceEvent{{action: "updated", path: "docs/a.txt"}},
		},
		{
			name:       "file to existing directory child without force conflicts",
			seeds:      []seed{{path: "a.txt", content: "src"}, {path: "docs", isDir: true}, {path: "docs/a.txt", content: "dst"}},
			req:        moveRequest{Source: "a.txt", Destination: "docs"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "docs", "docs/a.txt"},
		},
		{
			name:                 "file to existing directory child with force overwrites child",
			seeds:                []seed{{path: "a.txt", content: "src"}, {path: "docs", isDir: true}, {path: "docs/a.txt", content: "dst"}},
			req:                  moveRequest{Source: "a.txt", Destination: "docs", Force: true},
			wantStatus:           http.StatusOK,
			wantPaths:            []string{"docs", "docs/a.txt"},
			wantResponsePaths:    []string{"docs/a.txt"},
			wantEvents:           []resourceEvent{{action: "deleted", path: "docs/a.txt"}, {action: "updated", path: "docs/a.txt"}},
			wantDeletedBlobTexts: []string{"dst"},
		},
		{
			name:       "file to existing directory child directory conflicts",
			seeds:      []seed{{path: "a.txt", content: "src"}, {path: "docs", isDir: true}, {path: "docs/a.txt", isDir: true}},
			req:        moveRequest{Source: "a.txt", Destination: "docs", Force: true},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "docs", "docs/a.txt"},
		},
		{
			name: "directory to absent path updates whole tree",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/b.txt", content: "b"},
			},
			req:               moveRequest{Source: "docs", Destination: "archive/docs"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"archive", "archive/docs", "archive/docs/a.txt", "archive/docs/sub", "archive/docs/sub/b.txt"},
			wantResponsePaths: []string{"archive/docs", "archive/docs/a.txt", "archive/docs/sub", "archive/docs/sub/b.txt"},
			wantEvents: []resourceEvent{
				{action: "added", path: "archive"},
				{action: "updated", path: "archive/docs"},
				{action: "updated", path: "archive/docs/a.txt"},
				{action: "updated", path: "archive/docs/sub"},
				{action: "updated", path: "archive/docs/sub/b.txt"},
			},
		},
		{
			name:                 "directory to existing file with force replaces file with directory",
			seeds:                []seed{{path: "docs", isDir: true}, {path: "target", content: "file"}},
			req:                  moveRequest{Source: "docs", Destination: "target", Force: true},
			wantStatus:           http.StatusOK,
			wantPaths:            []string{"target"},
			wantResponsePaths:    []string{"target"},
			wantEvents:           []resourceEvent{{action: "deleted", path: "target"}, {action: "updated", path: "target"}},
			wantDeletedBlobTexts: []string{"file"},
		},
		{
			name:       "directory to existing directory without force conflicts",
			seeds:      []seed{{path: "docs", isDir: true}, {path: "archive", isDir: true}},
			req:        moveRequest{Source: "docs", Destination: "archive"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"archive", "docs"},
		},
		{
			name: "directory to existing directory with force merges",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/b.txt", content: "b"},
				{path: "archive", isDir: true},
			},
			req:               moveRequest{Source: "docs", Destination: "archive", Force: true},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"archive", "archive/a.txt", "archive/sub", "archive/sub/b.txt"},
			wantResponsePaths: []string{"archive/a.txt", "archive/sub", "archive/sub/b.txt"},
			wantEvents: []resourceEvent{
				{action: "updated", path: "archive/a.txt"},
				{action: "updated", path: "archive/sub"},
				{action: "updated", path: "archive/sub/b.txt"},
				{action: "deleted", path: "docs"},
			},
		},
		{
			name: "directory merge with existing file overwrites file",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "src-a"},
				{path: "archive", isDir: true},
				{path: "archive/a.txt", content: "dst-a"},
			},
			req:                  moveRequest{Source: "docs", Destination: "archive", Force: true},
			wantStatus:           http.StatusOK,
			wantPaths:            []string{"archive", "archive/a.txt"},
			wantResponsePaths:    []string{"archive/a.txt"},
			wantEvents:           []resourceEvent{{action: "deleted", path: "archive/a.txt"}, {action: "updated", path: "archive/a.txt"}, {action: "deleted", path: "docs"}},
			wantDeletedBlobTexts: []string{"dst-a"},
		},
		{
			name: "directory merge with existing subdirectory keeps destination directory",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/a.txt", content: "a"},
				{path: "archive", isDir: true},
				{path: "archive/sub", isDir: true},
			},
			req:               moveRequest{Source: "docs", Destination: "archive", Force: true},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"archive", "archive/sub", "archive/sub/a.txt"},
			wantResponsePaths: []string{"archive/sub/a.txt"},
			wantEvents: []resourceEvent{
				{action: "updated", path: "archive/sub/a.txt"},
				{action: "deleted", path: "docs/sub"},
				{action: "deleted", path: "docs"},
			},
		},
		{
			name:       "directory merge file over existing directory conflicts",
			seeds:      []seed{{path: "docs", isDir: true}, {path: "docs/a.txt", content: "a"}, {path: "archive", isDir: true}, {path: "archive/a.txt", isDir: true}},
			req:        moveRequest{Source: "docs", Destination: "archive", Force: true},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"archive", "archive/a.txt", "docs", "docs/a.txt"},
		},
		{
			name:       "directory into itself is invalid",
			seeds:      []seed{{path: "docs", isDir: true}},
			req:        moveRequest{Source: "docs", Destination: "docs/archive", Force: true},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"docs"},
		},
		{
			name:       "same source and destination is invalid",
			seeds:      []seed{{path: "a.txt", content: "a"}},
			req:        moveRequest{Source: "a.txt", Destination: "a.txt", Force: true},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"a.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupResourceServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &captureSubscriptions{}
			svc := NewResourceService(db, dataDir, ss)
			deletedHashes := map[string]string{}
			for _, rec := range tt.seeds {
				seeded := models.UserResource{
					UserID: 1,
					Name:   filepath.Base(rec.path),
					Path:   rec.path,
					IsDir:  rec.isDir,
				}
				if !rec.isDir {
					seeded.Hash = md5HexForService(rec.content)
					seeded.Size = int64(len(rec.content))
					writeResourceBlob(t, dataDir, seeded.Hash, rec.content)
					deletedHashes[rec.content] = seeded.Hash
				}
				seedResource(t, db, seeded)
			}

			bodyBytes, err := json.Marshal(tt.req)
			require.NoError(t, err)
			privs := tt.privs
			if privs == nil {
				privs = []string{"resources.edit"}
			}
			c, w := newResourceTestContext(http.MethodPut, "/resources/move", bytes.NewBuffer(bodyBytes), privs)
			c.Request.Header.Set("Content-Type", "application/json")

			svc.MoveResource(c)

			require.Equal(t, tt.wantStatus, w.Code)
			assert.ElementsMatch(t, tt.wantPaths, allResourcePaths(t, db))
			assert.Equal(t, tt.wantEvents, ss.events)
			if tt.wantStatus == http.StatusOK {
				list := decodeResourceListResponse(t, w)
				assert.ElementsMatch(t, tt.wantResponsePaths, resourcePaths(list.Items))
			}
			for _, content := range tt.wantDeletedBlobTexts {
				_, err := os.Lstat(resources.BlobPath(dataDir, deletedHashes[content]))
				assert.True(t, os.IsNotExist(err), "blob for %q should be removed", content)
			}
		})
	}
}

func TestResourceService_MoveResourceMalformedJSONReturnsBadRequest(t *testing.T) {
	db := setupResourceServiceTestDB(t)
	dataDir := t.TempDir()
	ss := &captureSubscriptions{}
	svc := NewResourceService(db, dataDir, ss)
	seedResource(t, db, models.UserResource{UserID: 1, Name: "a.txt", Path: "a.txt", Hash: md5HexForService("src"), Size: 3})

	c, w := newResourceTestContext(
		http.MethodPut,
		"/resources/move",
		bytes.NewBufferString(`{not valid json`),
		[]string{"resources.edit"},
	)
	c.Request.Header.Set("Content-Type", "application/json")

	svc.MoveResource(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, ss.events)
	assert.ElementsMatch(t, []string{"a.txt"}, allResourcePaths(t, db))
}

func TestResourceService_QueryResources(t *testing.T) {
	db := setupResourceServiceTestDB(t)
	svc := NewResourceService(db, t.TempDir(), nil)

	seedResource(t, db, models.UserResource{UserID: 1, Hash: md5HexForService("root"), Name: "root.txt", Path: "root.txt", Size: 4})
	seedResource(t, db, models.UserResource{UserID: 1, Name: "docs", Path: "docs", IsDir: true})
	seedResource(t, db, models.UserResource{UserID: 1, Hash: md5HexForService("a"), Name: "a.txt", Path: "docs/a.txt", Size: 1})
	seedResource(t, db, models.UserResource{UserID: 1, Name: "sub", Path: "docs/sub", IsDir: true})
	seedResource(t, db, models.UserResource{UserID: 1, Hash: md5HexForService("b"), Name: "b.txt", Path: "docs/sub/b.txt", Size: 1})
	seedResource(t, db, models.UserResource{UserID: 2, Hash: md5HexForService("other"), Name: "other.txt", Path: "other.txt", Size: 5})

	root, err := svc.queryResources(1, false, "", false)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"root.txt", "docs"}, resourcePaths(root))

	docs, err := svc.queryResources(1, false, "docs", false)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"docs", "docs/a.txt", "docs/sub"}, resourcePaths(docs))

	recursive, err := svc.queryResources(1, false, "docs", true)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"docs", "docs/a.txt", "docs/sub", "docs/sub/b.txt"}, resourcePaths(recursive))

	adminRoot, err := svc.queryResources(1, true, "", false)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"root.txt", "docs", "other.txt"}, resourcePaths(adminRoot))
}

func TestResourceService_CleanupOrphanBlobsScenarios(t *testing.T) {
	type seed struct {
		userID  uint64
		path    string
		content string
	}

	tests := []struct {
		name             string
		seeds            []seed
		hashKeys         []string
		extraBlobs       []string
		wantPresentBlobs []string
		wantMissingBlobs []string
	}{
		{
			name:             "no-op when hashes argument is empty",
			seeds:            []seed{{path: "kept.txt", content: "kept"}},
			hashKeys:         nil,
			wantPresentBlobs: []string{"kept"},
		},
		{
			name:             "removes only orphan hashes when mixed with referenced",
			seeds:            []seed{{path: "kept.txt", content: "kept"}},
			extraBlobs:       []string{"orphan"},
			hashKeys:         []string{"kept", "orphan"},
			wantPresentBlobs: []string{"kept"},
			wantMissingBlobs: []string{"orphan"},
		},
		{
			name:             "all referenced hashes are kept",
			seeds:            []seed{{path: "a.txt", content: "alpha"}, {path: "b.txt", content: "beta"}},
			hashKeys:         []string{"alpha", "beta"},
			wantPresentBlobs: []string{"alpha", "beta"},
		},
		{
			name:             "removes all orphan hashes when none are referenced",
			extraBlobs:       []string{"x", "y"},
			hashKeys:         []string{"x", "y"},
			wantMissingBlobs: []string{"x", "y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupResourceServiceTestDB(t)
			dataDir := t.TempDir()
			svc := NewResourceService(db, dataDir, nil)
			hashes := map[string]string{}
			for _, rec := range tt.seeds {
				userID := rec.userID
				if userID == 0 {
					userID = 1
				}
				hash := md5HexForService(rec.content)
				writeResourceBlob(t, dataDir, hash, rec.content)
				seedResource(t, db, models.UserResource{
					UserID: userID,
					Hash:   hash,
					Name:   filepath.Base(rec.path),
					Path:   rec.path,
					Size:   int64(len(rec.content)),
				})
				hashes[rec.content] = hash
			}
			for _, key := range tt.extraBlobs {
				hash := md5HexForService(key)
				writeResourceBlob(t, dataDir, hash, key)
				hashes[key] = hash
			}

			hashSlice := make([]string, 0, len(tt.hashKeys))
			for _, key := range tt.hashKeys {
				hashSlice = append(hashSlice, hashes[key])
			}

			svc.cleanupOrphanBlobs(context.Background(), hashSlice)

			for _, key := range tt.wantPresentBlobs {
				_, err := os.Lstat(resources.BlobPath(dataDir, hashes[key]))
				assert.NoError(t, err, "blob for %q must remain", key)
			}
			for _, key := range tt.wantMissingBlobs {
				_, err := os.Lstat(resources.BlobPath(dataDir, hashes[key]))
				assert.True(t, os.IsNotExist(err), "blob for %q must be removed", key)
			}
		})
	}
}

func TestResourceService_DeleteOrphanBlobScenarios(t *testing.T) {
	t.Run("removes orphan hash blob", func(t *testing.T) {
		db := setupResourceServiceTestDB(t)
		dataDir := t.TempDir()
		svc := NewResourceService(db, dataDir, nil)
		hash := md5HexForService("orphan")
		writeResourceBlob(t, dataDir, hash, "orphan")

		svc.deleteOrphanBlob(context.Background(), hash)

		_, err := os.Lstat(resources.BlobPath(dataDir, hash))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("keeps blob when still referenced", func(t *testing.T) {
		db := setupResourceServiceTestDB(t)
		dataDir := t.TempDir()
		svc := NewResourceService(db, dataDir, nil)
		hash := md5HexForService("kept")
		writeResourceBlob(t, dataDir, hash, "kept")
		seedResource(t, db, models.UserResource{UserID: 1, Hash: hash, Name: "kept.txt", Path: "kept.txt", Size: 4})

		svc.deleteOrphanBlob(context.Background(), hash)

		_, err := os.Lstat(resources.BlobPath(dataDir, hash))
		assert.NoError(t, err)
	})

	t.Run("empty hash is a no-op", func(t *testing.T) {
		db := setupResourceServiceTestDB(t)
		dataDir := t.TempDir()
		svc := NewResourceService(db, dataDir, nil)

		assert.NotPanics(t, func() {
			svc.deleteOrphanBlob(context.Background(), "")
		})
	})
}

func TestResourceService_ConvertResourceToModel(t *testing.T) {
	entry := models.ResourceEntry{
		ID:     42,
		UserID: 7,
		Name:   "report.txt",
		Path:   "docs/report.txt",
		Size:   123,
		IsDir:  false,
	}

	modelResource := convertResourceToModel(entry)

	require.NotNil(t, modelResource)
	assert.Equal(t, int64(entry.ID), modelResource.ID)
	assert.Equal(t, int64(entry.UserID), modelResource.UserID)
	assert.Equal(t, entry.Name, modelResource.Name)
	assert.Equal(t, entry.Path, modelResource.Path)
	assert.Equal(t, int(entry.Size), modelResource.Size)
	assert.Equal(t, entry.IsDir, modelResource.IsDir)
}

func writeResourceBlob(t *testing.T, dataDir, hash, content string) {
	t.Helper()

	blobPath := resources.BlobPath(dataDir, hash)
	require.NoError(t, os.MkdirAll(filepath.Dir(blobPath), 0755))
	require.NoError(t, os.WriteFile(blobPath, []byte(content), 0644))
}

func md5HexForService(content string) string {
	sum := md5.Sum([]byte(content))
	return hex.EncodeToString(sum[:])
}
