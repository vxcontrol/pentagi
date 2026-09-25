package services

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/flowfiles"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/resources"
	"pentagi/pkg/server/models"
	"pentagi/pkg/version"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ subscriptions.FlowPublisher           = (*captureFlowPublisher)(nil)
	_ subscriptions.ResourcePublisher       = (*captureResourcePublisherForFlow)(nil)
	_ subscriptions.SubscriptionsController = (*flowFileCaptureSubscriptions)(nil)
	_ docker.DockerClient                   = (*fakeDockerClient)(nil)
)

func TestFlowFiles_ConvertModelFlowFile_CopiesEveryField(t *testing.T) {
	modelFile := convertModelFlowFile(models.FlowFile{
		ID:         "flow-file-id",
		Name:       "reports",
		Path:       "uploads/reports",
		Size:       42,
		IsDir:      true,
		ModifiedAt: time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC),
	})

	assert.Equal(t, &model.FlowFile{
		ID:         "flow-file-id",
		Name:       "reports",
		Path:       "uploads/reports",
		Size:       42,
		IsDir:      true,
		ModifiedAt: time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC),
	}, modelFile)
}

func TestFlowFiles_ConvertContainerFiles_JoinsTheBasePathAndSortsByName(t *testing.T) {
	mtime := time.Now()
	files := convertContainerFiles("/work", []container.PathStat{
		{
			Name:  "zeta.txt",
			Size:  10,
			Mode:  0644,
			Mtime: mtime,
		},
		{
			Name:  "alpha",
			Mode:  os.ModeDir | 0755,
			Mtime: mtime.Add(time.Second),
		},
	})

	require.Len(t, files, 2)
	assert.Equal(t, "alpha", files[0].Name)
	assert.Equal(t, "/work/alpha", files[0].Path)
	assert.Equal(t, flowfiles.ID("/work/alpha"), files[0].ID)
	assert.True(t, files[0].IsDir)
	assert.Equal(t, int64(0), files[0].Size)

	assert.Equal(t, "zeta.txt", files[1].Name)
	assert.Equal(t, "/work/zeta.txt", files[1].Path)
	assert.Equal(t, flowfiles.ID("/work/zeta.txt"), files[1].ID)
	assert.False(t, files[1].IsDir)
	assert.Equal(t, int64(10), files[1].Size)
}

func TestFlowFiles_PrimaryContainerName_NamesTheFlowsTerminal(t *testing.T) {
	assert.Equal(t, "pentagi-terminal-42", primaryContainerName("", 42))
}

func TestFlowFiles_SortFlowFiles_PutsTheNewestFirstAndBreaksTiesByName(t *testing.T) {
	now := time.Now()
	files := []models.FlowFile{
		{Name: "b.txt", ModifiedAt: now.Add(-2 * time.Hour)},
		{Name: "a.txt", ModifiedAt: now.Add(-1 * time.Hour)},
		{Name: "c.txt", ModifiedAt: now.Add(-1 * time.Hour)},
	}
	sortFlowFiles(files)

	assert.Equal(t, "a.txt", files[0].Name)
	assert.Equal(t, "c.txt", files[1].Name)
	assert.Equal(t, "b.txt", files[2].Name)
}

func TestFlowFiles_ShellQuote_QuotesForAPOSIXShell(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain path", "/work/uploads/report.txt", "'/work/uploads/report.txt'"},
		{"path with space", "/work/uploads/my report.txt", "'/work/uploads/my report.txt'"},
		{"path with single quote", "/tmp/it's-mine.txt", `'/tmp/it'"'"'s-mine.txt'`},
		{"empty", "", "''"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shellQuote(tt.input))
		})
	}
}

func TestFlowFiles_ParseFlowIDParam_AcceptsOnlyAnUnsignedInteger(t *testing.T) {
	tests := []struct {
		name    string
		param   string
		want    uint64
		wantErr bool
	}{
		{"numeric id", "42", 42, false},
		{"zero", "0", 0, false},
		{"negative", "-1", 0, true},
		{"non-numeric", "abc", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Params = gin.Params{{Key: "flowID", Value: tt.param}}

			got, err := parseFlowIDParam(c)
			if tt.wantErr {
				require.ErrorIs(t, err, strconv.ErrSyntax)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFlowFiles_CleanupPendingUploads_RemovesOnlyUncommittedTempFiles(t *testing.T) {
	dir := t.TempDir()
	keep := filepath.Join(dir, "keep.txt")
	tmp1 := filepath.Join(dir, "tmp1.txt")
	tmp2 := filepath.Join(dir, "tmp2.txt")
	require.NoError(t, os.WriteFile(keep, []byte("keep"), 0644))
	require.NoError(t, os.WriteFile(tmp1, []byte("a"), 0644))
	require.NoError(t, os.WriteFile(tmp2, []byte("b"), 0644))

	cleanupPendingUploads([]pendingUpload{
		{tmpPath: tmp1},
		{tmpPath: ""}, // committed already, must be skipped
		{tmpPath: tmp2},
	})

	_, err := os.Lstat(tmp1)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Lstat(tmp2)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Lstat(keep)
	assert.NoError(t, err, "non-pending files must not be touched")
}

func TestFlowFiles_FlowScopeForFiles_GrantsByPrivilegeAndOwnership(t *testing.T) {
	const caller = 42

	tests := []struct {
		name        string
		privs       []string
		writeAccess bool
		wantDenied  bool
		wantForeign bool // the scope also reaches a flow another user owns
	}{
		{name: "admin reaches any flow on read", privs: []string{"flow_files.admin"}, wantForeign: true},
		{name: "admin reaches any flow on write", privs: []string{"flow_files.admin"}, writeAccess: true, wantForeign: true},
		{name: "upload privilege writes to the caller's own flow only", privs: []string{"flow_files.upload"}, writeAccess: true},
		{name: "view privilege reads the caller's own flow only", privs: []string{"flow_files.view"}},
		{name: "view privilege does not grant write", privs: []string{"flow_files.view"}, writeAccess: true, wantDenied: true},
		{name: "upload privilege does not grant read", privs: []string{"flow_files.upload"}, wantDenied: true},
		{name: "no privileges denies access", privs: []string{}, writeAccess: true, wantDenied: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantDenied {
				assert.Nil(t, flowScopeForFiles(tt.privs, caller, 1, tt.writeAccess))
				return
			}

			db := setupFlowFileServiceTestDB(t)
			seedFlow(t, db, 1, caller)
			seedFlow(t, db, 2, 7)

			reaches := func(flowID uint64) bool {
				scope := flowScopeForFiles(tt.privs, caller, flowID, tt.writeAccess)
				require.NotNil(t, scope)

				var flow models.Flow
				err := db.Model(&flow).Scopes(scope).Take(&flow).Error
				if gorm.IsRecordNotFoundError(err) {
					return false
				}
				require.NoError(t, err)
				assert.Equal(t, flowID, flow.ID)

				return true
			}

			assert.True(t, reaches(1), "the caller's own flow")
			assert.Equal(t, tt.wantForeign, reaches(2), "another user's flow")
		})
	}
}

func decodeFlowFilesResponse(t *testing.T, w *httptest.ResponseRecorder) models.FlowFiles {
	t.Helper()

	var resp struct {
		Status string           `json:"status"`
		Data   models.FlowFiles `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "success", resp.Status)
	return resp.Data
}

type tarTestEntry struct {
	name     string
	typeflag byte
	content  string
	linkname string
}

// buildContainerTar packages entries the way docker's CopyFromContainer streams them.
func buildContainerTar(entries []tarTestEntry) []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range entries {
		hdr := &tar.Header{
			Name:     e.name,
			Typeflag: e.typeflag,
			Mode:     0644,
			Size:     int64(len(e.content)),
			Linkname: e.linkname,
		}
		if e.typeflag == tar.TypeDir {
			hdr.Mode = 0755
		}
		_ = tw.WriteHeader(hdr)
		if len(e.content) > 0 {
			_, _ = tw.Write([]byte(e.content))
		}
	}
	tw.Close()
	return buf.Bytes()
}

func flowFilesTarNames(t *testing.T, archive []byte) []string {
	t.Helper()

	names := []string{}
	tr := tar.NewReader(bytes.NewReader(archive))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return names
		}
		require.NoError(t, err)
		names = append(names, hdr.Name)
	}
}

// flowFilesTree maps every entry under root to its content; a directory's key
// ends in a slash.
func flowFilesTree(t *testing.T, root string) map[string]string {
	t.Helper()

	tree := map[string]string{}
	require.NoError(t, filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		rel, err := filepath.Rel(root, p)
		require.NoError(t, err)
		switch {
		case rel == ".":
		case info.IsDir():
			tree[filepath.ToSlash(rel)+"/"] = ""
		default:
			data, err := os.ReadFile(p)
			require.NoError(t, err)
			tree[filepath.ToSlash(rel)] = string(data)
		}
		return nil
	}))

	return tree
}

func flowFilesContainerPathsQuery(n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = "paths[]=/p" + strconv.Itoa(i)
	}
	return strings.Join(parts, "&")
}

func TestFlowFiles_GetFlowFiles_ListsTheFlowCacheToPermittedCallers(t *testing.T) {
	type seedFile struct {
		dir     string // "uploads", "container/<sub>", "resources/<sub>"
		name    string
		content string
	}

	tests := []struct {
		name              string
		seedFiles         []seedFile
		seedFlow          *struct{ id, userID uint64 }
		flowID            uint64
		uid               uint64
		privs             []string
		flowIDOverride    string
		wantStatus        int
		wantResponsePaths []string
	}{
		{
			name:              "view privilege lists own flow files",
			seedFlow:          &struct{ id, userID uint64 }{1, 1},
			flowID:            1,
			uid:               1,
			privs:             []string{"flow_files.view"},
			seedFiles:         []seedFile{{dir: "uploads", name: "report.txt", content: "r"}},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"uploads/report.txt"},
		},
		{
			name:              "admin lists other user's flow files",
			seedFlow:          &struct{ id, userID uint64 }{1, 2},
			flowID:            1,
			uid:               1,
			privs:             []string{"flow_files.admin"},
			seedFiles:         []seedFile{{dir: "uploads", name: "report.txt", content: "r"}},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"uploads/report.txt"},
		},
		{
			name:       "non-admin cannot list other user's flow",
			seedFlow:   &struct{ id, userID uint64 }{1, 2},
			flowID:     1,
			uid:        1,
			privs:      []string{"flow_files.view"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "missing privilege returns forbidden",
			seedFlow:   &struct{ id, userID uint64 }{1, 1},
			flowID:     1,
			uid:        1,
			privs:      []string{"resources.view"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:           "non-numeric flow id returns bad request",
			flowIDOverride: "abc",
			privs:          []string{"flow_files.view"},
			uid:            1,
			wantStatus:     http.StatusBadRequest,
		},
		{
			name:       "missing flow returns not found",
			flowID:     999,
			uid:        1,
			privs:      []string{"flow_files.view"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "list aggregates uploads, container, resources",
			seedFlow: &struct{ id, userID uint64 }{1, 1},
			flowID:   1,
			uid:      1,
			privs:    []string{"flow_files.view"},
			seedFiles: []seedFile{
				{dir: "uploads", name: "report.txt", content: "r"},
				{dir: "container/etc", name: "nginx.conf", content: "n"},
				{dir: "resources/creds", name: "p.txt", content: "p"},
			},
			wantStatus: http.StatusOK,
			wantResponsePaths: []string{
				"uploads/report.txt",
				"container/etc",
				"container/etc/nginx.conf",
				"resources/creds",
				"resources/creds/p.txt",
			},
		},
		{
			name:              "a flow with no cache directories lists nothing",
			seedFlow:          &struct{ id, userID uint64 }{1, 1},
			flowID:            1,
			uid:               1,
			privs:             []string{"flow_files.view"},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &flowFileCaptureSubscriptions{}
			svc := NewFlowFileService(db, dataDir, "", nil, ss)

			if tt.seedFlow != nil {
				seedFlow(t, db, tt.seedFlow.id, tt.seedFlow.userID)
			}
			for _, f := range tt.seedFiles {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), filepath.FromSlash(f.dir))
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, f.name), []byte(f.content), 0644))
			}

			c, w := newFlowFileTestContext(http.MethodGet, "/flows/1/files/", nil, tt.privs, tt.uid, tt.flowID)
			if tt.flowIDOverride != "" {
				c.Params = gin.Params{{Key: "flowID", Value: tt.flowIDOverride}}
			}
			svc.GetFlowFiles(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				resp := decodeFlowFilesResponse(t, w)
				paths := make([]string, len(resp.Files))
				for i, f := range resp.Files {
					paths[i] = f.Path
				}
				assert.ElementsMatch(t, tt.wantResponsePaths, paths)
				assert.Equal(t, uint64(len(tt.wantResponsePaths)), resp.Total)
			}
		})
	}
}

func TestFlowFiles_UploadFlowFiles_StoresEachFileUnderUploadsOnly(t *testing.T) {
	tests := []struct {
		name          string
		flowOwner     uint64 // 0: the caller
		noFlow        bool
		privs         []string // nil: flow_files.upload
		files         []uploadTestFile
		fieldName     string
		rawFileName   string // sent unescaped, with "payload" as content
		nonMultipart  bool
		seedExisting  []string // names already under uploads/, holding "old"
		dockerRunning bool
		wantStatus    int
		wantStored    map[string]string // name under uploads/ -> content
		wantPushed    []string          // tar entries streamed to /work
	}{
		{
			name:          "a plain name is stored as sent and pushed to the running container",
			files:         []uploadTestFile{{name: "report.txt", content: "r"}},
			dockerRunning: true,
			wantStatus:    http.StatusOK,
			wantStored:    map[string]string{"report.txt": "r"},
			wantPushed:    []string{"uploads", "uploads/report.txt"}, // the directory header, then the file
		},
		{
			name:       "the singular file field works as a fallback",
			files:      []uploadTestFile{{name: "report.txt", content: "r"}},
			fieldName:  "file",
			wantStatus: http.StatusOK,
			wantStored: map[string]string{"report.txt": "r"},
		},
		{
			name:       "several files are stored in one request",
			files:      []uploadTestFile{{name: "a.txt", content: "a"}, {name: "b.txt", content: "b"}},
			wantStatus: http.StatusOK,
			wantStored: map[string]string{"a.txt": "a", "b.txt": "b"},
		},
		{
			name:       "admin can upload to another user's flow",
			flowOwner:  2,
			privs:      []string{"flow_files.admin"},
			files:      []uploadTestFile{{name: "report.txt", content: "r"}},
			wantStatus: http.StatusOK,
			wantStored: map[string]string{"report.txt": "r"},
		},
		{
			name:       "view privilege cannot upload",
			privs:      []string{"flow_files.view"},
			files:      []uploadTestFile{{name: "report.txt", content: "r"}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "missing flow returns not found",
			noFlow:     true,
			files:      []uploadTestFile{{name: "report.txt", content: "r"}},
			wantStatus: http.StatusNotFound,
		},
		{
			name:         "non-multipart body returns bad request",
			nonMultipart: true,
			wantStatus:   http.StatusBadRequest,
		},
		{
			name:       "empty multipart returns bad request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:         "a name already in uploads conflicts and keeps the old file",
			seedExisting: []string{"report.txt"},
			files:        []uploadTestFile{{name: "report.txt", content: "new"}},
			wantStatus:   http.StatusConflict,
		},
		{
			name:       "parent traversal is reduced to basename",
			files:      []uploadTestFile{{name: "../report.txt", content: "payload"}},
			wantStatus: http.StatusOK,
			wantStored: map[string]string{"report.txt": "payload"},
		},
		{
			name:       "deep parent traversal targeting /etc/passwd is reduced to basename",
			files:      []uploadTestFile{{name: "../../etc/passwd", content: "payload"}},
			wantStatus: http.StatusOK,
			wantStored: map[string]string{"passwd": "payload"},
		},
		{
			name:       "absolute unix path is reduced to basename",
			files:      []uploadTestFile{{name: "/etc/shadow", content: "payload"}},
			wantStatus: http.StatusOK,
			wantStored: map[string]string{"shadow": "payload"},
		},
		{
			name:       "windows-style backslash separators are reduced to basename",
			files:      []uploadTestFile{{name: `nested\..\evil.txt`, content: "payload"}},
			wantStatus: http.StatusOK,
			wantStored: map[string]string{"evil.txt": "payload"},
		},
		{
			name:       "double-slash and dotdot mix collapses to basename",
			files:      []uploadTestFile{{name: "..//..//.//attack.bin", content: "payload"}},
			wantStatus: http.StatusOK,
			wantStored: map[string]string{"attack.bin": "payload"},
		},
		{
			name:       "leading dot file is preserved verbatim",
			files:      []uploadTestFile{{name: ".env", content: "payload"}},
			wantStatus: http.StatusOK,
			wantStored: map[string]string{".env": "payload"},
		},
		{
			name:       "literal parent directory is rejected",
			files:      []uploadTestFile{{name: "..", content: "payload"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "literal current directory is rejected",
			files:      []uploadTestFile{{name: ".", content: "payload"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "lone slash is rejected",
			files:      []uploadTestFile{{name: "/", content: "payload"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "embedded NUL byte is rejected",
			files:      []uploadTestFile{{name: "evil\x00.txt", content: "payload"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "newline in filename arrives percent-encoded and stays inside uploads",
			files:      []uploadTestFile{{name: "evil\nname.txt", content: "payload"}},
			wantStatus: http.StatusOK,
			wantStored: map[string]string{"evil%0Aname.txt": "payload"},
		},
		{
			name:       "carriage return in filename arrives percent-encoded and stays inside uploads",
			files:      []uploadTestFile{{name: "evil\rname.txt", content: "payload"}},
			wantStatus: http.StatusOK,
			wantStored: map[string]string{"evil%0Dname.txt": "payload"},
		},
		{
			name:        "a raw line feed in the filename is rejected",
			rawFileName: "evil\nname.txt",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "a raw carriage return in the filename is rejected",
			rawFileName: "evil\rname.txt",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:       "DEL control character is rejected",
			files:      []uploadTestFile{{name: "evil\x7fname.txt", content: "payload"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "windows reserved colon is rejected",
			files:      []uploadTestFile{{name: "con:1.txt", content: "payload"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wildcard star is rejected",
			files:      []uploadTestFile{{name: "evil*.txt", content: "payload"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "redirection chevrons are rejected",
			files:      []uploadTestFile{{name: "evil<>name.txt", content: "payload"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "filename longer than MaxFileNameLength is rejected",
			files:      []uploadTestFile{{name: strings.Repeat("a", flowfiles.MaxFileNameLength+1) + ".txt", content: "payload"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "blank filename is rejected",
			files:      []uploadTestFile{{name: "   ", content: "payload"}},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &flowFileCaptureSubscriptions{}
			fakeDocker := &fakeDockerClient{running: tt.dockerRunning}
			svc := NewFlowFileService(db, dataDir, "", fakeDocker, ss)

			if !tt.noFlow {
				owner := tt.flowOwner
				if owner == 0 {
					owner = 1
				}
				seedFlow(t, db, 1, owner)
			}
			want := map[string]string{}
			for _, name := range tt.seedExisting {
				dir := filepath.Join(dataDir, "flow-1-data", "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("old"), 0644))
				want["flow-1-data/uploads/"+name] = "old"
			}

			var body io.Reader
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
			privs := tt.privs
			if privs == nil {
				privs = []string{"flow_files.upload"}
			}

			c, w := newFlowFileTestContext(http.MethodPost, "/flows/1/files/", body, privs, 1, 1)
			c.Request.Header.Set("Content-Type", contentType)
			svc.UploadFlowFiles(c)

			require.Equal(t, tt.wantStatus, w.Code)

			stored := map[string]string{}
			for p, content := range flowFilesTree(t, dataDir) {
				assert.True(t, strings.HasPrefix(p, "flow-1-data/"), "%q escaped the flow data directory", p)
				if !strings.HasSuffix(p, "/") {
					stored[p] = content
				}
			}
			wantPaths := []string{}
			wantEvents := []string{}
			for name, content := range tt.wantStored {
				want["flow-1-data/uploads/"+name] = content
				wantPaths = append(wantPaths, "uploads/"+name)
				wantEvents = append(wantEvents, "flow added uploads/"+name)
			}
			assert.Equal(t, want, stored, "regular files under the data directory")

			events := []string{}
			for _, ev := range ss.snapshot() {
				events = append(events, ev.channel+" "+ev.action+" "+ev.path)
			}
			assert.ElementsMatch(t, wantEvents, events)

			if tt.wantPushed == nil {
				assert.Empty(t, fakeDocker.copyToCalls)
			} else {
				require.Len(t, fakeDocker.copyToCalls, 1)
				assert.Equal(t, "/work", fakeDocker.copyToCalls[0].dstPath)
				assert.Equal(t, tt.wantPushed, flowFilesTarNames(t, fakeDocker.copyToCalls[0].body))
			}

			if tt.wantStatus != http.StatusOK {
				return
			}
			resp := decodeFlowFilesResponse(t, w)
			paths := make([]string, len(resp.Files))
			for i, f := range resp.Files {
				paths[i] = f.Path
			}
			assert.ElementsMatch(t, wantPaths, paths)
		})
	}
}

func TestFlowFiles_DeleteFlowFile_RemovesFromTheCacheAndTheContainer(t *testing.T) {
	type seedFile struct {
		relPath string
		content string
		isDir   bool
	}

	tests := []struct {
		name              string
		flowOwner         uint64
		uid               uint64
		flowID            uint64
		privs             []string
		seedFlow          bool
		seedFiles         []seedFile
		queryPath         string // builds ?path=<value>; empty = no path param
		rawQuery          string // when non-empty, used verbatim as query string (overrides queryPath)
		dockerRunning     bool
		execCreateErr     error // simulate ContainerExecCreate failure
		execInspectCode   int   // simulate non-zero exec exit code (must have dockerRunning:true)
		wantStatus        int
		wantDeletedPath   string   // single path expected in "deleted" subscription events
		wantDeletedPaths  []string // additional paths expected in "deleted" events (for bulk tests)
		wantFilesGone     []string // relative paths inside flow data dir that must be absent
		wantFilesExist    []string // relative paths that must still exist (fail-safe atomicity)
		wantExec          []string // every docker exec command, in order; nil: none
		wantResponseTotal int      // > 0: verify response Total and Files length
	}{
		{
			name:            "delete uploaded file",
			flowOwner:       1,
			uid:             1,
			flowID:          1,
			privs:           []string{"flow_files.upload"},
			seedFlow:        true,
			seedFiles:       []seedFile{{relPath: "uploads/report.txt", content: "r"}},
			queryPath:       "uploads/report.txt",
			wantStatus:      http.StatusOK,
			wantDeletedPath: "uploads/report.txt",
		},
		{
			name:      "delete container directory recursively",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.upload"},
			seedFlow: true,
			seedFiles: []seedFile{
				{relPath: "container/etc", isDir: true},
				{relPath: "container/etc/nginx.conf", content: "n"},
			},
			queryPath:       "container/etc",
			wantStatus:      http.StatusOK,
			wantDeletedPath: "container/etc",
		},
		{
			name:      "delete resources file",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:           []string{"flow_files.upload"},
			seedFlow:        true,
			seedFiles:       []seedFile{{relPath: "resources/creds/p.txt", content: "p"}},
			queryPath:       "resources/creds/p.txt",
			wantStatus:      http.StatusOK,
			wantDeletedPath: "resources/creds/p.txt",
		},
		{
			name:      "delete uploads file invokes container exec when running",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:           []string{"flow_files.upload"},
			seedFlow:        true,
			seedFiles:       []seedFile{{relPath: "uploads/report.txt", content: "r"}},
			queryPath:       "uploads/report.txt",
			dockerRunning:   true,
			wantStatus:      http.StatusOK,
			wantDeletedPath: "uploads/report.txt",
			wantExec:        []string{"sh -c rm -rf -- '/work/uploads/report.txt'"},
		},
		{
			name:      "view privilege cannot delete",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view"},
			seedFlow:   true,
			queryPath:  "uploads/report.txt",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "missing flow returns not found",
			uid:        1,
			flowID:     99,
			privs:      []string{"flow_files.upload"},
			queryPath:  "uploads/report.txt",
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "missing path returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload"},
			seedFlow:   true,
			queryPath:  "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "wrong-prefix path returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload"},
			seedFlow:   true,
			queryPath:  "tmp/x.txt",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "path traversal returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload"},
			seedFlow:   true,
			queryPath:  "uploads/../../etc/passwd",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "non-existent file returns not found",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload"},
			seedFlow:   true,
			queryPath:  "uploads/missing.txt",
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "only paths[] param used deletes single file",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:           []string{"flow_files.upload"},
			seedFlow:        true,
			seedFiles:       []seedFile{{relPath: "uploads/report.txt", content: "r"}},
			rawQuery:        "paths[]=uploads/report.txt",
			wantStatus:      http.StatusOK,
			wantDeletedPath: "uploads/report.txt",
			wantFilesGone:   []string{"uploads/report.txt"},
		},
		{
			name:      "batch delete two uploads files via paths[]",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.upload"},
			seedFlow: true,
			seedFiles: []seedFile{
				{relPath: "uploads/a.txt", content: "a"},
				{relPath: "uploads/b.txt", content: "b"},
			},
			rawQuery:          "paths[]=uploads/a.txt&paths[]=uploads/b.txt",
			wantStatus:        http.StatusOK,
			wantDeletedPaths:  []string{"uploads/a.txt", "uploads/b.txt"},
			wantFilesGone:     []string{"uploads/a.txt", "uploads/b.txt"},
			wantResponseTotal: 2,
		},
		{
			name:      "path and paths[] combined delete two files",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.upload"},
			seedFlow: true,
			seedFiles: []seedFile{
				{relPath: "uploads/a.txt", content: "a"},
				{relPath: "uploads/b.txt", content: "b"},
			},
			rawQuery:          "path=uploads/a.txt&paths[]=uploads/b.txt",
			wantStatus:        http.StatusOK,
			wantDeletedPaths:  []string{"uploads/a.txt", "uploads/b.txt"},
			wantFilesGone:     []string{"uploads/a.txt", "uploads/b.txt"},
			wantResponseTotal: 2,
		},
		{
			name:      "absolute path in the batch refuses the whole request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:          []string{"flow_files.upload"},
			seedFlow:       true,
			seedFiles:      []seedFile{{relPath: "uploads/report.txt", content: "r"}},
			rawQuery:       "paths[]=uploads/report.txt&paths[]=/etc/passwd",
			wantStatus:     http.StatusBadRequest,
			wantFilesExist: []string{"uploads/report.txt"},
		},
		{
			name:      "parent traversal in the batch refuses the whole request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:          []string{"flow_files.upload"},
			seedFlow:       true,
			seedFiles:      []seedFile{{relPath: "uploads/report.txt", content: "r"}},
			rawQuery:       "paths[]=uploads/report.txt&paths[]=../../etc/shadow",
			wantStatus:     http.StatusBadRequest,
			wantFilesExist: []string{"uploads/report.txt"},
		},
		{
			name:      "duplicate paths[] entries deduplicated",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:             []string{"flow_files.upload"},
			seedFlow:          true,
			seedFiles:         []seedFile{{relPath: "uploads/report.txt", content: "r"}},
			rawQuery:          "paths[]=uploads/report.txt&paths[]=uploads/report.txt",
			wantStatus:        http.StatusOK,
			wantDeletedPath:   "uploads/report.txt",
			wantFilesGone:     []string{"uploads/report.txt"},
			wantResponseTotal: 1,
		},
		{
			name:      "paths[] parent covers child - child not separately processed",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.upload"},
			seedFlow: true,
			seedFiles: []seedFile{
				{relPath: "uploads/dir", isDir: true},
				{relPath: "uploads/dir/file.txt", content: "x"},
			},
			rawQuery:          "paths[]=uploads/dir&paths[]=uploads/dir/file.txt",
			wantStatus:        http.StatusOK,
			wantDeletedPath:   "uploads/dir",
			wantDeletedPaths:  []string{"uploads/dir/file.txt"},
			wantFilesGone:     []string{"uploads/dir", "uploads/dir/file.txt"},
			wantResponseTotal: 2, // dir entry + nested file
		},
		{
			name:      "directory deletion reports all nested files and dirs in response",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.upload"},
			seedFlow: true,
			seedFiles: []seedFile{
				{relPath: "resources/folder", isDir: true},
				{relPath: "resources/folder/scope.md", content: "scope"},
				{relPath: "resources/folder/sub", isDir: true},
				{relPath: "resources/folder/sub/detail.txt", content: "detail"},
			},
			queryPath:       "resources/folder",
			wantStatus:      http.StatusOK,
			wantDeletedPath: "resources/folder",
			wantDeletedPaths: []string{
				"resources/folder/scope.md",
				"resources/folder/sub",
				"resources/folder/sub/detail.txt",
			},
			wantFilesGone:     []string{"resources/folder"},
			wantResponseTotal: 4, // dir + 1 file + sub-dir + 1 nested file
		},
		{
			name:      "response contains metadata for all deleted files",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.upload"},
			seedFlow: true,
			seedFiles: []seedFile{
				{relPath: "uploads/a.txt", content: "a"},
				{relPath: "resources/creds/p.txt", content: "p"},
			},
			rawQuery:          "paths[]=uploads/a.txt&paths[]=resources/creds/p.txt",
			wantStatus:        http.StatusOK,
			wantDeletedPaths:  []string{"uploads/a.txt", "resources/creds/p.txt"},
			wantResponseTotal: 2,
		},
		{
			name:      "single Docker exec issued for two uploads paths",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.upload"},
			seedFlow: true,
			seedFiles: []seedFile{
				{relPath: "uploads/a.txt", content: "a"},
				{relPath: "uploads/b.txt", content: "b"},
			},
			rawQuery:         "paths[]=uploads/a.txt&paths[]=uploads/b.txt",
			dockerRunning:    true,
			wantStatus:       http.StatusOK,
			wantDeletedPaths: []string{"uploads/a.txt", "uploads/b.txt"},
			wantExec:         []string{"sh -c rm -rf -- '/work/uploads/a.txt' '/work/uploads/b.txt'"},
		},
		{
			// container/ paths are host-only and must never be sent to the container.
			name:      "container path does not trigger Docker exec",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.upload"},
			seedFlow: true,
			seedFiles: []seedFile{
				{relPath: "container/etc", isDir: true},
				{relPath: "container/etc/nginx.conf", content: "n"},
			},
			rawQuery:        "paths[]=container/etc",
			dockerRunning:   true,
			wantStatus:      http.StatusOK,
			wantDeletedPath: "container/etc",
		},
		{
			// resources/ paths are mirrored in the container; exec must be triggered.
			name:      "resources path triggers Docker exec",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:           []string{"flow_files.upload"},
			seedFlow:        true,
			seedFiles:       []seedFile{{relPath: "resources/creds/p.txt", content: "p"}},
			rawQuery:        "paths[]=resources/creds/p.txt",
			dockerRunning:   true,
			wantStatus:      http.StatusOK,
			wantDeletedPath: "resources/creds/p.txt",
			wantExec:        []string{"sh -c rm -rf -- '/work/resources/creds/p.txt'"},
		},
		{
			name:      "mixed namespaces produce single exec with only mirrored paths",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.upload"},
			seedFlow: true,
			seedFiles: []seedFile{
				{relPath: "uploads/a.txt", content: "a"},
				{relPath: "container/etc", isDir: true},
				{relPath: "container/etc/nginx.conf", content: "n"},
				{relPath: "resources/creds/p.txt", content: "p"},
			},
			rawQuery:         "paths[]=uploads/a.txt&paths[]=container/etc&paths[]=resources/creds/p.txt",
			dockerRunning:    true,
			wantStatus:       http.StatusOK,
			wantDeletedPaths: []string{"uploads/a.txt", "container/etc", "resources/creds/p.txt"},
			wantFilesGone:    []string{"uploads/a.txt", "container/etc", "resources/creds/p.txt"},
			wantExec:         []string{"sh -c rm -rf -- '/work/uploads/a.txt' '/work/resources/creds/p.txt'"},
		},
		{
			name:      "admin can delete from other user flow",
			flowOwner: 2, uid: 1, flowID: 1,
			privs:           []string{"flow_files.admin"},
			seedFlow:        true,
			seedFiles:       []seedFile{{relPath: "uploads/report.txt", content: "r"}},
			queryPath:       "uploads/report.txt",
			wantStatus:      http.StatusOK,
			wantDeletedPath: "uploads/report.txt",
		},
		{
			name:      "all whitespace paths[] values return bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload"},
			seedFlow:   true,
			rawQuery:   "paths[]=%20%20%20&paths[]=%09",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "invalid prefix in second paths[] fails atomically",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.upload"},
			seedFlow: true,
			seedFiles: []seedFile{
				{relPath: "uploads/a.txt", content: "a"},
			},
			rawQuery:       "paths[]=uploads/a.txt&paths[]=tmp/evil.txt",
			wantStatus:     http.StatusBadRequest,
			wantFilesExist: []string{"uploads/a.txt"},
		},
		{
			name:      "missing file in second paths[] fails atomically",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.upload"},
			seedFlow: true,
			seedFiles: []seedFile{
				{relPath: "uploads/a.txt", content: "a"},
			},
			rawQuery:       "paths[]=uploads/a.txt&paths[]=uploads/missing.txt",
			wantStatus:     http.StatusNotFound,
			wantFilesExist: []string{"uploads/a.txt"},
		},
		{
			// the cache keeps the file, or it would diverge from the container
			name:      "exec create failure returns 500 and preserves local file",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:          []string{"flow_files.upload"},
			seedFlow:       true,
			seedFiles:      []seedFile{{relPath: "uploads/report.txt", content: "r"}},
			queryPath:      "uploads/report.txt",
			dockerRunning:  true,
			execCreateErr:  fmt.Errorf("docker daemon unreachable"),
			wantStatus:     http.StatusInternalServerError,
			wantFilesExist: []string{"uploads/report.txt"},
			wantExec:       []string{"sh -c rm -rf -- '/work/uploads/report.txt'"},
		},
		{
			name:      "exec non-zero exit code returns 500 and preserves local file",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:           []string{"flow_files.upload"},
			seedFlow:        true,
			seedFiles:       []seedFile{{relPath: "uploads/report.txt", content: "r"}},
			queryPath:       "uploads/report.txt",
			dockerRunning:   true,
			execInspectCode: 1,
			wantStatus:      http.StatusInternalServerError,
			wantFilesExist:  []string{"uploads/report.txt"},
			wantExec:        []string{"sh -c rm -rf -- '/work/uploads/report.txt'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &flowFileCaptureSubscriptions{}
			fakeDocker := &fakeDockerClient{
				running:         tt.dockerRunning,
				execCreateErr:   tt.execCreateErr,
				execInspectCode: tt.execInspectCode,
			}
			svc := NewFlowFileService(db, dataDir, "", fakeDocker, ss)

			if tt.seedFlow {
				owner := tt.flowOwner
				if owner == 0 {
					owner = tt.uid
				}
				seedFlow(t, db, tt.flowID, owner)
			}
			for _, f := range tt.seedFiles {
				abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), filepath.FromSlash(f.relPath))
				if f.isDir {
					require.NoError(t, os.MkdirAll(abs, 0755))
					continue
				}
				require.NoError(t, os.MkdirAll(filepath.Dir(abs), 0755))
				require.NoError(t, os.WriteFile(abs, []byte(f.content), 0644))
			}

			target := "/flows/1/files/"
			switch {
			case tt.rawQuery != "":
				target += "?" + tt.rawQuery
			case tt.queryPath != "":
				target += "?path=" + tt.queryPath
			}
			c, w := newFlowFileTestContext(http.MethodDelete, target, nil, tt.privs, tt.uid, tt.flowID)

			svc.DeleteFlowFile(c)

			require.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				if tt.queryPath != "" {
					abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), filepath.FromSlash(tt.queryPath))
					_, err := os.Lstat(abs)
					assert.True(t, os.IsNotExist(err), "deleted file/dir must be gone from cache: %s", tt.queryPath)
				}
				for _, gone := range tt.wantFilesGone {
					abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), filepath.FromSlash(gone))
					_, err := os.Lstat(abs)
					assert.True(t, os.IsNotExist(err), "expected %q to be gone from cache", gone)
				}

				deleted := make([]string, 0)
				for _, ev := range ss.snapshot() {
					if ev.channel == "flow" && ev.action == "deleted" {
						deleted = append(deleted, ev.path)
					}
				}
				if tt.wantDeletedPath != "" {
					assert.Contains(t, deleted, tt.wantDeletedPath)
				}
				for _, p := range tt.wantDeletedPaths {
					assert.Contains(t, deleted, p, "expected deleted event for %q", p)
				}

				if tt.wantResponseTotal > 0 {
					resp := decodeFlowFilesResponse(t, w)
					assert.Equal(t, uint64(tt.wantResponseTotal), resp.Total, "response Total mismatch")
					assert.Len(t, resp.Files, tt.wantResponseTotal, "response Files length mismatch")
				}
			}

			assert.Equal(t, tt.wantExec, fakeDocker.execCommands)

			for _, exists := range tt.wantFilesExist {
				abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), filepath.FromSlash(exists))
				_, err := os.Lstat(abs)
				assert.NoError(t, err, "expected %q to still exist on disk (atomicity guarantee)", exists)
			}
		})
	}
}

func TestFlowFiles_DownloadFlowFile_ServesAFileOrAZip(t *testing.T) {
	tests := []struct {
		name             string
		flowOwner        uint64
		uid              uint64
		flowID           uint64
		privs            []string
		queryPath        string // builds ?path=<value>
		rawQuery         string // when set, used verbatim as query string (overrides queryPath)
		setupFile        func(t *testing.T, dataDir string, flowID uint64)
		wantStatus       int
		wantBody         string
		wantContentType  string
		wantDispContains string
		wantZipEntries   map[string]string
	}{
		{
			name:      "download regular uploaded file",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:     []string{"flow_files.view"},
			queryPath: "uploads/report.txt",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "report.txt"), []byte("payload"), 0644))
			},
			wantStatus:       http.StatusOK,
			wantBody:         "payload",
			wantDispContains: "report.txt",
		},
		{
			name:      "download container directory as zip",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:     []string{"flow_files.view"},
			queryPath: "container/etc",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "container", "etc")
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "nginx"), 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "nginx", "nginx.conf"), []byte("nginx"), 0644))
			},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "etc.zip",
			// a lone directory is zipped relative to itself
			wantZipEntries: map[string]string{"nginx/nginx.conf": "nginx"},
		},
		{
			name:      "admin downloads other user's file",
			flowOwner: 2, uid: 1, flowID: 1,
			privs:     []string{"flow_files.admin"},
			queryPath: "uploads/report.txt",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "report.txt"), []byte("admin"), 0644))
			},
			wantStatus: http.StatusOK,
			wantBody:   "admin",
		},
		{
			name:      "non-admin cannot download other user's file",
			flowOwner: 2, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view"},
			queryPath:  "uploads/report.txt",
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "missing privilege returns forbidden",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"resources.view"},
			queryPath:  "uploads/report.txt",
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "empty path returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view"},
			queryPath:  "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "non-existent path returns not found",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view"},
			queryPath:  "uploads/missing.txt",
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "symlink is rejected",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:     []string{"flow_files.view"},
			queryPath: "uploads/link.txt",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				target := filepath.Join(dir, "real.txt")
				require.NoError(t, os.WriteFile(target, []byte("x"), 0644))
				link := filepath.Join(dir, "link.txt")
				if err := os.Symlink(target, link); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "single file via paths[] downloaded as direct attachment",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view"},
			rawQuery: "paths[]=uploads/report.txt",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "report.txt"), []byte("payload"), 0644))
			},
			wantStatus:       http.StatusOK,
			wantBody:         "payload",
			wantDispContains: "report.txt",
		},
		{
			name:      "single directory via paths[] uses dir-relative zip paths",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view"},
			rawQuery: "paths[]=container/etc",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "container", "etc")
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "nginx"), 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "nginx", "nginx.conf"), []byte("nginx"), 0644))
			},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "etc.zip",
			wantZipEntries:   map[string]string{"nginx/nginx.conf": "nginx"},
		},
		{
			name:      "two files via paths[] packaged into zip with cache-relative paths",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view"},
			rawQuery: "paths[]=uploads/a.txt&paths[]=uploads/b.txt",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("alpha"), 0644))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("bravo"), 0644))
			},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "download.zip",
			wantZipEntries: map[string]string{
				"uploads/a.txt": "alpha",
				"uploads/b.txt": "bravo",
			},
		},
		{
			name:      "path= and paths[] combined produce multi-entry zip",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view"},
			rawQuery: "path=uploads/a.txt&paths[]=resources/b.txt",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				base := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID))
				require.NoError(t, os.MkdirAll(filepath.Join(base, "uploads"), 0755))
				require.NoError(t, os.WriteFile(filepath.Join(base, "uploads", "a.txt"), []byte("alpha"), 0644))
				require.NoError(t, os.MkdirAll(filepath.Join(base, "resources"), 0755))
				require.NoError(t, os.WriteFile(filepath.Join(base, "resources", "b.txt"), []byte("bravo"), 0644))
			},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "download.zip",
			wantZipEntries: map[string]string{
				"uploads/a.txt":   "alpha",
				"resources/b.txt": "bravo",
			},
		},
		{
			name:      "file and directory via paths[] combined in zip",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view"},
			rawQuery: "paths[]=uploads/a.txt&paths[]=container/etc",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				base := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID))
				require.NoError(t, os.MkdirAll(filepath.Join(base, "uploads"), 0755))
				require.NoError(t, os.WriteFile(filepath.Join(base, "uploads", "a.txt"), []byte("alpha"), 0644))
				require.NoError(t, os.MkdirAll(filepath.Join(base, "container", "etc"), 0755))
				require.NoError(t, os.WriteFile(filepath.Join(base, "container", "etc", "nginx.conf"), []byte("nginx"), 0644))
			},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "download.zip",
			wantZipEntries: map[string]string{
				"uploads/a.txt":            "alpha",
				"container/etc/nginx.conf": "nginx",
			},
		},
		{
			name:      "paths[] parent covers child - single dir zip returned",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view"},
			rawQuery: "paths[]=uploads/dir&paths[]=uploads/dir/file.txt",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "uploads", "dir")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644))
			},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "dir.zip",
			wantZipEntries:   map[string]string{"file.txt": "content"},
		},
		{
			name:      "duplicate paths in paths[] produce single-file download",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view"},
			rawQuery: "paths[]=uploads/report.txt&paths[]=uploads/report.txt",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "report.txt"), []byte("payload"), 0644))
			},
			wantStatus:       http.StatusOK,
			wantBody:         "payload",
			wantDispContains: "report.txt",
		},
		{
			name:      "whitespace-only paths[] returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view"},
			rawQuery:   "paths[]=%20%20&paths[]=%09",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "wrong-prefix path in paths[] returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view"},
			rawQuery:   "paths[]=tmp/evil.txt",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "path traversal in paths[] returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view"},
			rawQuery:   "paths[]=uploads/../../etc/passwd",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "absolute path in the batch refuses the whole request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view"},
			rawQuery: "paths[]=uploads/a.txt&paths[]=/etc/passwd",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0644))
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "missing file in batch returns not found",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view"},
			rawQuery: "paths[]=uploads/a.txt&paths[]=uploads/missing.txt",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0644))
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "symlink in paths[] batch returns not found",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view"},
			rawQuery: "paths[]=uploads/real.txt&paths[]=uploads/link.txt",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "real.txt"), []byte("x"), 0644))
				if err := os.Symlink(filepath.Join(dir, "real.txt"), filepath.Join(dir, "link.txt")); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			svc := NewFlowFileService(db, dataDir, "", nil, nil)

			if tt.flowOwner != 0 {
				seedFlow(t, db, tt.flowID, tt.flowOwner)
			}
			if tt.setupFile != nil {
				tt.setupFile(t, dataDir, tt.flowID)
			}

			target := "/flows/1/files/download"
			switch {
			case tt.rawQuery != "":
				target += "?" + tt.rawQuery
			case tt.queryPath != "":
				target += "?path=" + tt.queryPath
			}
			c, w := newFlowFileTestContext(http.MethodGet, target, nil, tt.privs, tt.uid, tt.flowID)

			svc.DownloadFlowFile(c)

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

func TestFlowFiles_PullFlowFiles_ReportsEachPathSyncedFromTheContainer(t *testing.T) {
	tests := []struct {
		name                       string
		flowOwner                  uint64
		uid                        uint64
		flowID                     uint64
		privs                      []string
		body                       any // marshaled via json
		rawBody                    string
		seedExisting               bool     // seeds container/etc/nginx.conf
		seedExistingContainerFiles []string // additional relative paths to seed under container cache
		dockerSetup                func(*fakeDockerClient)
		dockerNil                  bool
		wantStatus                 int
		wantResponsePath           string   // for single-file response
		wantResponsePaths          []string // for multi-file response (ordered)
		wantEventChannel           string   // "added" | "updated" (checked when set, single-event tests)
		wantEventCount             int      // > 0: assert total event count
		wantCopyFromCount          int      // checked when > 0 or wantNoCopy
		wantNoCopy                 bool
	}{
		{
			name:      "pull new file",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body:  models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBody = buildContainerTar([]tarTestEntry{
					{name: "nginx.conf", typeflag: tar.TypeReg, content: "nginx-config"},
				})
			},
			wantStatus:       http.StatusOK,
			wantResponsePath: "container/etc/nginx.conf",
			wantEventChannel: "added",
		},
		{
			name:      "pull existing without force conflicts",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:        []string{"flow_files.upload", "containers.view"},
			body:         models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			seedExisting: true,
			wantStatus:   http.StatusConflict,
		},
		{
			name:      "pull existing with force overwrites",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:        []string{"flow_files.upload", "containers.view"},
			body:         models.PullFlowFilesRequest{Path: "/etc/nginx.conf", Force: true},
			seedExisting: true,
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBody = buildContainerTar([]tarTestEntry{
					{name: "nginx.conf", typeflag: tar.TypeReg, content: "new"},
				})
			},
			wantStatus:       http.StatusOK,
			wantResponsePath: "container/etc/nginx.conf",
			wantEventChannel: "updated",
		},
		{
			name:      "a path the daemon does not have answers not found, as the listing door does",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body:  models.PullFlowFilesRequest{Path: "/etc/gone.conf"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromErr = fmt.Errorf(
					"Error response from daemon: lstat /etc/gone.conf: no such file or directory: %w",
					cerrdefs.ErrNotFound,
				)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "missing flow_files privilege",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view", "containers.view"},
			body:       models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "missing containers.view privilege",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload"},
			body:       models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "containers.admin privilege is sufficient",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.admin"},
			body:  models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBody = buildContainerTar([]tarTestEntry{
					{name: "nginx.conf", typeflag: tar.TypeReg, content: "data"},
				})
			},
			wantStatus:       http.StatusOK,
			wantResponsePath: "container/etc/nginx.conf",
			wantEventChannel: "added",
		},
		{
			name:      "malformed json returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload", "containers.view"},
			rawBody:    `{not json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "empty container path returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload", "containers.view"},
			body:       models.PullFlowFilesRequest{Path: "  "},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "invalid container path returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload", "containers.view"},
			body:       models.PullFlowFilesRequest{Path: "/"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "docker client not configured returns internal",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload", "containers.view"},
			body:       models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerNil:  true,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "container not running returns 400",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body:  models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = false
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "running check error returns internal",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body:  models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerSetup: func(d *fakeDockerClient) {
				d.runningErr = fmt.Errorf("docker daemon down")
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "copy from container error returns internal",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body:  models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromErr = fmt.Errorf("not found")
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "a symlink cannot be pulled and is a client error",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body:  models.PullFlowFilesRequest{Path: "/etc/mtab"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBody = buildContainerTar([]tarTestEntry{
					{name: "mtab", typeflag: tar.TypeSymlink, linkname: "/proc/self/mounts"},
				})
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "archive missing expected entry returns internal",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body:  models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBody = buildContainerTar([]tarTestEntry{
					{name: "other.conf", typeflag: tar.TypeReg, content: "x"},
				})
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			// /etc covers /etc/nginx.conf, so only /etc is copied
			name:      "parent path covers child path - single docker call, directory contents listed",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body: models.PullFlowFilesRequest{
				Paths: []string{"/etc", "/etc/nginx.conf"},
			},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBodyMap = map[string][]byte{
					"/etc": buildContainerTar([]tarTestEntry{
						{name: "etc/", typeflag: tar.TypeDir},
						{name: "etc/nginx.conf", typeflag: tar.TypeReg, content: "nginx"},
					}),
				}
			},
			wantStatus: http.StatusOK,
			wantResponsePaths: []string{
				"container/etc",
				"container/etc/nginx.conf",
			},
			wantCopyFromCount: 1, // /etc/nginx.conf is covered → no separate pull
		},
		{
			name:      "two files via paths array pulled and returned",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body: models.PullFlowFilesRequest{
				Paths: []string{"/etc/nginx.conf", "/etc/hosts"},
			},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBodyMap = map[string][]byte{
					"/etc/nginx.conf": buildContainerTar([]tarTestEntry{
						{name: "nginx.conf", typeflag: tar.TypeReg, content: "nginx"},
					}),
					"/etc/hosts": buildContainerTar([]tarTestEntry{
						{name: "hosts", typeflag: tar.TypeReg, content: "hosts"},
					}),
				}
			},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"container/etc/hosts", "container/etc/nginx.conf"},
			wantEventCount:    2,
			wantCopyFromCount: 2,
		},
		{
			name:      "path and paths combined pull both files",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body: models.PullFlowFilesRequest{
				Path:  "/etc/nginx.conf",
				Paths: []string{"/etc/hosts"},
			},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBodyMap = map[string][]byte{
					"/etc/nginx.conf": buildContainerTar([]tarTestEntry{
						{name: "nginx.conf", typeflag: tar.TypeReg, content: "nginx"},
					}),
					"/etc/hosts": buildContainerTar([]tarTestEntry{
						{name: "hosts", typeflag: tar.TypeReg, content: "hosts"},
					}),
				}
			},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"container/etc/hosts", "container/etc/nginx.conf"},
			wantEventCount:    2,
		},
		{
			name:      "duplicate path in path and paths deduplicated to single operation",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body: models.PullFlowFilesRequest{
				Path:  "/etc/nginx.conf",
				Paths: []string{"/etc/nginx.conf"},
			},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBody = buildContainerTar([]tarTestEntry{
					{name: "nginx.conf", typeflag: tar.TypeReg, content: "nginx"},
				})
			},
			wantStatus:        http.StatusOK,
			wantResponsePath:  "container/etc/nginx.conf",
			wantCopyFromCount: 1,
		},
		{
			name:      "whitespace-only paths returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body: models.PullFlowFilesRequest{
				Path:  "   ",
				Paths: []string{"  ", "\t"},
			},
			dockerSetup: func(d *fakeDockerClient) { d.running = true },
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:      "second path already exists without force fails fast in phase 1",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body: models.PullFlowFilesRequest{
				Paths: []string{"/etc/nginx.conf", "/etc/hosts"},
			},
			seedExistingContainerFiles: []string{"etc/hosts"},
			dockerSetup:                func(d *fakeDockerClient) { d.running = true },
			wantStatus:                 http.StatusConflict,
			wantNoCopy:                 true,
		},
		{
			name:      "force overwrites multiple existing cache entries",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body: models.PullFlowFilesRequest{
				Paths: []string{"/etc/nginx.conf", "/etc/hosts"},
				Force: true,
			},
			seedExisting:               true, // seeds etc/nginx.conf
			seedExistingContainerFiles: []string{"etc/hosts"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBodyMap = map[string][]byte{
					"/etc/nginx.conf": buildContainerTar([]tarTestEntry{
						{name: "nginx.conf", typeflag: tar.TypeReg, content: "new-nginx"},
					}),
					"/etc/hosts": buildContainerTar([]tarTestEntry{
						{name: "hosts", typeflag: tar.TypeReg, content: "new-hosts"},
					}),
				}
			},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"container/etc/hosts", "container/etc/nginx.conf"},
			wantEventCount:    2,
		},
		{
			name:      "second container copy fails returns internal",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body: models.PullFlowFilesRequest{
				Paths: []string{"/etc/nginx.conf", "/etc/hosts"},
			},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBodyMap = map[string][]byte{
					"/etc/nginx.conf": buildContainerTar([]tarTestEntry{
						{name: "nginx.conf", typeflag: tar.TypeReg, content: "nginx"},
					}),
				}
				d.copyFromErrMap = map[string]error{
					"/etc/hosts": fmt.Errorf("path not found in container"),
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "response sorted by cache path regardless of input order",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body: models.PullFlowFilesRequest{
				Paths: []string{"/var/log/app.log", "/etc/nginx.conf"},
			},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBodyMap = map[string][]byte{
					"/var/log/app.log": buildContainerTar([]tarTestEntry{
						{name: "app.log", typeflag: tar.TypeReg, content: "log"},
					}),
					"/etc/nginx.conf": buildContainerTar([]tarTestEntry{
						{name: "nginx.conf", typeflag: tar.TypeReg, content: "nginx"},
					}),
				}
			},
			wantStatus: http.StatusOK,
			wantResponsePaths: []string{
				"container/etc/nginx.conf",
				"container/var/log/app.log",
			},
		},
		{
			name:      "added and updated events both published in batch",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body: models.PullFlowFilesRequest{
				Paths: []string{"/etc/nginx.conf", "/etc/hosts"},
				Force: true,
			},
			seedExistingContainerFiles: []string{"etc/hosts"}, // hosts already cached
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBodyMap = map[string][]byte{
					"/etc/nginx.conf": buildContainerTar([]tarTestEntry{
						{name: "nginx.conf", typeflag: tar.TypeReg, content: "nginx"},
					}),
					"/etc/hosts": buildContainerTar([]tarTestEntry{
						{name: "hosts", typeflag: tar.TypeReg, content: "new-hosts"},
					}),
				}
			},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"container/etc/hosts", "container/etc/nginx.conf"},
			wantEventCount:    2, // 1 "added" (nginx.conf) + 1 "updated" (hosts)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &flowFileCaptureSubscriptions{}

			var dockerClient docker.DockerClient
			fakeDocker := &fakeDockerClient{}
			if tt.dockerSetup != nil {
				tt.dockerSetup(fakeDocker)
			}
			if !tt.dockerNil {
				dockerClient = fakeDocker
			}
			svc := NewFlowFileService(db, dataDir, "", dockerClient, ss)

			seedFlow(t, db, tt.flowID, tt.flowOwner)

			if tt.seedExisting {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), "container", "etc")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "nginx.conf"), []byte("old"), 0644))
			}
			for _, relPath := range tt.seedExistingContainerFiles {
				abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), "container", filepath.FromSlash(relPath))
				require.NoError(t, os.MkdirAll(filepath.Dir(abs), 0755))
				require.NoError(t, os.WriteFile(abs, []byte("old"), 0644))
			}

			var bodyReader io.Reader
			if tt.rawBody != "" {
				bodyReader = bytes.NewBufferString(tt.rawBody)
			} else {
				payload, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyReader = bytes.NewBuffer(payload)
			}
			c, w := newFlowFileTestContext(http.MethodPost, "/flows/1/files/pull", bodyReader, tt.privs, tt.uid, tt.flowID)
			c.Request.Header.Set("Content-Type", "application/json")

			svc.PullFlowFiles(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantCopyFromCount > 0 || tt.wantNoCopy {
				assert.Equal(t, tt.wantCopyFromCount, fakeDocker.copyFromCount, "unexpected CopyFromContainer call count")
			}
			if tt.wantStatus != http.StatusOK {
				return
			}

			resp := decodeFlowFilesResponse(t, w)

			if len(tt.wantResponsePaths) > 0 {
				paths := make([]string, len(resp.Files))
				for i, f := range resp.Files {
					paths[i] = f.Path
				}
				assert.Equal(t, tt.wantResponsePaths, paths)
			} else if tt.wantResponsePath != "" {
				require.Len(t, resp.Files, 1)
				assert.Equal(t, tt.wantResponsePath, resp.Files[0].Path)
			}

			events := ss.snapshot()
			if tt.wantEventCount > 0 {
				assert.Len(t, events, tt.wantEventCount, "unexpected subscription event count")
			}
			if tt.wantEventChannel != "" {
				require.Len(t, events, 1)
				assert.Equal(t, "flow", events[0].channel)
				assert.Equal(t, tt.wantEventChannel, events[0].action)
				if tt.wantResponsePath != "" {
					assert.Equal(t, tt.wantResponsePath, events[0].path)
				}
			}
		})
	}
}

// flowFilesDaemonDetail carries what must not leave a release build: a container id
// and the daemon's address.
const flowFilesDaemonDetail = "Error response from daemon: container 9f8e7d6c5b4a on tcp://10.0.0.5:2376 lstat failed"

func flowFilesDaemonDetailDockerSetup(d *fakeDockerClient) {
	d.running = true
	d.statPathMap = map[string]container.PathStat{"/work": {Mode: os.ModeDir | 0755}}
	d.listDirMap = map[string][]container.PathStat{"/work": {{Name: "readme", Mode: 0644, Size: 1}}}
	d.listDirFailMap = map[string][]docker.ContainerEntryError{
		"/work": {{Name: "secret", Path: "/work/secret", Err: fmt.Errorf("%s", flowFilesDaemonDetail)}},
	}
}

// flowFilesReadAndFailedDockerSetup makes /work/x fail inside the /work listing
// while a stat of /work/x itself succeeds.
func flowFilesReadAndFailedDockerSetup(d *fakeDockerClient) {
	d.running = true
	d.statPathMap = map[string]container.PathStat{
		"/work":   {Mode: os.ModeDir | 0755},
		"/work/x": {Name: "x", Mode: 0644, Size: 1},
	}
	d.listDirMap = map[string][]container.PathStat{"/work": {}}
	d.listDirFailMap = map[string][]docker.ContainerEntryError{
		"/work": {{Name: "x", Path: "/work/x", Err: fmt.Errorf("stat: no such file")}},
	}
}

func TestFlowFiles_GetFlowContainerFiles_ListsReadableEntriesAndSurfacesFailures(t *testing.T) {
	tests := []struct {
		name             string
		flowOwner        uint64
		uid              uint64
		flowID           uint64
		privs            []string
		queryPath        string // builds ?path=<value>
		rawQuery         string // when set, used verbatim as query string (overrides queryPath)
		dockerSetup      func(*fakeDockerClient)
		dockerNil        bool
		wantStatus       int
		wantPathInResp   string
		wantFileNames    []string // expected file names in sorted order (nil = don't check)
		wantTotal        uint64   // > 0: verify response Total
		wantFailureNames []string // expected Failures[].Name (nil = don't check)
		wantFailureTexts []string // expected Failures[].Message (nil = don't check)
		wantTruncated    bool
		wantErrorCode    string // when set, the "code" of the error body
		build            string // "release" or "develop" sets version.PackageVer for the row; empty keeps it
	}{
		{
			name:      "default path lists work directory when no param given",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.view", "containers.view"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPath = container.PathStat{Mode: os.ModeDir | 0755}
				d.listDir = []container.PathStat{
					{Name: "uploads", Mode: os.ModeDir | 0755},
					{Name: "report.txt", Size: 7, Mode: 0644},
				}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "/work",
			wantFileNames:  []string{"report.txt", "uploads"},
		},
		{
			name:      "custom single path= is honoured",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:     []string{"flow_files.view", "containers.view"},
			queryPath: "/etc",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPath = container.PathStat{Mode: os.ModeDir | 0755}
				d.listDir = []container.PathStat{
					{Name: "passwd", Size: 1, Mode: 0644},
				}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "/etc",
			wantFileNames:  []string{"passwd"},
		},
		{
			name:      "regular file path returns single file entry",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:     []string{"flow_files.view", "containers.view"},
			queryPath: "/etc/passwd",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPath = container.PathStat{Name: "passwd", Size: 5, Mode: 0644}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "/etc/passwd",
			wantFileNames:  []string{"passwd"},
		},
		{
			name:      "missing flow_files privilege returns forbidden",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"containers.view"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "missing containers privilege returns forbidden",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "containers.admin is sufficient for container listing",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.view", "containers.admin"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPath = container.PathStat{Mode: os.ModeDir | 0755}
				d.listDir = []container.PathStat{{Name: "a.txt", Mode: 0644}}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "/work",
			wantFileNames:  []string{"a.txt"},
		},
		{
			name:       "missing flow returns not found",
			uid:        1,
			flowID:     99,
			privs:      []string{"flow_files.view", "containers.view"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "docker client not configured returns internal error",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view", "containers.view"},
			dockerNil:  true,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "container not running returns 400",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.view", "containers.view"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = false
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "stat path error returns internal error",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.view", "containers.view"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPathErr = fmt.Errorf("not found")
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "a path the daemon does not have answers not found, not a server fault",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:     []string{"flow_files.view", "containers.view"},
			queryPath: "/wrok",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPathErr = fmt.Errorf(
					"Error response from daemon: lstat /wrok: no such file or directory: %w",
					cerrdefs.ErrNotFound,
				)
			},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "FlowFiles.NotFound",
		},
		{
			name:      "a directory listing the daemon reports missing answers not found",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:     []string{"flow_files.view", "containers.view"},
			queryPath: "/work/gone",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPath = container.PathStat{Mode: os.ModeDir | 0755}
				d.listDirErr = fmt.Errorf("Error response from daemon: %w", cerrdefs.ErrNotFound)
			},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "FlowFiles.NotFound",
		},
		{
			name:      "list dir error returns internal error",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.view", "containers.view"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPath = container.PathStat{Mode: os.ModeDir | 0755}
				d.listDirErr = fmt.Errorf("permission denied")
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "single paths[] behaves like single path=",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view", "containers.view"},
			rawQuery: "paths[]=/etc",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPath = container.PathStat{Mode: os.ModeDir | 0755}
				d.listDir = []container.PathStat{
					{Name: "passwd", Size: 1, Mode: 0644},
				}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "/etc",
			wantFileNames:  []string{"passwd"},
		},
		{
			name:      "per-entry stat failure returns partial listing not 500",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:     []string{"flow_files.view", "containers.view"},
			queryPath: "/work",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPathMap = map[string]container.PathStat{
					"/work": {Mode: os.ModeDir | 0755},
				}
				d.listDirMap = map[string][]container.PathStat{
					"/work": {{Name: "good.txt", Size: 3, Mode: 0644}},
				}
				d.listDirFailMap = map[string][]docker.ContainerEntryError{
					"/work": {{
						Name: "dangling",
						Path: "/work/dangling",
						Err:  fmt.Errorf("no such file or directory"),
					}},
				}
			},
			wantStatus:       http.StatusOK,
			wantPathInResp:   "/work",
			wantFileNames:    []string{"good.txt"},
			wantFailureNames: []string{"dangling"},
		},
		{
			// several paths have no single current directory, so Path is empty
			name:      "two directories via paths[] returns combined sorted listing",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view", "containers.view"},
			rawQuery: "paths[]=/etc&paths[]=/var",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPathMap = map[string]container.PathStat{
					"/etc": {Mode: os.ModeDir | 0755},
					"/var": {Mode: os.ModeDir | 0755},
				}
				d.listDirMap = map[string][]container.PathStat{
					"/etc": {
						{Name: "nginx.conf", Size: 10, Mode: 0644},
						{Name: "passwd", Size: 1, Mode: 0644},
					},
					"/var": {
						{Name: "log", Mode: os.ModeDir | 0755},
					},
				}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "",
			wantFileNames:  []string{"nginx.conf", "passwd", "log"},
			wantTotal:      3,
		},
		{
			name:      "path= and paths[] combined returns merged sorted listing",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view", "containers.view"},
			rawQuery: "path=/etc&paths[]=/var",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPathMap = map[string]container.PathStat{
					"/etc": {Mode: os.ModeDir | 0755},
					"/var": {Mode: os.ModeDir | 0755},
				}
				d.listDirMap = map[string][]container.PathStat{
					"/etc": {{Name: "hosts", Size: 5, Mode: 0644}},
					"/var": {{Name: "log", Mode: os.ModeDir | 0755}},
				}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "",
			wantFileNames:  []string{"hosts", "log"},
		},
		{
			name:      "duplicate paths in paths[] deduplicated",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view", "containers.view"},
			rawQuery: "paths[]=/etc&paths[]=/etc",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPath = container.PathStat{Mode: os.ModeDir | 0755}
				d.listDir = []container.PathStat{
					{Name: "passwd", Size: 1, Mode: 0644},
				}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "/etc",
			wantFileNames:  []string{"passwd"},
			wantTotal:      1,
		},
		{
			name:      "whitespace-only paths[] values return bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.view", "containers.view"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
			},
			rawQuery:   "paths[]=%20%20&paths[]=%09",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "mix of file and directory in paths[] combined in response",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view", "containers.view"},
			rawQuery: "paths[]=/etc/passwd&paths[]=/var",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPathMap = map[string]container.PathStat{
					"/etc/passwd": {Name: "passwd", Size: 5, Mode: 0644},
					"/var":        {Mode: os.ModeDir | 0755},
				}
				d.listDirMap = map[string][]container.PathStat{
					"/var": {{Name: "log", Mode: os.ModeDir | 0755}},
				}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "",
			wantFileNames:  []string{"passwd", "log"},
		},
		{
			name:      "output deduplicated when Docker API returns duplicate entries",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view", "containers.view"},
			rawQuery: "paths[]=/work",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPath = container.PathStat{Mode: os.ModeDir | 0755}
				d.listDir = []container.PathStat{
					{Name: "file.txt", Size: 5, Mode: 0644},
					{Name: "file.txt", Size: 5, Mode: 0644}, // duplicate
				}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "/work",
			wantFileNames:  []string{"file.txt"},
			wantTotal:      1,
		},
		{
			name:      "second path stat error in batch: good path served, bad path a failure",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view", "containers.view"},
			rawQuery: "paths[]=/work&paths[]=/etc",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPathMap = map[string]container.PathStat{
					"/work": {Mode: os.ModeDir | 0755},
				}
				d.statPathErrMap = map[string]error{
					"/etc": fmt.Errorf("path not found"),
				}
				d.listDirMap = map[string][]container.PathStat{
					"/work": {{Name: "uploads", Mode: os.ModeDir | 0755}},
				}
			},
			wantStatus:       http.StatusOK,
			wantFileNames:    []string{"uploads"},
			wantFailureNames: []string{"etc"},
		},
		{
			name:      "second path list dir error in batch: good path served, bad path a failure",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view", "containers.view"},
			rawQuery: "paths[]=/work&paths[]=/etc",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPathMap = map[string]container.PathStat{
					"/work": {Mode: os.ModeDir | 0755},
					"/etc":  {Mode: os.ModeDir | 0755},
				}
				d.listDirMap = map[string][]container.PathStat{
					"/work": {{Name: "uploads", Mode: os.ModeDir | 0755}},
				}
				d.listDirErrMap = map[string]error{
					"/etc": fmt.Errorf("permission denied"),
				}
			},
			wantStatus:       http.StatusOK,
			wantFileNames:    []string{"uploads"},
			wantFailureNames: []string{"etc"},
		},
		{
			name:      "a path read by one query is never also a failure of another",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:            []string{"flow_files.view", "containers.view"},
			rawQuery:         "paths[]=/work&paths[]=/work/x",
			dockerSetup:      flowFilesReadAndFailedDockerSetup,
			wantStatus:       http.StatusOK,
			wantFileNames:    []string{"x"},
			wantFailureNames: []string{},
		},
		{
			name:      "a path read by one query is never also a failure of an earlier one",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:            []string{"flow_files.view", "containers.view"},
			rawQuery:         "paths[]=/work/x&paths[]=/work",
			dockerSetup:      flowFilesReadAndFailedDockerSetup,
			wantStatus:       http.StatusOK,
			wantFileNames:    []string{"x"},
			wantFailureNames: []string{},
		},
		{
			name:      "every entry failing still answers 200 with the failures",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view", "containers.view"},
			rawQuery: "paths[]=/work",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPathMap = map[string]container.PathStat{"/work": {Mode: os.ModeDir | 0755}}
				d.listDirMap = map[string][]container.PathStat{"/work": {}}
				d.listDirFailMap = map[string][]docker.ContainerEntryError{
					"/work": {
						{Name: "a", Path: "/work/a", Err: fmt.Errorf("stat: gone")},
						{Name: "b", Path: "/work/b", Err: fmt.Errorf("stat: gone")},
					},
				}
			},
			wantStatus:       http.StatusOK,
			wantPathInResp:   "/work",
			wantFileNames:    []string{},
			wantFailureNames: []string{"a", "b"},
		},
		{
			name:      "a truncated listing says so",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:    []string{"flow_files.view", "containers.view"},
			rawQuery: "paths[]=/big",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPathMap = map[string]container.PathStat{"/big": {Mode: os.ModeDir | 0755}}
				d.listDirMap = map[string][]container.PathStat{"/big": {{Name: "f", Mode: 0644, Size: 1}}}
				d.listDirTruncated = map[string]bool{"/big": true}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "/big",
			wantFileNames:  []string{"f"},
			wantTruncated:  true,
		},
		{
			name:      "a release build hides the daemon's error text",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:            []string{"flow_files.view", "containers.view"},
			rawQuery:         "paths[]=/work",
			dockerSetup:      flowFilesDaemonDetailDockerSetup,
			build:            "release",
			wantStatus:       http.StatusOK,
			wantPathInResp:   "/work",
			wantFileNames:    []string{"readme"},
			wantFailureTexts: []string{"entry could not be read"},
		},
		{
			name:      "develop mode shows the daemon's error text",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:            []string{"flow_files.view", "containers.view"},
			rawQuery:         "paths[]=/work",
			dockerSetup:      flowFilesDaemonDetailDockerSetup,
			build:            "develop",
			wantStatus:       http.StatusOK,
			wantPathInResp:   "/work",
			wantFileNames:    []string{"readme"},
			wantFailureTexts: []string{flowFilesDaemonDetail},
		},
		{
			name:      "more paths than the cap are refused",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"flow_files.view", "containers.view"},
			rawQuery:    flowFilesContainerPathsQuery(maxContainerListPaths + 1),
			dockerSetup: func(d *fakeDockerClient) { d.running = true },
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:      "exactly the cap of paths is served",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"flow_files.view", "containers.view"},
			rawQuery:    flowFilesContainerPathsQuery(maxContainerListPaths),
			dockerSetup: func(d *fakeDockerClient) { d.running = true },
			wantStatus:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.build != "" {
				previous := version.PackageVer
				t.Cleanup(func() { version.PackageVer = previous })
				version.PackageVer = ""
				if tt.build == "release" {
					version.PackageVer = "1.0.0"
				}
			}

			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()

			var dockerClient docker.DockerClient
			fakeDocker := &fakeDockerClient{}
			if tt.dockerSetup != nil {
				tt.dockerSetup(fakeDocker)
			}
			if !tt.dockerNil {
				dockerClient = fakeDocker
			}
			svc := NewFlowFileService(db, dataDir, "", dockerClient, nil)

			if tt.flowOwner != 0 {
				seedFlow(t, db, tt.flowID, tt.flowOwner)
			}

			target := "/flows/1/files/container"
			switch {
			case tt.rawQuery != "":
				target += "?" + tt.rawQuery
			case tt.queryPath != "":
				target += "?path=" + tt.queryPath
			}
			c, w := newFlowFileTestContext(http.MethodGet, target, nil, tt.privs, tt.uid, tt.flowID)

			svc.GetFlowContainerFiles(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus != http.StatusOK {
				if tt.wantErrorCode != "" {
					var body struct {
						Code string `json:"code"`
					}
					require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
					assert.Equal(t, tt.wantErrorCode, body.Code)
				}
				return
			}
			resp := decodeContainerFilesResponse(t, w)
			assert.Equal(t, tt.wantPathInResp, resp.Path)
			assert.Equal(t, tt.wantTruncated, resp.Truncated)

			if tt.wantFileNames != nil {
				names := make([]string, len(resp.Files))
				for i, f := range resp.Files {
					names[i] = f.Name
				}
				assert.Equal(t, tt.wantFileNames, names)
			}
			if tt.wantTotal > 0 {
				assert.Equal(t, tt.wantTotal, resp.Total, "response Total mismatch")
				assert.Len(t, resp.Files, int(tt.wantTotal), "response Files length mismatch")
			}
			if tt.wantFailureNames != nil {
				names := make([]string, len(resp.Failures))
				for i, f := range resp.Failures {
					names[i] = f.Name
				}
				assert.ElementsMatch(t, tt.wantFailureNames, names)
			}
			if tt.wantFailureTexts != nil {
				texts := make([]string, len(resp.Failures))
				for i, f := range resp.Failures {
					texts[i] = f.Message
				}
				assert.Equal(t, tt.wantFailureTexts, texts)
			}
		})
	}
}

func TestFlowFiles_AddResourcesToFlow_CopiesPermittedResourcesIntoTheFlow(t *testing.T) {
	type seedRes struct {
		userID  uint64
		path    string
		content string
	}

	tests := []struct {
		name             string
		flowOwner        uint64
		uid              uint64
		flowID           uint64
		privs            []string
		seedResources    []seedRes
		body             any
		rawBody          string
		seedExisting     map[string]string // relative path under flow resources/ -> content
		wantStatus       int
		wantFileExists   string // relative to flow resources/
		wantContent      string
		wantEventChannel string // "added" or "updated"
	}{
		{
			name:      "copy single resource",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:         []string{"flow_files.upload", "resources.view"},
			seedResources: []seedRes{{path: "creds/p.txt", content: "secret"}},
			body: map[string]any{
				"ids": []uint64{1},
			},
			wantStatus:       http.StatusOK,
			wantFileExists:   "creds/p.txt",
			wantContent:      "secret",
			wantEventChannel: "added",
		},
		{
			name:      "skip existing resource without force",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:         []string{"flow_files.upload", "resources.view"},
			seedResources: []seedRes{{path: "creds/p.txt", content: "secret"}},
			seedExisting:  map[string]string{"creds/p.txt": "old"},
			body: map[string]any{
				"ids": []uint64{1},
			},
			wantStatus:     http.StatusOK,
			wantFileExists: "creds/p.txt",
			wantContent:    "old",
		},
		{
			name:      "force overwrite emits updated event",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:         []string{"flow_files.upload", "resources.view"},
			seedResources: []seedRes{{path: "creds/p.txt", content: "new"}},
			seedExisting:  map[string]string{"creds/p.txt": "old"},
			body: map[string]any{
				"ids":   []uint64{1},
				"force": true,
			},
			wantStatus:       http.StatusOK,
			wantFileExists:   "creds/p.txt",
			wantContent:      "new",
			wantEventChannel: "updated",
		},
		{
			name:      "admin can copy other user's resource",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:         []string{"flow_files.admin", "resources.admin"},
			seedResources: []seedRes{{userID: 2, path: "creds/p.txt", content: "secret"}},
			body: map[string]any{
				"ids": []uint64{1},
			},
			wantStatus:       http.StatusOK,
			wantFileExists:   "creds/p.txt",
			wantContent:      "secret",
			wantEventChannel: "added",
		},
		{
			name:      "non-admin cannot copy another user's resource",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:         []string{"flow_files.upload", "resources.view"},
			seedResources: []seedRes{{userID: 2, path: "creds/p.txt", content: "secret"}},
			body: map[string]any{
				"ids": []uint64{1},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "missing flow_files.upload privilege",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.view"},
			body: map[string]any{
				"ids": []uint64{1},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "missing resources.view privilege",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload"},
			body: map[string]any{
				"ids": []uint64{1},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "malformed json returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload", "resources.view"},
			rawBody:    `{not json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "empty ids returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "resources.view"},
			body: map[string]any{
				"ids": []uint64{},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "non-existent resource id returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "resources.view"},
			body: map[string]any{
				"ids": []uint64{999},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing flow returns not found",
			uid:        1,
			flowID:     99,
			privs:      []string{"flow_files.upload", "resources.view"},
			body:       map[string]any{"ids": []uint64{1}},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &flowFileCaptureSubscriptions{}
			svc := NewFlowFileService(db, dataDir, "", nil, ss)

			require.NoError(t, resources.EnsureResourcesDir(dataDir))
			if tt.flowOwner != 0 {
				seedFlow(t, db, tt.flowID, tt.flowOwner)
			}
			for _, r := range tt.seedResources {
				userID := r.userID
				if userID == 0 {
					userID = tt.uid
				}
				hash := md5HexForService(r.content)
				writeResourceBlob(t, dataDir, hash, r.content)
				seedResource(t, db, models.UserResource{
					UserID: userID,
					Hash:   hash,
					Name:   filepath.Base(r.path),
					Path:   r.path,
					Size:   int64(len(r.content)),
				})
			}
			for relPath, content := range tt.seedExisting {
				abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), "resources", filepath.FromSlash(relPath))
				require.NoError(t, os.MkdirAll(filepath.Dir(abs), 0755))
				require.NoError(t, os.WriteFile(abs, []byte(content), 0644))
			}

			var bodyReader io.Reader
			if tt.rawBody != "" {
				bodyReader = bytes.NewBufferString(tt.rawBody)
			} else {
				payload, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyReader = bytes.NewBuffer(payload)
			}
			c, w := newFlowFileTestContext(http.MethodPost, "/flows/1/files/resources", bodyReader, tt.privs, tt.uid, tt.flowID)
			c.Request.Header.Set("Content-Type", "application/json")

			svc.AddResourcesToFlow(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus != http.StatusOK {
				return
			}
			abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), "resources", filepath.FromSlash(tt.wantFileExists))
			data, err := os.ReadFile(abs)
			require.NoError(t, err)
			assert.Equal(t, tt.wantContent, string(data))

			if tt.wantEventChannel != "" {
				matched := false
				for _, ev := range ss.snapshot() {
					if ev.channel == "flow" && ev.action == tt.wantEventChannel {
						matched = true
						break
					}
				}
				assert.True(t, matched, "expected event %q not emitted", tt.wantEventChannel)
			} else {
				for _, ev := range ss.snapshot() {
					assert.NotEqual(t, "flow", ev.channel, "no flow events expected for skipped copy")
				}
			}
		})
	}
}

func TestFlowFiles_AddResourceFromFlow_PromotesCacheEntriesToResources(t *testing.T) {
	type sourceFile struct {
		relPath string
		content string
	}

	tests := []struct {
		name              string
		flowOwner         uint64
		uid               uint64
		flowID            uint64
		privs             []string
		sourceFiles       []sourceFile
		sourceDirs        []string // extra directories to create (without files)
		existingResource  *models.UserResource
		body              any
		rawBody           string
		wantStatus        int
		wantResourcePaths []string // every path expected to be present in the ResourceList response and DB
		wantEventChannel  string   // "added" | "updated"
	}{
		{
			name:      "force does not replace a file on the way to the destination",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "payload"}},
			existingResource: &models.UserResource{
				Hash: md5HexForService("irreplaceable"), Name: "victim.txt", Path: "victim.txt", Size: 13,
			},
			body: map[string]any{
				"source":      "uploads/report.txt",
				"destination": "victim.txt/report.txt",
				"force":       true,
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:      "force does not replace a file above a multi-source destination",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{
				{relPath: "uploads/a.txt", content: "a"},
				{relPath: "uploads/b.txt", content: "b"},
			},
			existingResource: &models.UserResource{
				Hash: md5HexForService("irreplaceable"), Name: "victim.txt", Path: "victim.txt", Size: 13,
			},
			body: map[string]any{
				"sources":     []string{"uploads/a.txt", "uploads/b.txt"},
				"destination": "victim.txt/batch",
				"force":       true,
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:      "promote uploaded file to new resource",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "payload"}},
			body: map[string]any{
				"source":      "uploads/report.txt",
				"destination": "promoted/report.txt",
			},
			wantStatus:        http.StatusOK,
			wantResourcePaths: []string{"promoted/report.txt"},
			wantEventChannel:  "added",
		},
		{
			name:      "a traversal source refuses the request instead of narrowing it",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "payload"}},
			body: map[string]any{
				"sources":     []string{"uploads/report.txt", "/etc/passwd"},
				"destination": "promoted",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "promote container file to user resources",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{{relPath: "container/etc/nginx.conf", content: "nginx"}},
			body: map[string]any{
				"source":      "container/etc/nginx.conf",
				"destination": "configs/nginx.conf",
			},
			wantStatus:        http.StatusOK,
			wantResourcePaths: []string{"configs/nginx.conf"},
			wantEventChannel:  "added",
		},
		{
			name:      "force overwrite existing resource",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "new"}},
			existingResource: &models.UserResource{
				Hash: md5HexForService("old"), Name: "report.txt", Path: "promoted/report.txt", Size: 3,
			},
			body: map[string]any{
				"source":      "uploads/report.txt",
				"destination": "promoted/report.txt",
				"force":       true,
			},
			wantStatus:        http.StatusOK,
			wantResourcePaths: []string{"promoted/report.txt"},
			wantEventChannel:  "updated",
		},
		{
			name:      "existing resource without force returns conflict",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "new"}},
			existingResource: &models.UserResource{
				Hash: md5HexForService("old"), Name: "report.txt", Path: "promoted/report.txt", Size: 3,
			},
			body: map[string]any{
				"source":      "uploads/report.txt",
				"destination": "promoted/report.txt",
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:      "resources admin still requires flow_files privilege to read flow",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.admin"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "payload"}},
			body: map[string]any{
				"source":      "uploads/report.txt",
				"destination": "promoted/report.txt",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "resources admin with flow_files.admin can promote any flow file",
			flowOwner: 2, uid: 1, flowID: 1,
			privs:       []string{"resources.admin", "flow_files.admin"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "payload"}},
			body: map[string]any{
				"source":      "uploads/report.txt",
				"destination": "promoted/report.txt",
			},
			wantStatus:        http.StatusOK,
			wantResourcePaths: []string{"promoted/report.txt"},
			wantEventChannel:  "added",
		},
		{
			name:      "missing resources.upload returns forbidden",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.view"},
			body: map[string]any{
				"source": "uploads/report.txt", "destination": "promoted/report.txt",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "missing flow_files.view returns forbidden",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload"},
			body: map[string]any{
				"source": "uploads/report.txt", "destination": "promoted/report.txt",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "missing flow returns not found",
			uid:  1, flowID: 99,
			privs: []string{"resources.upload", "flow_files.view"},
			body: map[string]any{
				"source": "uploads/report.txt", "destination": "promoted/report.txt",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "malformed json returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"resources.upload", "flow_files.view"},
			rawBody:    `{not json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "missing required fields returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"resources.upload", "flow_files.view"},
			body:       map[string]any{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "invalid source path returns bad data",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			body: map[string]any{
				"source": "tmp/evil.txt", "destination": "promoted/report.txt",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "missing source file returns not found",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			body: map[string]any{
				"source": "uploads/missing.txt", "destination": "promoted/report.txt",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "invalid destination path returns bad data",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "x"}},
			body: map[string]any{
				"source": "uploads/report.txt", "destination": "../escape.txt",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "promote container directory with nested files to resources",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{
				{relPath: "container/scan/hosts.txt", content: "10.0.0.1"},
				{relPath: "container/scan/sub/ports.txt", content: "80,443"},
			},
			body: map[string]any{
				"source":      "container/scan",
				"destination": "promoted/scan",
			},
			wantStatus: http.StatusOK,
			wantResourcePaths: []string{
				"promoted/scan",
				"promoted/scan/hosts.txt",
				"promoted/scan/sub",
				"promoted/scan/sub/ports.txt",
			},
			wantEventChannel: "added",
		},
		{
			name:      "promote uploads directory to resources",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{
				{relPath: "uploads/payloads/shell.sh", content: "#!/bin/sh"},
				{relPath: "uploads/payloads/enum.sh", content: "#!/bin/sh\nls"},
			},
			body: map[string]any{
				"source":      "uploads/payloads",
				"destination": "scripts/payloads",
			},
			wantStatus: http.StatusOK,
			wantResourcePaths: []string{
				"scripts/payloads",
				"scripts/payloads/shell.sh",
				"scripts/payloads/enum.sh",
			},
			wantEventChannel: "added",
		},
		{
			name:      "directory promotion without force returns conflict when file exists",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{
				{relPath: "uploads/data/report.txt", content: "new content"},
			},
			existingResource: &models.UserResource{
				Hash: md5HexForService("old content"), Name: "report.txt",
				Path: "promoted/data/report.txt", Size: 11,
			},
			body: map[string]any{
				"source":      "uploads/data",
				"destination": "promoted/data",
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:      "directory promotion with force overwrites existing file",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{
				{relPath: "container/results/output.txt", content: "updated"},
			},
			existingResource: &models.UserResource{
				Hash: md5HexForService("original"), Name: "output.txt",
				Path: "archive/results/output.txt", Size: 8,
			},
			body: map[string]any{
				"source":      "container/results",
				"destination": "archive/results",
				"force":       true,
			},
			wantStatus: http.StatusOK,
			wantResourcePaths: []string{
				"archive/results",
				"archive/results/output.txt",
			},
			wantEventChannel: "updated",
		},
		{
			name:      "single file at deep path creates parent directory records",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{{relPath: "uploads/note.txt", content: "deep"}},
			body: map[string]any{
				"source":      "uploads/note.txt",
				"destination": "deep/nested/folder/note.txt",
			},
			wantStatus: http.StatusOK,
			wantResourcePaths: []string{
				"deep",
				"deep/nested",
				"deep/nested/folder",
				"deep/nested/folder/note.txt",
			},
			wantEventChannel: "added",
		},
		{
			name:      "empty source directory creates only the destination dir entry",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"resources.upload", "flow_files.view"},
			sourceDirs: []string{"uploads/empty"},
			body: map[string]any{
				"source":      "uploads/empty",
				"destination": "promoted/empty",
			},
			wantStatus: http.StatusOK,
			wantResourcePaths: []string{
				"promoted/empty",
			},
			wantEventChannel: "added",
		},
		{
			name:      "multi-source: two files promoted to a common base directory",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{
				{relPath: "uploads/hosts.txt", content: "10.0.0.1"},
				{relPath: "container/ports.txt", content: "80,443"},
			},
			body: map[string]any{
				"sources":     []string{"uploads/hosts.txt", "container/ports.txt"},
				"destination": "scan-results",
			},
			wantStatus: http.StatusOK,
			wantResourcePaths: []string{
				"scan-results/hosts.txt",
				"scan-results/ports.txt",
			},
			wantEventChannel: "added",
		},
		{
			name:      "multi-source: source and sources are merged and deduplicated",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{
				{relPath: "uploads/a.txt", content: "aaa"},
				{relPath: "uploads/b.txt", content: "bbb"},
			},
			body: map[string]any{
				"source":      "uploads/a.txt",
				"sources":     []string{"uploads/a.txt", "uploads/b.txt"},
				"destination": "merged",
			},
			wantStatus: http.StatusOK,
			wantResourcePaths: []string{
				"merged/a.txt",
				"merged/b.txt",
			},
			wantEventChannel: "added",
		},
		{
			name:      "multi-source: two directories promoted to a common base",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{
				{relPath: "container/web/index.html", content: "<html/>"},
				{relPath: "container/web/style.css", content: "body{}"},
				{relPath: "uploads/docs/readme.md", content: "# doc"},
			},
			body: map[string]any{
				"sources":     []string{"container/web", "uploads/docs"},
				"destination": "artifacts",
			},
			wantStatus: http.StatusOK,
			wantResourcePaths: []string{
				"artifacts/web",
				"artifacts/web/index.html",
				"artifacts/web/style.css",
				"artifacts/docs",
				"artifacts/docs/readme.md",
			},
			wantEventChannel: "added",
		},
		{
			name:      "multi-source: mixed file and directory",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{
				{relPath: "uploads/report.txt", content: "report"},
				{relPath: "container/scan/out.txt", content: "scan output"},
			},
			body: map[string]any{
				"sources":     []string{"uploads/report.txt", "container/scan"},
				"destination": "bundle",
			},
			wantStatus: http.StatusOK,
			wantResourcePaths: []string{
				"bundle/report.txt",
				"bundle/scan",
				"bundle/scan/out.txt",
			},
			wantEventChannel: "added",
		},
		{
			name:      "multi-source: force overwrites existing resource",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{
				{relPath: "uploads/a.txt", content: "new-a"},
				{relPath: "uploads/b.txt", content: "new-b"},
			},
			existingResource: &models.UserResource{
				Hash: md5HexForService("old-a"), Name: "a.txt", Path: "out/a.txt", Size: 5,
			},
			body: map[string]any{
				"sources":     []string{"uploads/a.txt", "uploads/b.txt"},
				"destination": "out",
				"force":       true,
			},
			wantStatus: http.StatusOK,
			wantResourcePaths: []string{
				"out/a.txt",
				"out/b.txt",
			},
			wantEventChannel: "added",
		},
		{
			name:      "multi-source: without force returns conflict when destination exists",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{
				{relPath: "uploads/a.txt", content: "new"},
				{relPath: "uploads/extra.txt", content: "extra"}, // second source makes multiSource=true
			},
			existingResource: &models.UserResource{
				Hash: md5HexForService("old"), Name: "a.txt", Path: "out/a.txt", Size: 3,
			},
			body: map[string]any{
				"sources":     []string{"uploads/a.txt", "uploads/extra.txt"},
				"destination": "out",
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:      "multi-source: empty sources returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			body: map[string]any{
				"sources":     []string{},
				"destination": "out",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "multi-source: sources with blank entries deduplicated to none returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			body: map[string]any{
				"sources":     []string{"   ", ""},
				"destination": "out",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "multi-source: one missing source returns not found",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{
				{relPath: "uploads/good.txt", content: "ok"},
			},
			body: map[string]any{
				"sources":     []string{"uploads/good.txt", "uploads/missing.txt"},
				"destination": "out",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "multi-source: empty directory in sources creates dir entry",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"resources.upload", "flow_files.view"},
			sourceDirs: []string{"uploads/empty-dir"},
			sourceFiles: []sourceFile{
				{relPath: "uploads/note.txt", content: "hi"},
			},
			body: map[string]any{
				"sources":     []string{"uploads/empty-dir", "uploads/note.txt"},
				"destination": "multi-out",
			},
			wantStatus: http.StatusOK,
			wantResourcePaths: []string{
				"multi-out/empty-dir",
				"multi-out/note.txt",
			},
			wantEventChannel: "added",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &flowFileCaptureSubscriptions{}
			svc := NewFlowFileService(db, dataDir, "", nil, ss)

			require.NoError(t, resources.EnsureResourcesDir(dataDir))
			if tt.flowOwner != 0 {
				seedFlow(t, db, tt.flowID, tt.flowOwner)
			}
			for _, sf := range tt.sourceFiles {
				abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), filepath.FromSlash(sf.relPath))
				require.NoError(t, os.MkdirAll(filepath.Dir(abs), 0755))
				require.NoError(t, os.WriteFile(abs, []byte(sf.content), 0644))
			}
			for _, sd := range tt.sourceDirs {
				abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), filepath.FromSlash(sd))
				require.NoError(t, os.MkdirAll(abs, 0755))
			}
			if tt.existingResource != nil {
				existing := *tt.existingResource
				existing.UserID = tt.uid
				if existing.Hash != "" {
					writeResourceBlob(t, dataDir, existing.Hash, "old")
				}
				seedResource(t, db, existing)
			}

			var bodyReader io.Reader
			if tt.rawBody != "" {
				bodyReader = bytes.NewBufferString(tt.rawBody)
			} else {
				payload, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyReader = bytes.NewBuffer(payload)
			}
			c, w := newFlowFileTestContext(http.MethodPost, "/flows/1/files/to-resources", bodyReader, tt.privs, tt.uid, tt.flowID)
			c.Request.Header.Set("Content-Type", "application/json")

			svc.AddResourceFromFlow(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus != http.StatusOK {
				var rows []models.UserResource
				require.NoError(t, db.Find(&rows).Error)
				assert.Empty(t, ss.snapshot())
				if tt.existingResource == nil {
					assert.Empty(t, rows, "a refused promotion must write nothing")
					return
				}
				require.Len(t, rows, 1, "a refused promotion must write nothing")
				assert.Equal(t, tt.existingResource.Path, rows[0].Path)
				assert.Equal(t, tt.existingResource.Hash, rows[0].Hash)
				assert.False(t, rows[0].IsDir)
				return
			}

			list := decodeResourceListResponse(t, w)
			assert.GreaterOrEqual(t, list.Total, uint64(len(tt.wantResourcePaths)))
			itemsByPath := make(map[string]models.ResourceEntry, len(list.Items))
			for _, item := range list.Items {
				itemsByPath[item.Path] = item
			}
			for _, wantPath := range tt.wantResourcePaths {
				_, inResponse := itemsByPath[wantPath]
				assert.True(t, inResponse, "expected resource at path %q in response", wantPath)

				var rec models.UserResource
				require.NoError(t, db.Where("user_id = ? AND path = ?", tt.uid, wantPath).First(&rec).Error,
					"resource record missing for path %q", wantPath)
			}

			if tt.wantEventChannel != "" {
				events := ss.snapshot()
				matched := false
				for _, ev := range events {
					if ev.channel == "resource" && ev.action == tt.wantEventChannel {
						matched = true
						break
					}
				}
				assert.True(t, matched, "expected resource event %q not emitted", tt.wantEventChannel)
			}
		})
	}
}

func decodeContainerFilesResponse(t *testing.T, w *httptest.ResponseRecorder) models.ContainerFiles {
	t.Helper()

	var resp struct {
		Status string                `json:"status"`
		Data   models.ContainerFiles `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "success", resp.Status)
	return resp.Data
}

type flowFileCaptureSubscriptions struct {
	mu     sync.Mutex
	events []flowFileEvent
}

func (s *flowFileCaptureSubscriptions) NewFlowSubscriber(int64, int64) subscriptions.FlowSubscriber {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewFlowPublisher(int64, int64) subscriptions.FlowPublisher {
	return &captureFlowPublisher{events: s}
}
func (s *flowFileCaptureSubscriptions) NewResourceSubscriber(int64) subscriptions.ResourceSubscriber {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewResourcePublisher(int64) subscriptions.ResourcePublisher {
	return &captureResourcePublisherForFlow{events: s}
}
func (s *flowFileCaptureSubscriptions) NewProviderSubscriber(int64) subscriptions.ProviderSubscriber {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewProviderPublisher(int64) subscriptions.ProviderPublisher {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewAPITokenSubscriber(int64) subscriptions.APITokenSubscriber {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewAPITokenPublisher(int64) subscriptions.APITokenPublisher {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewSettingsSubscriber(int64) subscriptions.SettingsSubscriber {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewSettingsPublisher(int64) subscriptions.SettingsPublisher {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewFlowTemplateSubscriber(int64) subscriptions.FlowTemplateSubscriber {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewFlowTemplatePublisher(int64) subscriptions.FlowTemplatePublisher {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewKnowledgeSubscriber(int64) subscriptions.KnowledgeSubscriber {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewKnowledgePublisher(int64) subscriptions.KnowledgePublisher {
	return nil
}

func (s *flowFileCaptureSubscriptions) record(e flowFileEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
}

func (s *flowFileCaptureSubscriptions) snapshot() []flowFileEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]flowFileEvent, len(s.events))
	copy(out, s.events)
	return out
}

type fakeDockerClient struct {
	mu sync.Mutex

	running    bool
	runningErr error

	statPath    container.PathStat
	statPathErr error

	listDir    []container.PathStat
	listDirErr error

	// A path with an entry in a *Map field is answered from it, not from the single-value field.
	statPathMap      map[string]container.PathStat
	statPathErrMap   map[string]error
	listDirMap       map[string][]container.PathStat
	listDirFailMap   map[string][]docker.ContainerEntryError
	listDirErrMap    map[string]error
	listDirTruncated map[string]bool

	copyFromBody    []byte
	copyFromStat    container.PathStat
	copyFromErr     error
	copyFromBodyMap map[string][]byte
	copyFromErrMap  map[string]error
	copyFromCount   int

	copyToErr   error
	copyToCalls []copyToCall

	execCreateID    string
	execCreateErr   error
	execAttachOut   string
	execAttachErr   error
	execInspectCode int
	execInspectErr  error
	execCommands    []string
}

func (f *fakeDockerClient) RunContainer(_ context.Context, _ string, _ database.ContainerType,
	_ int64, _ *container.Config, _ *container.HostConfig) (database.Container, error) {
	return database.Container{}, nil
}
func (f *fakeDockerClient) StopContainer(_ context.Context, _ string, _ int64) error   { return nil }
func (f *fakeDockerClient) RemoveContainer(_ context.Context, _ string, _ int64) error { return nil }
func (f *fakeDockerClient) IsContainerRunning(_ context.Context, _ string) (bool, error) {
	return f.running, f.runningErr
}
func (f *fakeDockerClient) KillFlowCommands(_ context.Context, _ string) error {
	return nil
}
func (f *fakeDockerClient) ContainerExecCreate(_ context.Context, _ string, opts client.ExecCreateOptions) (client.ExecCreateResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(opts.Cmd) > 0 {
		f.execCommands = append(f.execCommands, strings.Join(opts.Cmd, " "))
	}
	if f.execCreateErr != nil {
		return client.ExecCreateResult{}, f.execCreateErr
	}
	id := f.execCreateID
	if id == "" {
		id = "exec-id"
	}
	return client.ExecCreateResult{ID: id}, nil
}
func (f *fakeDockerClient) ContainerExecAttach(_ context.Context, _ string, _ client.ExecAttachOptions) (client.HijackedResponse, error) {
	if f.execAttachErr != nil {
		return client.HijackedResponse{}, f.execAttachErr
	}
	pr, pw := net.Pipe()
	go func() {
		_, _ = pw.Write([]byte(f.execAttachOut))
		_ = pw.Close()
	}()
	return client.HijackedResponse{Conn: pr, Reader: bufio.NewReader(pr)}, nil
}
func (f *fakeDockerClient) ContainerExecInspect(_ context.Context, _ string) (client.ExecInspectResult, error) {
	if f.execInspectErr != nil {
		return client.ExecInspectResult{}, f.execInspectErr
	}
	return client.ExecInspectResult{ExitCode: f.execInspectCode}, nil
}
func (f *fakeDockerClient) ContainerStatPath(_ context.Context, _ string, p string) (container.PathStat, error) {
	if f.statPathErrMap != nil {
		if err, ok := f.statPathErrMap[p]; ok {
			return container.PathStat{}, err
		}
	}
	if f.statPathMap != nil {
		if stat, ok := f.statPathMap[p]; ok {
			return stat, nil
		}
	}
	return f.statPath, f.statPathErr
}
func (f *fakeDockerClient) ListContainerDir(_ context.Context, _ string, p string) (docker.ContainerDirListing, error) {
	if f.listDirErrMap != nil {
		if err, ok := f.listDirErrMap[p]; ok {
			return docker.ContainerDirListing{}, err
		}
	}
	if f.listDirFailMap != nil {
		if fails, ok := f.listDirFailMap[p]; ok {
			return docker.ContainerDirListing{Files: f.listDirMap[p], Failures: fails, Truncated: f.listDirTruncated[p]}, nil
		}
	}
	if f.listDirMap != nil {
		if dir, ok := f.listDirMap[p]; ok {
			return docker.ContainerDirListing{Files: dir, Truncated: f.listDirTruncated[p]}, nil
		}
	}
	return docker.ContainerDirListing{Files: f.listDir}, f.listDirErr
}
func (f *fakeDockerClient) CopyToContainer(_ context.Context, containerID string, dstPath string, content io.Reader, options client.CopyToContainerOptions) error {
	body, _ := io.ReadAll(content)
	f.mu.Lock()
	f.copyToCalls = append(f.copyToCalls, copyToCall{
		containerID: containerID,
		dstPath:     dstPath,
		body:        body,
		options:     options,
	})
	f.mu.Unlock()
	return f.copyToErr
}
func (f *fakeDockerClient) CopyFromContainer(_ context.Context, _ string, containerPath string) (io.ReadCloser, container.PathStat, error) {
	f.mu.Lock()
	f.copyFromCount++
	f.mu.Unlock()

	if f.copyFromErrMap != nil {
		if err, ok := f.copyFromErrMap[containerPath]; ok {
			return nil, container.PathStat{}, err
		}
	}
	if f.copyFromErr != nil {
		return nil, container.PathStat{}, f.copyFromErr
	}
	if f.copyFromBodyMap != nil {
		if body, ok := f.copyFromBodyMap[containerPath]; ok {
			return io.NopCloser(bytes.NewReader(body)), f.copyFromStat, nil
		}
	}
	return io.NopCloser(bytes.NewReader(f.copyFromBody)), f.copyFromStat, nil
}
func (f *fakeDockerClient) Cleanup(_ context.Context) error                        { return nil }
func (f *fakeDockerClient) GetDefaultImage() string                                { return "test-image" }
func (f *fakeDockerClient) VerifyWorkerDockerPolicy(context.Context, string) error { return nil }

type flowFileEvent struct {
	channel string // "flow" or "resource"
	action  string // "added", "updated", "deleted"
	path    string
	id      string
}

// captureFlowPublisher records FlowFile events; all other methods are no-ops.
type captureFlowPublisher struct {
	flowID int64
	userID int64
	events *flowFileCaptureSubscriptions
}

func (p *captureFlowPublisher) GetFlowID() int64   { return p.flowID }
func (p *captureFlowPublisher) SetFlowID(id int64) { p.flowID = id }
func (p *captureFlowPublisher) GetUserID() int64   { return p.userID }
func (p *captureFlowPublisher) SetUserID(id int64) { p.userID = id }
func (p *captureFlowPublisher) FlowCreated(_ context.Context, _ database.Flow, _ []database.Container) {
}
func (p *captureFlowPublisher) FlowDeleted(_ context.Context, _ database.Flow, _ []database.Container) {
}
func (p *captureFlowPublisher) FlowUpdated(_ context.Context, _ database.Flow, _ []database.Container) {
}
func (p *captureFlowPublisher) TaskCreated(_ context.Context, _ database.Task, _ []database.Subtask) {
}
func (p *captureFlowPublisher) TaskUpdated(_ context.Context, _ database.Task, _ []database.Subtask) {
}
func (p *captureFlowPublisher) AssistantCreated(_ context.Context, _ database.Assistant) {}
func (p *captureFlowPublisher) AssistantUpdated(_ context.Context, _ database.Assistant) {}
func (p *captureFlowPublisher) AssistantDeleted(_ context.Context, _ database.Assistant) {}
func (p *captureFlowPublisher) FlowFileAdded(_ context.Context, file *model.FlowFile) {
	p.events.record(flowFileEvent{channel: "flow", action: "added", path: file.Path, id: file.ID})
}
func (p *captureFlowPublisher) FlowFileUpdated(_ context.Context, file *model.FlowFile) {
	p.events.record(flowFileEvent{channel: "flow", action: "updated", path: file.Path, id: file.ID})
}
func (p *captureFlowPublisher) FlowFileDeleted(_ context.Context, file *model.FlowFile) {
	p.events.record(flowFileEvent{channel: "flow", action: "deleted", path: file.Path, id: file.ID})
}
func (p *captureFlowPublisher) ScreenshotAdded(_ context.Context, _ database.Screenshot) {}
func (p *captureFlowPublisher) TerminalLogAdded(_ context.Context, _ database.Termlog)   {}
func (p *captureFlowPublisher) MessageLogAdded(_ context.Context, _ database.Msglog)     {}
func (p *captureFlowPublisher) MessageLogUpdated(_ context.Context, _ database.Msglog)   {}
func (p *captureFlowPublisher) AgentLogAdded(_ context.Context, _ database.Agentlog)     {}
func (p *captureFlowPublisher) SearchLogAdded(_ context.Context, _ database.Searchlog)   {}
func (p *captureFlowPublisher) VectorStoreLogAdded(_ context.Context, _ database.Vecstorelog) {
}
func (p *captureFlowPublisher) ToolCallLogAdded(_ context.Context, _ database.Toolcall) {
}
func (p *captureFlowPublisher) ToolCallLogUpdated(_ context.Context, _ database.Toolcall) {
}
func (p *captureFlowPublisher) AssistantLogAdded(_ context.Context, _ database.Assistantlog) {}
func (p *captureFlowPublisher) AssistantLogUpdated(_ context.Context, _ database.Assistantlog, _ bool) {
}
func (p *captureFlowPublisher) KnowledgeDocumentCreated(_ context.Context, _ *model.KnowledgeDocument) {
}

// captureResourcePublisherForFlow records the Resource events AddResourceFromFlow emits.
type captureResourcePublisherForFlow struct {
	userID int64
	events *flowFileCaptureSubscriptions
}

func (p *captureResourcePublisherForFlow) GetUserID() int64   { return p.userID }
func (p *captureResourcePublisherForFlow) SetUserID(id int64) { p.userID = id }
func (p *captureResourcePublisherForFlow) ResourceAdded(_ context.Context, r *model.UserResource) {
	p.events.record(flowFileEvent{channel: "resource", action: "added", path: r.Path})
}
func (p *captureResourcePublisherForFlow) ResourceUpdated(_ context.Context, r *model.UserResource) {
	p.events.record(flowFileEvent{channel: "resource", action: "updated", path: r.Path})
}
func (p *captureResourcePublisherForFlow) ResourceDeleted(_ context.Context, r *model.UserResource) {
	p.events.record(flowFileEvent{channel: "resource", action: "deleted", path: r.Path})
}

type copyToCall struct {
	containerID string
	dstPath     string
	body        []byte
	options     client.CopyToContainerOptions
}
