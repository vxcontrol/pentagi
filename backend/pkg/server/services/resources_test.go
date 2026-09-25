package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/graph/model"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/resources"
	"pentagi/pkg/server/models"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

type resourcesTransferSeed struct {
	userID  uint64 // 0: the caller
	path    string
	isDir   bool
	content string
}

type resourcesTransferRequest struct {
	Source      string   `json:"source,omitempty"`
	Sources     []string `json:"sources,omitempty"`
	Destination string   `json:"destination"`
	Force       bool     `json:"force,omitempty"`
}

// resourcesTransferCase is one row of the copy and move tables.
type resourcesTransferCase struct {
	name                 string
	seeds                []resourcesTransferSeed
	req                  resourcesTransferRequest
	rawBody              string   // sent instead of req when set
	privs                []string // nil: resources.edit
	wantStatus           int
	wantPaths            []string
	wantResponsePaths    []string
	wantEvents           []resourceEvent
	wantDeletedBlobTexts []string
	wantHashes           map[string]string // path -> content whose md5 the row references
	wantOwners           map[string]uint64
}

// resourcesRunTransferCase also checks that a refused request changes no row.
func resourcesRunTransferCase(
	t *testing.T,
	tt resourcesTransferCase,
	method, target string,
	handle func(*ResourceService, *gin.Context),
) {
	t.Helper()

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
			seeded.Hash = md5HexForService(rec.content)
			seeded.Size = int64(len(rec.content))
			writeResourceBlob(t, dataDir, seeded.Hash, rec.content)
			hashes[rec.content] = seeded.Hash
		}
		seedResource(t, db, seeded)
	}
	before := resourcesRows(t, db)

	body := tt.rawBody
	if body == "" {
		payload, err := json.Marshal(tt.req)
		require.NoError(t, err)
		body = string(payload)
	}
	privs := tt.privs
	if privs == nil {
		privs = []string{"resources.edit"}
	}
	c, w := newResourceTestContext(method, target, bytes.NewBufferString(body), privs)
	c.Request.Header.Set("Content-Type", "application/json")

	handle(svc, c)

	require.Equal(t, tt.wantStatus, w.Code)
	assert.ElementsMatch(t, tt.wantPaths, allResourcePaths(t, db))
	assert.Equal(t, tt.wantEvents, ss.events)
	if tt.wantStatus == http.StatusOK {
		list := decodeResourceListResponse(t, w)
		assert.ElementsMatch(t, tt.wantResponsePaths, resourcePaths(list.Items))
	} else {
		assert.Equal(t, before, resourcesRows(t, db), "a refused request must change no row")
	}
	for _, content := range tt.wantDeletedBlobTexts {
		_, err := os.Lstat(resources.BlobPath(dataDir, hashes[content]))
		assert.True(t, os.IsNotExist(err), "blob for %q should be removed", content)
	}
	resourcesRequireHashes(t, db, tt.wantHashes)
	resourcesRequireOwners(t, db, tt.wantOwners)
	resourcesRequireBlobsMatchRows(t, db, dataDir)
}

// resourcesRows describes every row by what a request may change, in id order.
func resourcesRows(t *testing.T, db *gorm.DB) []string {
	t.Helper()

	var rows []models.UserResource
	require.NoError(t, db.Order("id ASC").Find(&rows).Error)
	described := make([]string, len(rows))
	for i, row := range rows {
		described[i] = fmt.Sprintf("%d %s dir=%t hash=%s owner=%d", row.ID, row.Path, row.IsDir, row.Hash, row.UserID)
	}
	return described
}

func resourcesRequireOwners(t *testing.T, db *gorm.DB, want map[string]uint64) {
	t.Helper()

	for vPath, owner := range want {
		var rows []models.UserResource
		require.NoError(t, db.Where("path = ?", vPath).Find(&rows).Error)
		require.Len(t, rows, 1, "rows at %q", vPath)
		assert.Equal(t, owner, rows[0].UserID, "owner of %q", vPath)
	}
}

// resourcesRequireBlobsMatchRows checks the data directory holds exactly one blob per
// hash a file row references: no orphan, no dangling row, no leftover temp file.
func resourcesRequireBlobsMatchRows(t *testing.T, db *gorm.DB, dataDir string) {
	t.Helper()

	var rows []models.UserResource
	require.NoError(t, db.Find(&rows).Error)
	want := map[string]bool{}
	for _, row := range rows {
		if !row.IsDir {
			want[resources.BlobPath(dataDir, row.Hash)] = true
		}
	}

	got := map[string]bool{}
	require.NoError(t, filepath.Walk(dataDir, func(p string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		if !info.IsDir() {
			got[p] = true
		}
		return nil
	}))

	assert.Equal(t, want, got, "files under the data directory")
}

// resourcesRequireHashes checks each path's row references the blob of the given content.
func resourcesRequireHashes(t *testing.T, db *gorm.DB, want map[string]string) {
	t.Helper()

	for vPath, content := range want {
		var row models.UserResource
		require.NoError(t, db.Where("path = ?", vPath).First(&row).Error)
		assert.Equal(t, md5HexForService(content), row.Hash, "blob referenced by %q", vPath)
	}
}

func TestResources_ListResources_ListsWhatTheCallerMaySee(t *testing.T) {
	type seed struct {
		userID  uint64
		path    string
		isDir   bool
		content string
		updated time.Duration // updated_at relative to now; 0 keeps the insert time
	}

	tests := []struct {
		name              string
		seeds             []seed
		sameStamp         bool // every row carries one updated_at
		path              string
		paths             []string // additional paths for paths[] param
		rawQuery          string   // overrides path/paths/recursive URL building when set
		recursive         bool
		privs             []string
		uid               uint64
		wantStatus        int
		wantResponsePaths []string
		wantSorted        bool              // wantResponsePaths is also the order
		wantOwners        map[string]uint64 // owner of the row listed at a path
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
		{
			name: "two directories via paths[] returns combined deduplicated results",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
				{path: "other", isDir: true},
				{path: "other/b.txt", content: "b"},
			},
			paths:             []string{"docs", "other"},
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"docs", "docs/a.txt", "other", "other/b.txt"},
		},
		{
			name: "path= and paths[] combined return merged results",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
				{path: "extra", isDir: true},
				{path: "extra/c.txt", content: "c"},
			},
			path:              "docs",
			paths:             []string{"extra"},
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"docs", "docs/a.txt", "extra", "extra/c.txt"},
		},
		{
			name: "duplicate paths in paths[] deduplicated",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
			},
			paths:             []string{"docs", "docs"},
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"docs", "docs/a.txt"},
		},
		{
			name: "path= and paths[] with same value deduplicated",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
			},
			path:              "docs",
			paths:             []string{"docs"},
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"docs", "docs/a.txt"},
		},
		{
			name: "whitespace-only paths[] falls back to root listing",
			seeds: []seed{
				{path: "root.txt", content: "r"},
			},
			rawQuery:          "paths[]=%20%20&paths[]=%09",
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"root.txt"},
		},
		{
			name:       "invalid path in paths[] returns bad request",
			rawQuery:   "paths[]=docs&paths[]=../escape",
			privs:      []string{"resources.view"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "output sorted by path across multiple queried paths",
			seeds: []seed{
				{path: "z", isDir: true},
				{path: "z/c.txt", content: "c"},
				{path: "a", isDir: true},
				{path: "a/b.txt", content: "b"},
			},
			paths:             []string{"z", "a"},
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"a", "a/b.txt", "z", "z/c.txt"},
			wantSorted:        true,
		},
		{
			// the parent comes along so a client never draws a dangling node
			name: "listing nested path includes parent directory as ancestor",
			seeds: []seed{
				{path: "base", isDir: true},
				{path: "base/sub", isDir: true},
				{path: "base/sub/file.txt", content: "x"},
			},
			paths:             []string{"base/sub"},
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"base", "base/sub", "base/sub/file.txt"},
		},
		{
			name: "listing deeply nested path includes all ancestor directories",
			seeds: []seed{
				{path: "a", isDir: true},
				{path: "a/b", isDir: true},
				{path: "a/b/c", isDir: true},
				{path: "a/b/c/file.txt", content: "y"},
			},
			paths:             []string{"a/b/c"},
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"a", "a/b", "a/b/c", "a/b/c/file.txt"},
		},
		{
			name: "missing ancestor directory not added to response",
			seeds: []seed{
				// "base" parent is intentionally not seeded
				{path: "base/sub", isDir: true},
				{path: "base/sub/file.txt", content: "x"},
			},
			paths:             []string{"base/sub"},
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"base/sub", "base/sub/file.txt"},
		},
		{
			name: "ancestor directory belonging to another user is not included",
			seeds: []seed{
				{userID: 2, path: "shared", isDir: true},
				{userID: 1, path: "shared/sub", isDir: true},
				{userID: 1, path: "shared/sub/file.txt", content: "x"},
			},
			paths:             []string{"shared/sub"},
			privs:             []string{"resources.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"shared/sub", "shared/sub/file.txt"},
		},
		{
			name: "an administrator's own row wins a root path collision with a newer foreign row",
			seeds: []seed{
				{userID: 2, path: "report.txt", content: "theirs", updated: time.Hour},
				{userID: 1, path: "report.txt", content: "mine"},
			},
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"report.txt"},
			wantOwners:        map[string]uint64{"report.txt": 1},
		},
		{
			name: "an administrator's own row wins a directory listing collision",
			seeds: []seed{
				{userID: 1, path: "docs", isDir: true},
				{userID: 2, path: "docs/a.txt", content: "theirs", updated: time.Hour},
				{userID: 1, path: "docs/a.txt", content: "mine"},
			},
			path:              "docs",
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"docs", "docs/a.txt"},
			wantOwners:        map[string]uint64{"docs/a.txt": 1},
		},
		{
			name: "an administrator's own row wins an ancestor collision",
			seeds: []seed{
				{userID: 2, path: "docs", isDir: true, updated: time.Hour},
				{userID: 1, path: "docs", isDir: true},
				{userID: 1, path: "docs/sub", isDir: true},
			},
			path:              "docs/sub",
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"docs", "docs/sub"},
			wantOwners:        map[string]uint64{"docs": 1},
		},
		{
			name: "of two foreign rows on a path the newer one is listed",
			seeds: []seed{
				{userID: 2, path: "report.txt", content: "older", updated: -time.Hour},
				{userID: 3, path: "report.txt", content: "newer"},
			},
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"report.txt"},
			wantOwners:        map[string]uint64{"report.txt": 3},
		},
		{
			name: "of two foreign rows with one timestamp the lower id is listed",
			seeds: []seed{
				{userID: 2, path: "report.txt", content: "older"},
				{userID: 3, path: "report.txt", content: "newer"},
			},
			sameStamp:         true,
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"report.txt"},
			wantOwners:        map[string]uint64{"report.txt": 2},
		},
		{
			name: "a directory listing shows the newer of two foreign directory rows",
			seeds: []seed{
				{userID: 2, path: "docs", isDir: true, updated: -time.Hour},
				{userID: 3, path: "docs", isDir: true},
				{userID: 1, path: "docs/sub", isDir: true},
			},
			path:              "docs",
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"docs", "docs/sub"},
			wantOwners:        map[string]uint64{"docs": 3},
		},
		{
			name: "the ancestor top-up shows the same foreign directory row as the directory listing",
			seeds: []seed{
				{userID: 2, path: "docs", isDir: true, updated: -time.Hour},
				{userID: 3, path: "docs", isDir: true},
				{userID: 1, path: "docs/sub", isDir: true},
			},
			path:              "docs/sub",
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"docs", "docs/sub"},
			wantOwners:        map[string]uint64{"docs": 3},
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
				seeded = seedResource(t, db, seeded)
				if rec.updated != 0 {
					require.NoError(t, db.Exec(
						"UPDATE user_resources SET updated_at = ? WHERE id = ?", time.Now().Add(rec.updated), seeded.ID,
					).Error)
				}
			}
			if tt.sameStamp {
				require.NoError(t, db.Exec("UPDATE user_resources SET updated_at = ?", time.Now()).Error)
			}

			var target string
			if tt.rawQuery != "" {
				target = "/resources/?" + tt.rawQuery
			} else {
				target = "/resources/"
				query := []string{}
				if tt.path != "" {
					query = append(query, "path="+tt.path)
				}
				for _, p := range tt.paths {
					query = append(query, "paths[]="+p)
				}
				if tt.recursive {
					query = append(query, "recursive=true")
				}
				if len(query) > 0 {
					target += "?" + strings.Join(query, "&")
				}
			}

			uid := tt.uid
			if uid == 0 {
				uid = 1
			}
			c, w := newResourceTestContextWithUID(http.MethodGet, target, nil, tt.privs, uid)

			svc.ListResources(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus != http.StatusOK {
				return
			}
			list := decodeResourceListResponse(t, w)
			if tt.wantSorted {
				assert.Equal(t, tt.wantResponsePaths, resourcePaths(list.Items))
			} else {
				assert.ElementsMatch(t, tt.wantResponsePaths, resourcePaths(list.Items))
			}
			assert.Equal(t, uint64(len(tt.wantResponsePaths)), list.Total)
			for vPath, owner := range tt.wantOwners {
				for _, item := range list.Items {
					if item.Path == vPath {
						assert.Equal(t, owner, item.UserID, "owner of the row listed at %q", vPath)
					}
				}
			}
		})
	}
}

func TestResources_MkdirResource_CreatesTheDirectoryAndItsMissingParents(t *testing.T) {
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
			name:             "deeply nested mkdir materialises every parent",
			path:             "a/b/c",
			privs:            []string{"resources.edit"},
			wantStatus:       http.StatusOK,
			wantPaths:        []string{"a", "a/b", "a/b/c"},
			wantResponsePath: "a/b/c",
			wantResponseDir:  true,
			wantEvents: []resourceEvent{
				{action: "added", path: "a"},
				{action: "added", path: "a/b"},
				{action: "added", path: "a/b/c"},
			},
		},
		{
			name:             "nested mkdir reuses the parents that already exist",
			seeds:            []seed{{path: "a", isDir: true}},
			path:             "a/b",
			privs:            []string{"resources.edit"},
			wantStatus:       http.StatusOK,
			wantPaths:        []string{"a", "a/b"},
			wantResponsePath: "a/b",
			wantResponseDir:  true,
			wantEvents:       []resourceEvent{{action: "added", path: "a/b"}},
		},
		{
			name:             "mkdir on an existing leaf fills in the parents it never had",
			seeds:            []seed{{path: "a/b/c", isDir: true}},
			path:             "a/b/c",
			privs:            []string{"resources.edit"},
			wantStatus:       http.StatusOK,
			wantPaths:        []string{"a", "a/b", "a/b/c"},
			wantResponsePath: "a/b/c",
			wantResponseDir:  true,
			wantEvents: []resourceEvent{
				{action: "added", path: "a"},
				{action: "added", path: "a/b"},
			},
		},
		{
			name:       "a file standing in for a parent aborts the whole chain",
			seeds:      []seed{{path: "a", content: "data"}},
			path:       "a/b/c",
			privs:      []string{"resources.edit"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a"},
			wantEvents: nil,
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
					Status string               `json:"status"`
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

func TestResources_UploadResources_StoresEachFileAsABlobRow(t *testing.T) {
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
		rawFileName       string // sent unescaped, with "payload" as content
		nonMultipart      bool
		privs             []string
		seeds             []seed
		wantStatus        int
		wantPaths         []string
		wantResponsePaths []string
		wantEvents        []resourceEvent
		wantMissingBlobs  []string
		wantPresentBlobs  []string
		wantHashes        map[string]string // path -> content whose md5 the row references
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
			files:            []uploadTestFile{{name: "bad<>name.txt", content: "payload"}},
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
		{
			name:         "non-multipart body returns bad request",
			nonMultipart: true,
			privs:        []string{"resources.upload"},
			wantStatus:   http.StatusBadRequest,
			wantPaths:    []string{},
		},
		{
			name:        "a raw line feed in the filename is rejected",
			rawFileName: "evil\nname.txt",
			privs:       []string{"resources.upload"},
			wantStatus:  http.StatusBadRequest,
			wantPaths:   []string{},
		},
		{
			name:        "a raw carriage return in the filename is rejected",
			rawFileName: "evil\rname.txt",
			privs:       []string{"resources.upload"},
			wantStatus:  http.StatusBadRequest,
			wantPaths:   []string{},
		},
		{
			name:              "two files with one content share one blob",
			files:             []uploadTestFile{{name: "a.txt", content: "shared"}, {name: "b.txt", content: "shared"}},
			privs:             []string{"resources.upload"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"a.txt", "b.txt"},
			wantResponsePaths: []string{"a.txt", "b.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "a.txt"}, {action: "added", path: "b.txt"}},
			wantHashes:        map[string]string{"a.txt": "shared", "b.txt": "shared"},
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

			var body *bytes.Buffer
			var contentType string
			switch {
			case tt.nonMultipart:
				body, contentType = bytes.NewBufferString(`{"hello":"world"}`), "application/json"
			case tt.rawFileName != "":
				body, contentType = rawMultipartUpload("files", tt.rawFileName, "payload")
			default:
				fieldName := tt.fieldName
				if fieldName == "" {
					fieldName = "files"
				}
				body, contentType = multipartUploadBodyWithField(t, tt.files, fieldName)
			}
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
			resourcesRequireHashes(t, db, tt.wantHashes)
			resourcesRequireBlobsMatchRows(t, db, dataDir)
		})
	}
}

func TestResources_DownloadResource_ServesAFileOrAZip(t *testing.T) {
	type seed struct {
		userID        uint64
		path          string
		isDir         bool
		content       string
		skipBlobWrite bool          // the row stays, its blob file is never written
		updated       time.Duration // updated_at relative to now; 0 keeps the insert time
	}

	tests := []struct {
		name             string
		seeds            []seed
		sameStamp        bool     // every row carries one updated_at
		path             string   // builds ?path=<value>
		paths            []string // builds ?paths[]=<value> for each
		rawQuery         string   // when set, used verbatim (overrides path/paths)
		privs            []string
		uid              uint64
		wantStatus       int
		wantBody         string
		wantContentType  string
		wantDispContains string
		wantZipEntries   map[string]string
		wantErrCode      string
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
			// a lone directory is zipped relative to itself
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
		{
			name:             "single file via paths[] downloaded directly",
			seeds:            []seed{{path: "report.txt", content: "payload"}},
			paths:            []string{"report.txt"},
			privs:            []string{"resources.download"},
			wantStatus:       http.StatusOK,
			wantBody:         "payload",
			wantDispContains: "report.txt",
		},
		{
			name: "single directory via paths[] uses dir-relative zip paths",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a-data"},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/b.txt", content: "b-data"},
			},
			paths:            []string{"docs"},
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
			// Seed order fixes the autoincrement ids: the other owner's row is 1.
			name: "administrator reaches the other owner's row by id",
			seeds: []seed{
				{userID: 2, path: "report.txt", content: "theirs"},
				{userID: 1, path: "report.txt", content: "mine"},
			},
			rawQuery:         "id=1",
			privs:            []string{"resources.admin"},
			wantStatus:       http.StatusOK,
			wantBody:         "theirs",
			wantDispContains: "report.txt",
		},
		{
			name: "path still answers with the administrator's own row",
			seeds: []seed{
				{userID: 2, path: "report.txt", content: "theirs"},
				{userID: 1, path: "report.txt", content: "mine"},
			},
			path:             "report.txt",
			privs:            []string{"resources.admin"},
			wantStatus:       http.StatusOK,
			wantBody:         "mine",
			wantDispContains: "report.txt",
		},
		{
			name: "of two foreign rows on a path the newer one is downloaded, as the listing shows",
			seeds: []seed{
				{userID: 2, path: "report.txt", content: "older", updated: -time.Hour},
				{userID: 3, path: "report.txt", content: "newer"},
			},
			path:             "report.txt",
			privs:            []string{"resources.admin"},
			wantStatus:       http.StatusOK,
			wantBody:         "newer",
			wantDispContains: "report.txt",
		},
		{
			name: "of two foreign rows with one timestamp the lower id is downloaded, as the listing shows",
			seeds: []seed{
				{userID: 2, path: "report.txt", content: "older"},
				{userID: 3, path: "report.txt", content: "newer"},
			},
			sameStamp:        true,
			path:             "report.txt",
			privs:            []string{"resources.admin"},
			wantStatus:       http.StatusOK,
			wantBody:         "older",
			wantDispContains: "report.txt",
		},
		{
			name: "plain user cannot reach another owner's row by id, which is as absent as by path",
			seeds: []seed{
				{userID: 2, path: "report.txt", content: "theirs"},
				{userID: 1, path: "own.txt", content: "mine"},
			},
			rawQuery:   "id=1",
			privs:      []string{"resources.download"},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "an id whose path the archive already holds is refused, not dropped",
			seeds: []seed{
				{userID: 2, path: "report.txt", content: "theirs"},
				{userID: 1, path: "report.txt", content: "mine"},
			},
			rawQuery:    "id=1&paths[]=report.txt",
			privs:       []string{"resources.admin"},
			wantStatus:  http.StatusBadRequest,
			wantErrCode: "Resources.SharedPath",
		},
		{
			name: "two ids that share a path are refused, not collapsed",
			seeds: []seed{
				{userID: 2, path: "report.txt", content: "theirs"},
				{userID: 1, path: "report.txt", content: "mine"},
			},
			rawQuery:    "ids[]=1&ids[]=2",
			privs:       []string{"resources.admin"},
			wantStatus:  http.StatusBadRequest,
			wantErrCode: "Resources.SharedPath",
		},
		{
			name: "a directory requested by id is refused when one of its files shares the caller's path",
			seeds: []seed{
				{userID: 2, path: "docs", isDir: true},
				{userID: 2, path: "docs/a.txt", content: "theirs"},
				{userID: 1, path: "docs/a.txt", content: "mine"},
			},
			rawQuery:    "id=1&paths[]=docs/a.txt",
			privs:       []string{"resources.admin"},
			wantStatus:  http.StatusBadRequest,
			wantErrCode: "Resources.SharedPath",
		},
		{
			name: "an id and a path that name the same row give it once",
			seeds: []seed{
				{userID: 2, path: "report.txt", content: "theirs"},
				{userID: 1, path: "report.txt", content: "mine"},
			},
			rawQuery:         "id=2&paths[]=report.txt",
			privs:            []string{"resources.admin"},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "download.zip",
			wantZipEntries: map[string]string{
				"report.txt": "mine",
			},
		},
		{
			name:       "non-numeric id returns bad request",
			seeds:      []seed{{path: "report.txt", content: "payload"}},
			rawQuery:   "id=abc",
			privs:      []string{"resources.download"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown id returns not found",
			seeds:      []seed{{path: "report.txt", content: "payload"}},
			rawQuery:   "id=999",
			privs:      []string{"resources.download"},
			wantStatus: http.StatusNotFound,
		},
		{
			// without the owner predicate the order of paths[] would decide whose file is zipped
			name: "archive keeps the administrator's own file when the directory comes first",
			seeds: []seed{
				{userID: 2, path: "docs", isDir: true},
				{userID: 2, path: "docs/a.txt", content: "theirs"},
				{userID: 1, path: "docs/a.txt", content: "mine"},
			},
			paths:            []string{"docs", "docs/a.txt"},
			privs:            []string{"resources.admin"},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "download.zip",
			wantZipEntries: map[string]string{
				"docs/a.txt": "mine",
			},
		},
		{
			name: "archive keeps the administrator's own file when the file comes first",
			seeds: []seed{
				{userID: 2, path: "docs", isDir: true},
				{userID: 2, path: "docs/a.txt", content: "theirs"},
				{userID: 1, path: "docs/a.txt", content: "mine"},
			},
			paths:            []string{"docs/a.txt", "docs"},
			privs:            []string{"resources.admin"},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "download.zip",
			wantZipEntries: map[string]string{
				"docs/a.txt": "mine",
			},
		},
		{
			name: "two files via paths[] packaged into zip with full virtual paths",
			seeds: []seed{
				{path: "docs/a.txt", content: "a-data"},
				{path: "other/b.txt", content: "b-data"},
			},
			paths:            []string{"docs/a.txt", "other/b.txt"},
			privs:            []string{"resources.download"},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "download.zip",
			wantZipEntries: map[string]string{
				"docs/a.txt":  "a-data",
				"other/b.txt": "b-data",
			},
		},
		{
			name: "path= and paths[] combined produce multi-entry zip",
			seeds: []seed{
				{path: "a.txt", content: "a-data"},
				{path: "b.txt", content: "b-data"},
			},
			path:             "a.txt",
			paths:            []string{"b.txt"},
			privs:            []string{"resources.download"},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "download.zip",
			wantZipEntries: map[string]string{
				"a.txt": "a-data",
				"b.txt": "b-data",
			},
		},
		{
			name: "file and directory via paths[] combined in zip with full virtual paths",
			seeds: []seed{
				{path: "report.txt", content: "report-data"},
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a-data"},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/b.txt", content: "b-data"},
			},
			paths:            []string{"report.txt", "docs"},
			privs:            []string{"resources.download"},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "download.zip",
			wantZipEntries: map[string]string{
				"report.txt":     "report-data",
				"docs/a.txt":     "a-data",
				"docs/sub/b.txt": "b-data",
			},
		},
		{
			name:             "duplicate paths in paths[] deduplicated",
			seeds:            []seed{{path: "a.txt", content: "alpha"}},
			paths:            []string{"a.txt", "a.txt"},
			privs:            []string{"resources.download"},
			wantStatus:       http.StatusOK,
			wantBody:         "alpha",
			wantDispContains: "a.txt",
		},
		{
			name:       "whitespace-only paths[] returns bad request",
			rawQuery:   "paths[]=%20%20&paths[]=%09",
			privs:      []string{"resources.download"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid path in paths[] returns bad request",
			rawQuery:   "paths[]=docs&paths[]=../escape",
			privs:      []string{"resources.download"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing resource in batch returns not found",
			seeds: []seed{
				{path: "a.txt", content: "alpha"},
			},
			paths:      []string{"a.txt", "missing.txt"},
			privs:      []string{"resources.download"},
			wantStatus: http.StatusNotFound,
		},
		{
			// a missing blob fails the archive before a byte is streamed, never a truncated 200
			name: "blob missing on disk in a zip batch fails cleanly, no truncated 200",
			seeds: []seed{
				{path: "a.txt", content: "alpha"},
				{path: "b.txt", content: "beta", skipBlobWrite: true},
			},
			paths:      []string{"a.txt", "b.txt"},
			privs:      []string{"resources.download"},
			wantStatus: http.StatusInternalServerError,
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
				seeded = seedResource(t, db, seeded)
				if rec.updated != 0 {
					require.NoError(t, db.Exec(
						"UPDATE user_resources SET updated_at = ? WHERE id = ?", time.Now().Add(rec.updated), seeded.ID,
					).Error)
				}
			}
			if tt.sameStamp {
				require.NoError(t, db.Exec("UPDATE user_resources SET updated_at = ?", time.Now()).Error)
			}

			var target string
			if tt.rawQuery != "" {
				target = "/resources/download?" + tt.rawQuery
			} else {
				query := []string{}
				if tt.path != "" {
					query = append(query, "path="+tt.path)
				}
				for _, p := range tt.paths {
					query = append(query, "paths[]="+p)
				}
				target = "/resources/download"
				if len(query) > 0 {
					target += "?" + strings.Join(query, "&")
				}
			}

			uid := tt.uid
			if uid == 0 {
				uid = 1
			}
			c, w := newResourceTestContextWithUID(http.MethodGet, target, nil, tt.privs, uid)

			svc.DownloadResource(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus != http.StatusOK {
				if tt.wantErrCode != "" {
					assert.Contains(t, w.Body.String(), `"code":"`+tt.wantErrCode+`"`)
				}
				return
			}
			if tt.wantContentType != "" {
				assert.Equal(t, tt.wantContentType, w.Header().Get("Content-Type"))
			}
			if tt.wantDispContains != "" {
				assert.Contains(t, w.Header().Get("Content-Disposition"), tt.wantDispContains)
			}
			if tt.wantZipEntries != nil {
				// a streamed archive is never buffered, so it has no Content-Length
				assert.Empty(t, w.Header().Get("Content-Length"))
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

func TestResources_StreamZipArchive_StreamsUnbufferedAndReportsABuildError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("streams archive without content-length", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/download", nil)

		err := streamZipArchive(c, "out.zip", func(zw io.Writer) error {
			z := zip.NewWriter(zw)
			f, createErr := z.Create("hello.txt")
			require.NoError(t, createErr)
			if _, writeErr := f.Write([]byte("hello world")); writeErr != nil {
				return writeErr
			}
			return z.Close()
		})

		require.NoError(t, err)
		assert.False(t, c.IsAborted())
		assert.Equal(t, "application/zip", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "out.zip")
		assert.Empty(t, w.Header().Get("Content-Length"))

		zr, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
		require.NoError(t, err)
		require.Len(t, zr.File, 1)
		assert.Equal(t, "hello.txt", zr.File[0].Name)
		rc, err := zr.File[0].Open()
		require.NoError(t, err)
		data, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.NoError(t, rc.Close())
		assert.Equal(t, "hello world", string(data))
	})

	t.Run("propagates build error after partial stream and aborts", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/download", nil)

		sentinel := errors.New("source reader failed mid-stream")
		err := streamZipArchive(c, "out.zip", func(zw io.Writer) error {
			z := zip.NewWriter(zw)
			f, createErr := z.Create("partial.txt")
			require.NoError(t, createErr)
			if _, writeErr := f.Write([]byte("partial data")); writeErr != nil {
				return writeErr
			}
			// Flush bytes to the response, then fail as a slow/erroring source would.
			if flushErr := z.Flush(); flushErr != nil {
				return flushErr
			}
			return sentinel
		})

		require.ErrorIs(t, err, sentinel)
		assert.True(t, c.IsAborted())
	})

	t.Run("returns structured error when build fails before writing", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/download", nil)

		sentinel := errors.New("blob open failed before any write")
		err := streamZipArchive(c, "out.zip", func(zw io.Writer) error {
			return sentinel
		})

		require.ErrorIs(t, err, sentinel)
		// nothing was streamed, so a structured error replaces a truncated 200
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotEqual(t, "application/zip", w.Header().Get("Content-Type"))
		assert.True(t, c.IsAborted())
	})
}

func TestResources_DeleteResource_RemovesTheTreeAndItsOrphanBlobs(t *testing.T) {
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
		targetPath        string   // builds ?path=<value>
		targetPaths       []string // builds ?paths[]=<value> for each
		rawQuery          string   // overrides targetPath/targetPaths when set
		privs             []string
		wantStatus        int
		wantPaths         []string
		wantResponsePaths []string // in response order
		wantEvents        []resourceEvent
		wantMissingBlobs  []string
		wantPresentBlobs  []string
		wantOwners        map[string]uint64
		wantPublisherUIDs []int64 // whose channel the events went to (nil = don't check)
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
		{
			name: "delete two files via paths[] in batch",
			seeds: []seed{
				{path: "a.txt", content: "a"},
				{path: "b.txt", content: "b"},
				{path: "keep.txt", content: "keep"},
			},
			targetPaths:       []string{"a.txt", "b.txt"},
			privs:             []string{"resources.delete"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"keep.txt"},
			wantResponsePaths: []string{"a.txt", "b.txt"},
			wantEvents: []resourceEvent{
				{action: "deleted", path: "a.txt"},
				{action: "deleted", path: "b.txt"},
			},
			wantMissingBlobs: []string{"a", "b"},
			wantPresentBlobs: []string{"keep"},
		},
		{
			name: "path= and paths[] combined delete both targets",
			seeds: []seed{
				{path: "a.txt", content: "a"},
				{path: "b.txt", content: "b"},
			},
			targetPath:        "a.txt",
			targetPaths:       []string{"b.txt"},
			privs:             []string{"resources.delete"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{},
			wantResponsePaths: []string{"a.txt", "b.txt"},
			wantEvents: []resourceEvent{
				{action: "deleted", path: "a.txt"},
				{action: "deleted", path: "b.txt"},
			},
		},
		{
			name: "duplicate paths in paths[] deduplicated",
			seeds: []seed{
				{path: "a.txt", content: "a"},
			},
			targetPaths:       []string{"a.txt", "a.txt"},
			privs:             []string{"resources.delete"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{},
			wantResponsePaths: []string{"a.txt"},
			wantEvents:        []resourceEvent{{action: "deleted", path: "a.txt"}},
			wantMissingBlobs:  []string{"a"},
		},
		{
			name:       "whitespace-only paths[] returns bad request",
			rawQuery:   "paths[]=%20%20&paths[]=%09",
			privs:      []string{"resources.delete"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing second path in batch returns 404 without deleting first",
			seeds: []seed{
				{path: "a.txt", content: "a"},
			},
			targetPaths: []string{"a.txt", "missing.txt"},
			privs:       []string{"resources.delete"},
			wantStatus:  http.StatusNotFound,
			wantPaths:   []string{"a.txt"}, // not deleted due to fail-fast
		},
		{
			name: "overlapping paths parent and child both processed correctly",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/b.txt", content: "b"},
			},
			targetPath:        "docs",
			targetPaths:       []string{"docs/a.txt"},
			privs:             []string{"resources.delete"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{},
			wantResponsePaths: []string{"docs", "docs/a.txt", "docs/sub", "docs/sub/b.txt"},
			wantEvents: []resourceEvent{
				{action: "deleted", path: "docs"},
				{action: "deleted", path: "docs/a.txt"},
				{action: "deleted", path: "docs/sub"},
				{action: "deleted", path: "docs/sub/b.txt"},
			},
		},
		{
			name: "response sorted by path across multiple deleted targets",
			seeds: []seed{
				{path: "z.txt", content: "z"},
				{path: "a.txt", content: "a"},
				{path: "m.txt", content: "m"},
			},
			targetPaths:       []string{"z.txt", "a.txt", "m.txt"},
			privs:             []string{"resources.delete"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"a.txt", "m.txt", "z.txt"},
			wantEvents: []resourceEvent{
				{action: "deleted", path: "a.txt"},
				{action: "deleted", path: "m.txt"},
				{action: "deleted", path: "z.txt"},
			},
		},
		{
			name:              "an administrator deletes the foreign row the listing shows and tells its owner",
			seeds:             []seed{{userID: 2, path: "report.txt", content: "theirs"}},
			targetPaths:       []string{"report.txt"},
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{},
			wantResponsePaths: []string{"report.txt"},
			wantEvents:        []resourceEvent{{action: "deleted", path: "report.txt"}},
			wantMissingBlobs:  []string{"theirs"},
			wantPublisherUIDs: []int64{2},
		},
		{
			name:             "a caller without the admin privilege cannot reach a foreign row",
			seeds:            []seed{{userID: 2, path: "report.txt", content: "theirs"}},
			targetPaths:      []string{"report.txt"},
			privs:            []string{"resources.delete"},
			wantStatus:       http.StatusNotFound,
			wantPaths:        []string{"report.txt"},
			wantPresentBlobs: []string{"theirs"},
		},
		{
			name: "an administrator's own row wins over a foreign row seeded before it",
			seeds: []seed{
				{userID: 2, path: "report.txt", content: "theirs"},
				{userID: 1, path: "report.txt", content: "mine"},
			},
			targetPaths:       []string{"report.txt"},
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"report.txt"},
			wantResponsePaths: []string{"report.txt"},
			wantEvents:        []resourceEvent{{action: "deleted", path: "report.txt"}},
			wantMissingBlobs:  []string{"mine"},
			wantPresentBlobs:  []string{"theirs"},
			wantOwners:        map[string]uint64{"report.txt": 2},
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

			var target string
			if tt.rawQuery != "" {
				target = "/resources/?" + tt.rawQuery
			} else {
				query := []string{}
				if tt.targetPath != "" {
					query = append(query, "path="+tt.targetPath)
				}
				for _, p := range tt.targetPaths {
					query = append(query, "paths[]="+p)
				}
				target = "/resources/?" + strings.Join(query, "&")
			}
			c, w := newResourceTestContext(http.MethodDelete, target, nil, tt.privs)
			svc.DeleteResource(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantPaths != nil {
				assert.ElementsMatch(t, tt.wantPaths, allResourcePaths(t, db))
			}
			assert.Equal(t, tt.wantEvents, ss.events)
			if tt.wantPublisherUIDs != nil {
				assert.Equal(t, tt.wantPublisherUIDs, ss.publisherUIDs)
			}
			if tt.wantStatus == http.StatusOK {
				list := decodeResourceListResponse(t, w)
				assert.Equal(t, tt.wantResponsePaths, resourcePaths(list.Items))
			}
			resourcesRequireOwners(t, db, tt.wantOwners)
			resourcesRequireBlobsMatchRows(t, db, dataDir)
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

func TestResources_CopyResource_CopiesFilesAndTrees(t *testing.T) {
	type seed = resourcesTransferSeed
	type copyRequest = resourcesTransferRequest

	tests := []resourcesTransferCase{
		{
			name:              "file to absent path adds a copy of the same blob",
			seeds:             []seed{{path: "a.txt", content: "src"}},
			req:               copyRequest{Source: "a.txt", Destination: "copies/a.txt"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"a.txt", "copies", "copies/a.txt"},
			wantResponsePaths: []string{"copies", "copies/a.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "copies"}, {action: "added", path: "copies/a.txt"}},
			wantHashes:        map[string]string{"a.txt": "src", "copies/a.txt": "src"},
		},
		{
			name:       "malformed json returns bad request",
			seeds:      []seed{{path: "a.txt", content: "src"}},
			rawBody:    `{not valid json`,
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"a.txt"},
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
			name:              "force replaces a file at a nested named destination",
			seeds:             []seed{{path: "a.txt", content: "src"}, {path: "docs", isDir: true}, {path: "docs/sub", content: "blocking"}},
			req:               copyRequest{Source: "a.txt", Destination: "docs/sub/", Force: true},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"a.txt", "docs", "docs/sub", "docs/sub/a.txt"},
			wantResponsePaths: []string{"docs/sub", "docs/sub/a.txt"},
			wantEvents: []resourceEvent{
				{action: "deleted", path: "docs/sub"},
				{action: "added", path: "docs/sub"},
				{action: "added", path: "docs/sub/a.txt"},
			},
			wantDeletedBlobTexts: []string{"blocking"},
		},
		{
			name:       "force stops at a file one level above a nested destination",
			seeds:      []seed{{path: "a.txt", content: "src"}, {path: "docs", content: "blocking"}},
			req:        copyRequest{Source: "a.txt", Destination: "docs/sub/", Force: true},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "docs"},
		},
		{
			name:       "force does not replace a file on the way to the destination",
			seeds:      []seed{{path: "a.txt", content: "src"}, {path: "victim.txt", content: "irreplaceable"}},
			req:        copyRequest{Source: "a.txt", Destination: "victim.txt/nested/copy.txt", Force: true},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "victim.txt"},
		},
		{
			name:       "force does not replace a file at the destination's parent",
			seeds:      []seed{{path: "a.txt", content: "src"}, {path: "docs", content: "blocking"}},
			req:        copyRequest{Source: "a.txt", Destination: "docs/copy.txt", Force: true},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "docs"},
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
		{
			name: "multi-source: two files copied to a common base directory",
			seeds: []seed{
				{path: "a.txt", content: "aaa"},
				{path: "b.txt", content: "bbb"},
			},
			req:               copyRequest{Sources: []string{"a.txt", "b.txt"}, Destination: "backup"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"a.txt", "b.txt", "backup", "backup/a.txt", "backup/b.txt"},
			wantResponsePaths: []string{"backup", "backup/a.txt", "backup/b.txt"},
			wantEvents: []resourceEvent{
				{action: "added", path: "backup"},
				{action: "added", path: "backup/a.txt"},
				{action: "added", path: "backup/b.txt"},
			},
		},
		{
			name: "sources differing only by separator are one source",
			seeds: []seed{
				{path: "dir", isDir: true},
				{path: "dir/z.txt", content: "zzz"},
			},
			req:               copyRequest{Sources: []string{"dir/z.txt", "dir\\z.txt"}, Destination: "backup"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"dir", "dir/z.txt", "backup"},
			wantResponsePaths: []string{"backup"},
			wantEvents: []resourceEvent{
				{action: "added", path: "backup"},
			},
		},
		{
			name: "multi-source: source and sources merged and deduplicated",
			seeds: []seed{
				{path: "a.txt", content: "aaa"},
				{path: "b.txt", content: "bbb"},
			},
			req:               copyRequest{Source: "a.txt", Sources: []string{"a.txt", "b.txt"}, Destination: "backup"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"a.txt", "b.txt", "backup", "backup/a.txt", "backup/b.txt"},
			wantResponsePaths: []string{"backup", "backup/a.txt", "backup/b.txt"},
			wantEvents: []resourceEvent{
				{action: "added", path: "backup"},
				{action: "added", path: "backup/a.txt"},
				{action: "added", path: "backup/b.txt"},
			},
		},
		{
			name: "multi-source: file and directory copied to existing base directory",
			seeds: []seed{
				{path: "report.txt", content: "r"},
				{path: "docs", isDir: true},
				{path: "docs/readme.md", content: "readme"},
				{path: "dest", isDir: true},
			},
			req:        copyRequest{Sources: []string{"report.txt", "docs"}, Destination: "dest"},
			wantStatus: http.StatusOK,
			wantPaths:  []string{"report.txt", "docs", "docs/readme.md", "dest", "dest/report.txt", "dest/docs", "dest/docs/readme.md"},
			wantResponsePaths: []string{
				"dest/report.txt",
				"dest/docs",
				"dest/docs/readme.md",
			},
			wantEvents: []resourceEvent{
				{action: "added", path: "dest/report.txt"},
				{action: "added", path: "dest/docs"},
				{action: "added", path: "dest/docs/readme.md"},
			},
		},
		{
			name: "multi-source: force overwrites existing file at target",
			seeds: []seed{
				{path: "a.txt", content: "new-a"},
				{path: "b.txt", content: "new-b"},
				{path: "backup", isDir: true},
				{path: "backup/a.txt", content: "old-a"},
			},
			req:        copyRequest{Sources: []string{"a.txt", "b.txt"}, Destination: "backup", Force: true},
			wantStatus: http.StatusOK,
			wantPaths:  []string{"a.txt", "b.txt", "backup", "backup/a.txt", "backup/b.txt"},
			wantResponsePaths: []string{
				"backup/a.txt",
				"backup/b.txt",
			},
			// publish order: deleted, added, updated
			wantEvents: []resourceEvent{
				{action: "deleted", path: "backup/a.txt"},
				{action: "added", path: "backup/b.txt"},
				{action: "updated", path: "backup/a.txt"},
			},
			wantDeletedBlobTexts: []string{"old-a"},
		},
		{
			name: "multi-source: without force returns conflict when target exists",
			seeds: []seed{
				{path: "a.txt", content: "a"},
				{path: "b.txt", content: "b"},
				{path: "backup", isDir: true},
				{path: "backup/a.txt", content: "old"},
			},
			req:        copyRequest{Sources: []string{"a.txt", "b.txt"}, Destination: "backup"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "b.txt", "backup", "backup/a.txt"},
		},
		{
			name: "multi-source: duplicate basenames returns conflict",
			seeds: []seed{
				{path: "dir1/x.txt", content: "x1"},
				{path: "dir2/x.txt", content: "x2"},
			},
			req:        copyRequest{Sources: []string{"dir1/x.txt", "dir2/x.txt"}, Destination: "backup"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"dir1/x.txt", "dir2/x.txt"},
		},
		{
			name: "multi-source: one missing source returns not found",
			seeds: []seed{
				{path: "a.txt", content: "a"},
			},
			req:        copyRequest{Sources: []string{"a.txt", "ghost.txt"}, Destination: "backup"},
			wantStatus: http.StatusNotFound,
			wantPaths:  []string{"a.txt"},
		},
		{
			name:       "multi-source: empty sources returns bad request",
			req:        copyRequest{Sources: []string{}, Destination: "backup"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "multi-source: blank entries only returns bad request",
			req:        copyRequest{Sources: []string{"   ", ""}, Destination: "backup"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "multi-source: destination is a file returns conflict",
			seeds: []seed{
				{path: "a.txt", content: "a"},
				{path: "b.txt", content: "b"},
				{path: "dest.file", content: "file"},
			},
			req:        copyRequest{Sources: []string{"a.txt", "b.txt"}, Destination: "dest.file"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "b.txt", "dest.file"},
		},
		{
			name: "multi-source: directory into itself returns bad request",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "a.txt", content: "a"},
			},
			req:        copyRequest{Sources: []string{"docs", "a.txt"}, Destination: "docs/sub"},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"a.txt", "docs"},
		},
		{
			name: "multi-source: two directories copied to common base",
			seeds: []seed{
				{path: "src1", isDir: true},
				{path: "src1/file1.txt", content: "f1"},
				{path: "src2", isDir: true},
				{path: "src2/file2.txt", content: "f2"},
			},
			req:        copyRequest{Sources: []string{"src1", "src2"}, Destination: "all"},
			wantStatus: http.StatusOK,
			wantPaths: []string{
				"src1", "src1/file1.txt",
				"src2", "src2/file2.txt",
				"all", "all/src1", "all/src1/file1.txt",
				"all/src2", "all/src2/file2.txt",
			},
			wantResponsePaths: []string{
				"all", "all/src1", "all/src1/file1.txt",
				"all/src2", "all/src2/file2.txt",
			},
			wantEvents: []resourceEvent{
				{action: "added", path: "all"},
				{action: "added", path: "all/src1"},
				{action: "added", path: "all/src1/file1.txt"},
				{action: "added", path: "all/src2"},
				{action: "added", path: "all/src2/file2.txt"},
			},
		},
		{
			name: "multi-source: invalid source path returns bad request",
			seeds: []seed{
				{path: "a.txt", content: "a"},
			},
			req:        copyRequest{Sources: []string{"../escape.txt", "a.txt"}, Destination: "backup"},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"a.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourcesRunTransferCase(t, tt, http.MethodPost, "/resources/copy", (*ResourceService).CopyResource)
		})
	}
}

func TestResources_MoveResource_MovesFilesAndTrees(t *testing.T) {
	type seed = resourcesTransferSeed
	type moveRequest = resourcesTransferRequest

	tests := []resourcesTransferCase{
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
			name:       "malformed json returns bad request",
			seeds:      []seed{{path: "a.txt", content: "src"}},
			rawBody:    `{not valid json`,
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"a.txt"},
		},
		{
			name:              "an administrator's move leaves a foreign file with its owner",
			seeds:             []seed{{userID: 2, path: "report.txt", content: "theirs"}},
			req:               moveRequest{Source: "report.txt", Destination: "archive/report.txt"},
			privs:             []string{"resources.admin"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"archive", "archive/report.txt"},
			wantResponsePaths: []string{"archive", "archive/report.txt"},
			wantEvents:        []resourceEvent{{action: "added", path: "archive"}, {action: "updated", path: "archive/report.txt"}},
			wantOwners:        map[string]uint64{"archive": 2, "archive/report.txt": 2},
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
			wantResponsePaths: []string{"docs", "docs/a.txt"},
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
			wantResponsePaths:    []string{"docs", "docs/a.txt"},
			wantEvents:           []resourceEvent{{action: "deleted", path: "docs"}, {action: "added", path: "docs"}, {action: "updated", path: "docs/a.txt"}},
			wantDeletedBlobTexts: []string{"blocking"},
		},
		{
			name:       "force does not replace a file on the way to the destination",
			seeds:      []seed{{path: "a.txt", content: "src"}, {path: "victim.txt", content: "irreplaceable"}},
			req:        moveRequest{Source: "a.txt", Destination: "victim.txt/nested/moved.txt", Force: true},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "victim.txt"},
		},
		{
			name:       "force does not replace a file at the destination's parent",
			seeds:      []seed{{path: "a.txt", content: "src"}, {path: "docs", content: "blocking"}},
			req:        moveRequest{Source: "a.txt", Destination: "docs/moved.txt", Force: true},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "docs"},
		},
		{
			name: "force does not replace a file at a moved directory's parent",
			seeds: []seed{
				{path: "src", isDir: true},
				{path: "src/a.txt", content: "src"},
				{path: "docs", content: "blocking"},
			},
			req:        moveRequest{Source: "src", Destination: "docs/moved", Force: true},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"docs", "src", "src/a.txt"},
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
			wantResponsePaths: []string{"one", "one/two", "one/two/three", "one/two/three/a.txt"},
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
			wantResponsePaths: []string{"archive", "archive/docs", "archive/docs/a.txt", "archive/docs/sub", "archive/docs/sub/b.txt"},
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
			name:              "directory to existing directory moves inside",
			seeds:             []seed{{path: "docs", isDir: true}, {path: "archive", isDir: true}},
			req:               moveRequest{Source: "docs", Destination: "archive"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"archive", "archive/docs"},
			wantResponsePaths: []string{"archive/docs"},
			wantEvents:        []resourceEvent{{action: "updated", path: "archive/docs"}},
		},
		{
			name: "directory to existing directory moves whole tree inside",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/b.txt", content: "b"},
				{path: "archive", isDir: true},
			},
			req:               moveRequest{Source: "docs", Destination: "archive"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"archive", "archive/docs", "archive/docs/a.txt", "archive/docs/sub", "archive/docs/sub/b.txt"},
			wantResponsePaths: []string{"archive/docs", "archive/docs/a.txt", "archive/docs/sub", "archive/docs/sub/b.txt"},
			wantEvents: []resourceEvent{
				{action: "updated", path: "archive/docs"},
				{action: "updated", path: "archive/docs/a.txt"},
				{action: "updated", path: "archive/docs/sub"},
				{action: "updated", path: "archive/docs/sub/b.txt"},
			},
		},
		{
			name: "directory moved into existing dir that already contains same-name subdir conflicts without force",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "archive", isDir: true},
				{path: "archive/docs", isDir: true},
			},
			req:        moveRequest{Source: "docs", Destination: "archive"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"archive", "archive/docs", "docs"},
		},
		{
			name: "directory merged into existing subdirectory overwrites file with force",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "src-a"},
				{path: "archive", isDir: true},
				{path: "archive/docs", isDir: true},
				{path: "archive/docs/a.txt", content: "dst-a"},
			},
			req:                  moveRequest{Source: "docs", Destination: "archive", Force: true},
			wantStatus:           http.StatusOK,
			wantPaths:            []string{"archive", "archive/docs", "archive/docs/a.txt"},
			wantResponsePaths:    []string{"archive/docs/a.txt"},
			wantEvents:           []resourceEvent{{action: "deleted", path: "archive/docs/a.txt"}, {action: "updated", path: "archive/docs/a.txt"}, {action: "deleted", path: "docs"}},
			wantDeletedBlobTexts: []string{"dst-a"},
		},
		{
			name: "directory merged into existing subdirectory keeps destination subdirectory with force",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/sub", isDir: true},
				{path: "docs/sub/a.txt", content: "a"},
				{path: "archive", isDir: true},
				{path: "archive/docs", isDir: true},
				{path: "archive/docs/sub", isDir: true},
			},
			req:               moveRequest{Source: "docs", Destination: "archive", Force: true},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"archive", "archive/docs", "archive/docs/sub", "archive/docs/sub/a.txt"},
			wantResponsePaths: []string{"archive/docs/sub/a.txt"},
			wantEvents: []resourceEvent{
				{action: "updated", path: "archive/docs/sub/a.txt"},
				{action: "deleted", path: "docs/sub"},
				{action: "deleted", path: "docs"},
			},
		},
		{
			name: "directory merged into existing subdirectory with file-vs-dir conflict fails",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "docs/a.txt", content: "a"},
				{path: "archive", isDir: true},
				{path: "archive/docs", isDir: true},
				{path: "archive/docs/a.txt", isDir: true},
			},
			req:        moveRequest{Source: "docs", Destination: "archive", Force: true},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"archive", "archive/docs", "archive/docs/a.txt", "docs", "docs/a.txt"},
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
		{
			name: "multi-source: two files moved to a common base directory",
			seeds: []seed{
				{path: "a.txt", content: "aaa"},
				{path: "b.txt", content: "bbb"},
			},
			req:               moveRequest{Sources: []string{"a.txt", "b.txt"}, Destination: "archive"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"archive", "archive/a.txt", "archive/b.txt"},
			wantResponsePaths: []string{"archive", "archive/a.txt", "archive/b.txt"},
			wantEvents: []resourceEvent{
				{action: "added", path: "archive"},
				{action: "updated", path: "archive/a.txt"},
				{action: "updated", path: "archive/b.txt"},
			},
		},
		{
			name: "sources differing only by separator are one source",
			seeds: []seed{
				{path: "dir", isDir: true},
				{path: "dir/z.txt", content: "zzz"},
			},
			req:               moveRequest{Sources: []string{"dir/z.txt", `dir\z.txt`}, Destination: "backup"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"dir", "backup"},
			wantResponsePaths: []string{"backup"},
			wantEvents: []resourceEvent{
				{action: "updated", path: "backup"},
			},
		},
		{
			name: "multi-source: source and sources merged and deduplicated",
			seeds: []seed{
				{path: "a.txt", content: "aaa"},
				{path: "b.txt", content: "bbb"},
			},
			req:               moveRequest{Source: "a.txt", Sources: []string{"a.txt", "b.txt"}, Destination: "archive"},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"archive", "archive/a.txt", "archive/b.txt"},
			wantResponsePaths: []string{"archive", "archive/a.txt", "archive/b.txt"},
			wantEvents: []resourceEvent{
				{action: "added", path: "archive"},
				{action: "updated", path: "archive/a.txt"},
				{action: "updated", path: "archive/b.txt"},
			},
		},
		{
			name: "multi-source: file and directory moved to existing directory",
			seeds: []seed{
				{path: "report.txt", content: "r"},
				{path: "docs", isDir: true},
				{path: "docs/readme.md", content: "readme"},
				{path: "dest", isDir: true},
			},
			req:        moveRequest{Sources: []string{"report.txt", "docs"}, Destination: "dest"},
			wantStatus: http.StatusOK,
			wantPaths:  []string{"dest", "dest/report.txt", "dest/docs", "dest/docs/readme.md"},
			wantResponsePaths: []string{
				"dest/report.txt",
				"dest/docs",
				"dest/docs/readme.md",
			},
			wantEvents: []resourceEvent{
				{action: "updated", path: "dest/report.txt"},
				{action: "updated", path: "dest/docs"},
				{action: "updated", path: "dest/docs/readme.md"},
			},
		},
		{
			name: "multi-source: force overwrites existing file at target",
			seeds: []seed{
				{path: "a.txt", content: "new-a"},
				{path: "b.txt", content: "new-b"},
				{path: "archive/a.txt", content: "old-a"},
			},
			req:               moveRequest{Sources: []string{"a.txt", "b.txt"}, Destination: "archive", Force: true},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"archive", "archive/a.txt", "archive/b.txt"},
			wantResponsePaths: []string{"archive", "archive/a.txt", "archive/b.txt"},
			// publish order: deleted before, added, updated, deleted after
			wantEvents: []resourceEvent{
				{action: "deleted", path: "archive/a.txt"},
				{action: "added", path: "archive"},
				{action: "updated", path: "archive/a.txt"},
				{action: "updated", path: "archive/b.txt"},
			},
			wantDeletedBlobTexts: []string{"old-a"},
		},
		{
			name: "multi-source: without force returns conflict when target exists",
			seeds: []seed{
				{path: "a.txt", content: "a"},
				{path: "b.txt", content: "b"},
				// "archive" dir record absent; "archive/a.txt" causes conflict.
				{path: "archive/a.txt", content: "old"},
			},
			req:        moveRequest{Sources: []string{"a.txt", "b.txt"}, Destination: "archive"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "archive/a.txt", "b.txt"},
		},
		{
			name: "multi-source: duplicate basenames returns conflict",
			seeds: []seed{
				{path: "dir1/x.txt", content: "x1"},
				{path: "dir2/x.txt", content: "x2"},
			},
			req:        moveRequest{Sources: []string{"dir1/x.txt", "dir2/x.txt"}, Destination: "archive"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"dir1/x.txt", "dir2/x.txt"},
		},
		{
			name: "multi-source: one missing source returns not found",
			seeds: []seed{
				{path: "a.txt", content: "a"},
			},
			req:        moveRequest{Sources: []string{"a.txt", "ghost.txt"}, Destination: "archive"},
			wantStatus: http.StatusNotFound,
			wantPaths:  []string{"a.txt"},
		},
		{
			name:       "multi-source: empty sources returns bad request",
			req:        moveRequest{Sources: []string{}, Destination: "archive"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "multi-source: blank entries only returns bad request",
			req:        moveRequest{Sources: []string{"   ", ""}, Destination: "archive"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "multi-source: destination is a file returns conflict",
			seeds: []seed{
				{path: "a.txt", content: "a"},
				{path: "b.txt", content: "b"},
				{path: "dest.file", content: "file"},
			},
			req:        moveRequest{Sources: []string{"a.txt", "b.txt"}, Destination: "dest.file"},
			wantStatus: http.StatusConflict,
			wantPaths:  []string{"a.txt", "b.txt", "dest.file"},
		},
		{
			name: "multi-source: directory into itself returns bad request",
			seeds: []seed{
				{path: "docs", isDir: true},
				{path: "a.txt", content: "a"},
			},
			req:        moveRequest{Sources: []string{"docs", "a.txt"}, Destination: "docs/sub"},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"a.txt", "docs"},
		},
		{
			name: "multi-source: invalid source path returns bad request",
			seeds: []seed{
				{path: "a.txt", content: "a"},
			},
			req:        moveRequest{Sources: []string{"../escape.txt", "a.txt"}, Destination: "archive"},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"a.txt"},
		},
		{
			name:              "single source: file in subdirectory moved to root",
			seeds:             []seed{{path: "reports", isDir: true}, {path: "reports/openai-report.md", content: "rpt"}},
			req:               moveRequest{Source: "reports/openai-report.md", Destination: ""},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"reports", "openai-report.md"},
			wantResponsePaths: []string{"openai-report.md"},
			wantEvents:        []resourceEvent{{action: "updated", path: "openai-report.md"}},
		},
		{
			name:       "single source: file already at root moved to root returns same-location error",
			seeds:      []seed{{path: "report.md", content: "rpt"}},
			req:        moveRequest{Source: "report.md", Destination: ""},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"report.md"},
		},
		{
			name: "single source: directory in subdirectory moved to root",
			seeds: []seed{
				{path: "sub", isDir: true},
				{path: "sub/docs", isDir: true},
				{path: "sub/docs/a.txt", content: "a"},
			},
			req:               moveRequest{Source: "sub/docs", Destination: ""},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"sub", "docs", "docs/a.txt"},
			wantResponsePaths: []string{"docs", "docs/a.txt"},
			wantEvents: []resourceEvent{
				{action: "updated", path: "docs"},
				{action: "updated", path: "docs/a.txt"},
			},
		},
		{
			name:       "single source: directory already at root moved to root returns same-location error",
			seeds:      []seed{{path: "docs", isDir: true}},
			req:        moveRequest{Source: "docs", Destination: ""},
			wantStatus: http.StatusBadRequest,
			wantPaths:  []string{"docs"},
		},
		{
			name: "multi-source: files moved to root",
			seeds: []seed{
				{path: "sub", isDir: true},
				{path: "sub/a.txt", content: "aaa"},
				{path: "sub/b.txt", content: "bbb"},
			},
			req:               moveRequest{Sources: []string{"sub/a.txt", "sub/b.txt"}, Destination: ""},
			wantStatus:        http.StatusOK,
			wantPaths:         []string{"sub", "a.txt", "b.txt"},
			wantResponsePaths: []string{"a.txt", "b.txt"},
			wantEvents: []resourceEvent{
				{action: "updated", path: "a.txt"},
				{action: "updated", path: "b.txt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourcesRunTransferCase(t, tt, http.MethodPut, "/resources/move", (*ResourceService).MoveResource)
		})
	}
}

// MoveResource dedupes its sources, so only a direct call reaches this guard.
func TestResources_MoveMultipleSources_RefusesSourcesNamingOneResource(t *testing.T) {
	svc, db := resourcesServiceHoldingOneFile(t)

	_, err := svc.moveMultipleSources(1, []string{"a.txt", "a.txt"}, "archive", false)

	assert.ErrorIs(t, err, errResourceInvalid)
	assert.ErrorContains(t, err, "sources resolve to the same resource")
	assert.ElementsMatch(t, []string{"a.txt"}, allResourcePaths(t, db))
}

// CopyResource dedupes its sources, so only a direct call reaches this guard.
func TestResources_CopyMultipleSources_RefusesSourcesNamingOneResource(t *testing.T) {
	svc, db := resourcesServiceHoldingOneFile(t)

	_, err := svc.copyMultipleSources(1, []string{"a.txt", "a.txt"}, "backup", false)

	assert.ErrorIs(t, err, errResourceInvalid)
	assert.ErrorContains(t, err, "sources resolve to the same resource")
	assert.ElementsMatch(t, []string{"a.txt"}, allResourcePaths(t, db))
}

func resourcesServiceHoldingOneFile(t *testing.T) (*ResourceService, *gorm.DB) {
	t.Helper()

	db := setupResourceServiceTestDB(t)
	svc := NewResourceService(db, t.TempDir(), &captureSubscriptions{})
	seedResource(t, db, models.UserResource{
		UserID: 1,
		Name:   "a.txt",
		Path:   "a.txt",
		Hash:   md5HexForService("src"),
		Size:   3,
	})
	return svc, db
}

func TestResources_CleanupOrphanBlobs_RemovesOnlyUnreferencedBlobs(t *testing.T) {
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

func TestResources_DeleteOrphanBlob_RemovesOnlyAnUnreferencedBlob(t *testing.T) {
	tests := []struct {
		name       string
		hashOf     string // content whose hash is passed; "" passes an empty hash
		referenced bool   // a row references the blob
		wantKept   bool
	}{
		{name: "removes orphan hash blob", hashOf: "orphan"},
		{name: "keeps blob when still referenced", hashOf: "kept", referenced: true, wantKept: true},
		{name: "empty hash is a no-op", wantKept: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupResourceServiceTestDB(t)
			dataDir := t.TempDir()
			svc := NewResourceService(db, dataDir, nil)
			blob := md5HexForService("blob")
			hash := ""
			if tt.hashOf != "" {
				blob = md5HexForService(tt.hashOf)
				hash = blob
			}
			writeResourceBlob(t, dataDir, blob, "blob")
			if tt.referenced {
				seedResource(t, db, models.UserResource{UserID: 1, Hash: blob, Name: "kept.txt", Path: "kept.txt", Size: 4})
			}

			svc.deleteOrphanBlob(context.Background(), hash)

			_, err := os.Lstat(resources.BlobPath(dataDir, blob))
			if tt.wantKept {
				assert.NoError(t, err)
			} else {
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestResources_ConvertResourceToModel_CopiesEveryField(t *testing.T) {
	modelResource := convertResourceToModel(models.ResourceEntry{
		ID:        42,
		UserID:    7,
		Name:      "docs",
		Path:      "work/docs",
		Size:      123,
		IsDir:     true,
		CreatedAt: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2024, 6, 7, 8, 9, 10, 0, time.UTC),
	})

	assert.Equal(t, &model.UserResource{
		ID:        42,
		UserID:    7,
		Name:      "docs",
		Path:      "work/docs",
		Size:      123,
		IsDir:     true,
		CreatedAt: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2024, 6, 7, 8, 9, 10, 0, time.UTC),
	}, modelResource)
}

// The SQL listing and the REST doors must pick the same row when two owners share a path.
func TestResources_GetAllResources_OrdersByIDLast(t *testing.T) {
	raw, err := os.ReadFile("../../../sqlc/models/resources.sql")
	require.NoError(t, err, "the gate reads the queries themselves — a moved file must fail loudly")

	name := regexp.MustCompile(`(?m)^-- name: (\S+)`)
	blocks := name.Split(string(raw), -1)
	names := name.FindAllStringSubmatch(string(raw), -1)
	require.Len(t, blocks, len(names)+1)

	checked := 0
	for i, match := range names {
		query := names[i][1]
		if !strings.HasPrefix(query, "GetAllResources") {
			continue
		}

		checked++
		body := strings.TrimSpace(blocks[i+1])

		if !strings.Contains(body, "ORDER BY") {
			t.Errorf("%s has no order at all, so two owners on one path come back in whichever order the database chose", match[1])
			continue
		}
		if !strings.Contains(body, "id ASC;") {
			t.Errorf("%s orders by updated_at and name only: two rows equal on both leave the winner to the database", match[1])
		}
	}

	require.NotZero(t, checked, "no cross-owner query was found — the naming this gate keys on has changed")
}

func TestResources_IsUniqueViolation_MatchesPostgresAndSQLitePhrasings(t *testing.T) {
	assert.True(t, isUniqueViolation(errors.New(`pq: duplicate key value violates unique constraint "users_mail_unique"`)))
	assert.True(t, isUniqueViolation(errors.New("pq: error 23505")))
	assert.True(t, isUniqueViolation(errors.New("UNIQUE constraint failed: users.mail")), "sqlite phrasing is matched case-insensitively")
	assert.False(t, isUniqueViolation(errors.New("connection refused")))
	assert.False(t, isUniqueViolation(nil))
}

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

func newResourceTestContext(method, target string, body *bytes.Buffer, privs []string) (*gin.Context, *httptest.ResponseRecorder) {
	return newResourceTestContextWithUID(method, target, body, privs, 1)
}

type captureSubscriptions struct {
	events        []resourceEvent
	publisherUIDs []int64
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
func (s *captureSubscriptions) NewResourcePublisher(userID int64) subscriptions.ResourcePublisher {
	s.publisherUIDs = append(s.publisherUIDs, userID)
	return &captureResourcePublisher{events: &s.events, userID: userID}
}
func (s *captureSubscriptions) NewKnowledgeSubscriber(int64) subscriptions.KnowledgeSubscriber {
	return nil
}
func (s *captureSubscriptions) NewKnowledgePublisher(int64) subscriptions.KnowledgePublisher {
	return nil
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

type resourceEvent struct {
	action string
	path   string
}

type captureResourcePublisher struct {
	userID int64
	events *[]resourceEvent
}

func (p *captureResourcePublisher) GetUserID() int64 { return p.userID }
func (p *captureResourcePublisher) SetUserID(userID int64) {
	p.userID = userID
}
func (p *captureResourcePublisher) ResourceAdded(_ context.Context, resource *model.UserResource) {
	*p.events = append(*p.events, resourceEvent{action: "added", path: resource.Path})
}
func (p *captureResourcePublisher) ResourceUpdated(_ context.Context, resource *model.UserResource) {
	*p.events = append(*p.events, resourceEvent{action: "updated", path: resource.Path})
}
func (p *captureResourcePublisher) ResourceDeleted(_ context.Context, resource *model.UserResource) {
	*p.events = append(*p.events, resourceEvent{action: "deleted", path: resource.Path})
}
