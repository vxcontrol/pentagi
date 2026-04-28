package services

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("uid", uint64(1))
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
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, file := range files {
		part, err := writer.CreateFormFile("files", file.name)
		require.NoError(t, err)
		_, err = part.Write([]byte(file.content))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	return &body, writer.FormDataContentType()
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

func TestResourceService_ListResourcesAPI(t *testing.T) {
	db := setupResourceServiceTestDB(t)
	svc := NewResourceService(db, t.TempDir(), nil)
	seedResource(t, db, models.UserResource{UserID: 1, Name: "docs", Path: "docs", IsDir: true})
	seedResource(t, db, models.UserResource{UserID: 1, Hash: md5HexForService("a"), Name: "a.txt", Path: "docs/a.txt", Size: 1})
	seedResource(t, db, models.UserResource{UserID: 1, Hash: md5HexForService("b"), Name: "b.txt", Path: "docs/sub/b.txt", Size: 1})

	c, w := newResourceTestContext(http.MethodGet, "/resources/?path=docs&recursive=true", nil, []string{"resources.view"})

	svc.ListResources(c)

	require.Equal(t, http.StatusOK, w.Code)
	list := decodeResourceListResponse(t, w)
	assert.ElementsMatch(t, []string{"docs", "docs/a.txt", "docs/sub/b.txt"}, resourcePaths(list.Items))
	assert.Equal(t, uint64(3), list.Total)
}

func TestResourceService_MkdirResourceAPI(t *testing.T) {
	db := setupResourceServiceTestDB(t)
	svc := NewResourceService(db, t.TempDir(), nil)
	body := bytes.NewBufferString(`{"path":"docs"}`)
	c, w := newResourceTestContext(http.MethodPost, "/resources/mkdir", body, []string{"resources.edit"})
	c.Request.Header.Set("Content-Type", "application/json")

	svc.MkdirResource(c)

	require.Equal(t, http.StatusOK, w.Code)
	var rec models.UserResource
	require.NoError(t, db.Where("user_id = ? AND path = ?", 1, "docs").First(&rec).Error)
	assert.True(t, rec.IsDir)
	assert.Equal(t, "docs", rec.Name)
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
			name:              "upload into missing nested dir creates parents and files",
			dir:               "docs/sub",
			files:             []uploadTestFile{{name: "a.txt", content: "a"}, {name: "b.txt", content: "b"}},
			privs:             []string{"resources.upload"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"docs", "docs/sub", "docs/sub/a.txt", "docs/sub/b.txt"},
			wantResponsePaths: []string{"docs/sub/a.txt", "docs/sub/b.txt"},
			wantEvents: []resourceEvent{
				{action: "added", path: "docs"},
				{action: "added", path: "docs/sub"},
				{action: "added", path: "docs/sub/a.txt"},
				{action: "added", path: "docs/sub/b.txt"},
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

			body, contentType := multipartUploadBody(t, tt.files)
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

func TestResourceService_DownloadResourceAPI(t *testing.T) {
	db := setupResourceServiceTestDB(t)
	dataDir := t.TempDir()
	svc := NewResourceService(db, dataDir, nil)
	hash := md5HexForService("payload")
	writeResourceBlob(t, dataDir, hash, "payload")
	seedResource(t, db, models.UserResource{UserID: 1, Hash: hash, Name: "report.txt", Path: "report.txt", Size: 7})
	c, w := newResourceTestContext(http.MethodGet, "/resources/download?path=report.txt", nil, []string{"resources.download"})

	svc.DownloadResource(c)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "payload", w.Body.String())
	assert.Contains(t, w.Header().Get("Content-Disposition"), "report.txt")
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
			wantPaths:         []string{"a.txt", "copies/a.txt"},
			wantResponsePaths: []string{"copies/a.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "copies/a.txt"}},
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
			wantPaths:         []string{"docs", "docs/a.txt", "docs/sub", "docs/sub/b.txt", "copies/docs", "copies/docs/a.txt", "copies/docs/sub", "copies/docs/sub/b.txt"},
			wantResponsePaths: []string{"copies/docs", "copies/docs/a.txt", "copies/docs/sub", "copies/docs/sub/b.txt"},
			wantEvents: []resourceEvent{
				{action: "added", path: "copies/docs"},
				{action: "added", path: "copies/docs/a.txt"},
				{action: "added", path: "copies/docs/sub"},
				{action: "added", path: "copies/docs/sub/b.txt"},
			},
		},
		{
			name:       "directory to existing file conflicts",
			seeds:      []seed{{path: "docs", isDir: true}, {path: "target", content: "file"}},
			req:        copyRequest{Source: "docs", Destination: "target", Force: true},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"docs", "target"},
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
			c, w := newResourceTestContext(http.MethodPost, "/resources/copy", bytes.NewBuffer(bodyBytes), []string{"resources.edit"})
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
			wantPaths:         []string{"archive/docs", "archive/docs/a.txt", "archive/docs/sub", "archive/docs/sub/b.txt"},
			wantResponsePaths: []string{"archive/docs", "archive/docs/a.txt", "archive/docs/sub", "archive/docs/sub/b.txt"},
			wantEvents: []resourceEvent{
				{action: "updated", path: "archive/docs"},
				{action: "updated", path: "archive/docs/a.txt"},
				{action: "updated", path: "archive/docs/sub"},
				{action: "updated", path: "archive/docs/sub/b.txt"},
			},
		},
		{
			name:       "directory to existing file conflicts",
			seeds:      []seed{{path: "docs", isDir: true}, {path: "target", content: "file"}},
			req:        moveRequest{Source: "docs", Destination: "target", Force: true},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"docs", "target"},
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
			c, w := newResourceTestContext(http.MethodPut, "/resources/move", bytes.NewBuffer(bodyBytes), []string{"resources.edit"})
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

func TestResourceService_CleanupOrphanBlobs(t *testing.T) {
	db := setupResourceServiceTestDB(t)
	dataDir := t.TempDir()
	svc := NewResourceService(db, dataDir, nil)
	referencedHash := md5HexForService("referenced")
	orphanHash := md5HexForService("orphan")
	writeResourceBlob(t, dataDir, referencedHash, "referenced")
	writeResourceBlob(t, dataDir, orphanHash, "orphan")
	seedResource(t, db, models.UserResource{
		UserID: 1,
		Hash:   referencedHash,
		Name:   "referenced.txt",
		Path:   "referenced.txt",
		Size:   10,
	})

	svc.cleanupOrphanBlobs(context.Background(), []string{referencedHash, orphanHash})

	_, err := os.Lstat(resources.BlobPath(dataDir, referencedHash))
	require.NoError(t, err)
	_, err = os.Lstat(resources.BlobPath(dataDir, orphanHash))
	assert.True(t, os.IsNotExist(err))
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
	assert.Equal(t, entry.ID, modelResource.ID)
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
