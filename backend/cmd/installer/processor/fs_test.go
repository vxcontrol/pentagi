package processor

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fsStackFiles are what each stack extracts into the working directory.
var fsStackFiles = map[ProductStack][]string{
	ProductStackPentagi:       {"docker-compose.yml", "example.custom.provider.yml", "example.ollama.provider.yml"},
	ProductStackGraphiti:      {"docker-compose-graphiti.yml", "graphiti", "neo4j"},
	ProductStackLangfuse:      {"docker-compose-langfuse.yml"},
	ProductStackObservability: {"docker-compose-observability.yml", "observability"},
}

// fsEveryStackFile is every stack's files, in the order the stacks are processed.
var fsEveryStackFile = slices.Concat(
	fsStackFiles[ProductStackPentagi], fsStackFiles[ProductStackGraphiti],
	fsStackFiles[ProductStackLangfuse], fsStackFiles[ProductStackObservability],
)

// fsOperations is the real file system operations over mocked embedded files, in a temp directory.
func fsOperations(t *testing.T) (*fileSystemOperationsImpl, *mockFiles, string) {
	t.Helper()
	state, embedded := newMockState(t), newMockFiles()
	ops := newFileSystemOperations(&processor{state: state, files: embedded}).(*fileSystemOperationsImpl)
	return ops, embedded, filepath.Dir(state.envPath)
}

// fsCopied lists what was extracted, failing on a copy outside the working directory or without rewrite.
func fsCopied(t *testing.T, embedded *mockFiles, workingDir string) []string {
	t.Helper()
	var copied []string
	for _, c := range embedded.copies {
		assert.Equal(t, workingDir, c.Dst, "copy of %s", c.Src)
		assert.True(t, c.Rewrite, "copy of %s", c.Src)
		copied = append(copied, c.Src)
	}
	return copied
}

func TestFs_EnsureStackIntegrity_ExtractsEveryFileOfTheStack(t *testing.T) {
	for _, tc := range []struct {
		name    string
		stack   ProductStack
		want    []string
		wantErr string
	}{
		{"pentagi with its provider examples", ProductStackPentagi, fsStackFiles[ProductStackPentagi], ""},
		{"langfuse has its compose file only", ProductStackLangfuse, fsStackFiles[ProductStackLangfuse], ""},
		{"observability with its directory", ProductStackObservability, fsStackFiles[ProductStackObservability], ""},
		{"compose covers every stack", ProductStackCompose, fsEveryStackFile, ""},
		{"all covers every stack", ProductStackAll, fsEveryStackFile, ""},
		{"the worker has no files", ProductStackWorker, nil, "operation ensure integrity not applicable for stack worker"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ops, embedded, dir := fsOperations(t)
			err := ops.ensureStackIntegrity(t.Context(), tc.stack, testOperationState(t))
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.wantErr)
			}
			assert.Equal(t, tc.want, fsCopied(t, embedded, dir))
		})
	}
}

func TestFs_VerifyStackIntegrity_RestoresEveryMissingFileItVerifies(t *testing.T) {
	for _, tc := range []struct {
		name    string
		stack   ProductStack
		want    []string
		wantErr string
	}{
		{"pentagi verifies its compose file only", ProductStackPentagi, []string{"docker-compose.yml"}, ""},
		{"langfuse has its compose file only", ProductStackLangfuse, []string{"docker-compose-langfuse.yml"}, ""},
		{"observability with its directory", ProductStackObservability, fsStackFiles[ProductStackObservability], ""},
		{"compose covers every stack", ProductStackCompose, []string{
			"docker-compose.yml", "docker-compose-graphiti.yml", "graphiti", "neo4j",
			"docker-compose-langfuse.yml", "docker-compose-observability.yml", "observability",
		}, ""},
		{"all covers every stack", ProductStackAll, []string{
			"docker-compose.yml", "docker-compose-graphiti.yml", "graphiti", "neo4j",
			"docker-compose-langfuse.yml", "docker-compose-observability.yml", "observability",
		}, ""},
		{"the worker has no files", ProductStackWorker, nil, "operation verify integrity not applicable for stack worker"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ops, embedded, dir := fsOperations(t)
			err := ops.verifyStackIntegrity(t.Context(), tc.stack, testOperationState(t))
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.wantErr)
			}
			assert.Equal(t, tc.want, fsCopied(t, embedded, dir))
		})
	}
}

func TestFs_CleanupStackFiles_RemovesEveryFileOfTheStackAndNothingElse(t *testing.T) {
	for _, tc := range []struct {
		name    string
		stack   ProductStack
		removed []string
		wantErr string
	}{
		{"pentagi with its provider examples", ProductStackPentagi, fsStackFiles[ProductStackPentagi], ""},
		{"graphiti with its config directories", ProductStackGraphiti, fsStackFiles[ProductStackGraphiti], ""},
		{"langfuse has its compose file only", ProductStackLangfuse, fsStackFiles[ProductStackLangfuse], ""},
		{"observability with its directory", ProductStackObservability, fsStackFiles[ProductStackObservability], ""},
		{"compose covers every stack", ProductStackCompose, fsEveryStackFile, ""},
		{"all covers every stack", ProductStackAll, fsEveryStackFile, ""},
		{"the worker has no files", ProductStackWorker, nil, "operation cleanup not applicable for stack worker"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ops, _, dir := fsOperations(t)
			for _, name := range fsEveryStackFile {
				if filepath.Ext(name) == "" {
					require.NoError(t, os.MkdirAll(filepath.Join(dir, name, "conf"), 0o755))
				} else {
					require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644))
				}
			}

			err := ops.cleanupStackFiles(t.Context(), tc.stack, testOperationState(t))
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.wantErr)
			}
			for _, name := range append([]string{".env"}, fsEveryStackFile...) {
				_, statErr := os.Stat(filepath.Join(dir, name))
				if slices.Contains(tc.removed, name) {
					assert.ErrorIs(t, statErr, os.ErrNotExist, "%s survived the cleanup", name)
				} else {
					assert.NoError(t, statErr, "the cleanup removed %s, which is not the stack's", name)
				}
			}
		})
	}
}

func TestFs_EnsureFileFromEmbed_CopiesOnlyWhatIsMissingUnlessForced(t *testing.T) {
	for _, tc := range []struct {
		name          string
		onDisk, force bool
		want          []string
	}{
		{"a missing file is extracted", false, false, []string{"test.yml"}},
		{"a file on disk is kept", true, false, nil},
		{"a file on disk is overwritten when forced", true, true, []string{"test.yml"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ops, embedded, dir := fsOperations(t)
			embedded.AddFile("test.yml", []byte("test content"))
			if tc.onDisk {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "test.yml"), []byte("existing"), 0o644))
			}
			state := testOperationState(t)
			state.force = tc.force

			require.NoError(t, ops.ensureFileFromEmbed("test.yml", state))
			assert.Equal(t, tc.want, fsCopied(t, embedded, dir))
		})
	}
}

// Driven through verifyDirectoryIntegrity over a directory on disk, where a forced update enters it.
func TestFs_VerifyDirectoryContentIntegrity_RestoresMissingFilesAndSparesEditedOnes(t *testing.T) {
	regular, other := "observability/config1.yml", "observability/config2.yml"
	excluded := "observability/otel/config.yml"
	for _, tc := range []struct {
		name     string
		dir      string
		statuses map[string]files.FileStatus // the embedded directory's files; none means it is not shipped
		force    bool
		want     []string
		wantErr  string
	}{
		{name: "a directory the installer does not ship", dir: "nonexistent",
			wantErr: "embedded directory nonexistent not found"},
		{name: "every file intact", dir: "observability",
			statuses: map[string]files.FileStatus{regular: files.FileStatusOK, other: files.FileStatusOK}},
		{name: "a missing file is restored", dir: "observability",
			statuses: map[string]files.FileStatus{regular: files.FileStatusOK, other: files.FileStatusMissing},
			want:     []string{other}},
		{name: "an edited file is kept", dir: "observability",
			statuses: map[string]files.FileStatus{regular: files.FileStatusModified}},
		{name: "a missing excluded file is created", dir: "observability",
			statuses: map[string]files.FileStatus{excluded: files.FileStatusMissing},
			want:     []string{excluded}},
		{name: "an edited excluded file is kept", dir: "observability",
			statuses: map[string]files.FileStatus{excluded: files.FileStatusModified}},
		{name: "a forced update overwrites an edited file but not a user-editable one", dir: "neo4j", force: true,
			statuses: map[string]files.FileStatus{
				"neo4j/conf/neo4j.conf":       files.FileStatusModified,
				"neo4j/plugins/apoc-core.jar": files.FileStatusModified,
			},
			want: []string{"neo4j/plugins/apoc-core.jar"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ops, embedded, dir := fsOperations(t)
			if tc.statuses != nil {
				embedded.statuses = tc.statuses
				embedded.lists[tc.dir] = slices.Sorted(maps.Keys(tc.statuses))
			}
			require.NoError(t, os.MkdirAll(filepath.Join(dir, tc.dir), 0o755))
			state := testOperationState(t)
			state.force = tc.force

			err := ops.verifyDirectoryIntegrity(tc.dir, state)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.wantErr)
			}
			assert.Equal(t, tc.want, fsCopied(t, embedded, dir))
		})
	}
}

// Created once and never overwritten afterwards, even by a forced files update.
func TestFs_IsExcludedFromVerification_CoversUserEditableFilesAndTheJaegerPlugin(t *testing.T) {
	paths := []string{
		"neo4j/conf/neo4j.conf",
		"neo4j/conf/apoc.conf",
		"graphiti/custom.yaml",
		"graphiti/gemini.yaml",
		"graphiti/litellm.yaml",
		"graphiti/openai.yaml",
	}
	// Tied to the plugin table, so a third architecture cannot be added on one side only.
	for _, name := range checker.JaegerPluginBinaries {
		paths = append(paths, checker.JaegerPluginDir+"/"+name)
	}

	fs := &fileSystemOperationsImpl{}
	for _, path := range paths {
		assert.True(t, fs.isExcludedFromVerification(path), "%s is not excluded, so a forced update overwrites it", path)
	}
}

// Subtests are keyed by path; each asserts both fileExists and directoryExists.
func TestFs_TellsAFileFromADirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	require.NoError(t, os.WriteFile(file, []byte("test"), 0o644))

	fs := &fileSystemOperationsImpl{}
	for _, tc := range []struct {
		name          string
		path          string
		isFile, isDir bool
	}{
		{"a file", file, true, false},
		{"a directory", dir, false, true},
		{"a missing path", filepath.Join(dir, "missing"), false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.isFile, fs.fileExists(tc.path), "fileExists")
			assert.Equal(t, tc.isDir, fs.directoryExists(tc.path), "directoryExists")
		})
	}
}

func TestFs_ValidateYamlFile_RejectsOnlyMalformedYAML(t *testing.T) {
	fs := &fileSystemOperationsImpl{}
	for _, tc := range []struct {
		name, content, wantErr string
	}{
		{"a compose file", "\nversion: '3.8'\nservices:\n  app:\n    image: nginx\n    ports:\n      - \"80:80\"\n", ""},
		{"an unterminated quote",
			"\nversion: '3.8'\nservices:\n  app:\n    image: nginx\n    ports:\n      - \"80:80\n    # missing closing quote\n",
			"invalid YAML syntax"},
		{"an empty file", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "test.yml")
			require.NoError(t, os.WriteFile(path, []byte(tc.content), 0o644))

			err := fs.validateYamlFile(path)
			if tc.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}

func TestFs_CheckStackIntegrity_ReportsEveryFileOfTheStack(t *testing.T) {
	statuses := FilesCheckResult{
		"docker-compose.yml":               files.FileStatusOK,
		"docker-compose-graphiti.yml":      files.FileStatusOK,
		"graphiti/openai.yaml":             files.FileStatusOK,
		"neo4j/conf/neo4j.conf":            files.FileStatusOK,
		"neo4j/conf/apoc.conf":             files.FileStatusOK,
		"neo4j/plugins/apoc-core.jar":      files.FileStatusOK,
		"docker-compose-langfuse.yml":      files.FileStatusModified,
		"docker-compose-observability.yml": files.FileStatusMissing,
		"observability/config1.yml":        files.FileStatusOK,
		"observability/config2.yml":        files.FileStatusModified,
		"observability/subdir/config3.yml": files.FileStatusMissing,
	}
	lists := map[string][]string{
		"graphiti":      {"graphiti/openai.yaml"},
		"neo4j":         {"neo4j/conf/neo4j.conf", "neo4j/conf/apoc.conf", "neo4j/plugins/apoc-core.jar"},
		"observability": {"observability/config1.yml", "observability/config2.yml", "observability/subdir/config3.yml"},
	}
	for _, tc := range []struct {
		name    string
		stack   ProductStack
		want    FilesCheckResult
		wantErr string
	}{
		{"pentagi reports its compose file", ProductStackPentagi, FilesCheckResult{"docker-compose.yml": files.FileStatusOK}, ""},
		{"langfuse reports its compose file", ProductStackLangfuse, FilesCheckResult{"docker-compose-langfuse.yml": files.FileStatusModified}, ""},
		{"observability with its directory", ProductStackObservability, FilesCheckResult{
			"docker-compose-observability.yml": files.FileStatusMissing,
			"observability/config1.yml":        files.FileStatusOK,
			"observability/config2.yml":        files.FileStatusModified,
			"observability/subdir/config3.yml": files.FileStatusMissing,
		}, ""},
		{"compose covers every stack", ProductStackCompose, statuses, ""},
		{"all covers every stack", ProductStackAll, statuses, ""},
		{"the worker has no files", ProductStackWorker, FilesCheckResult{},
			"operation check integrity not applicable for stack worker"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ops, embedded, _ := fsOperations(t)
			embedded.statuses, embedded.lists = statuses, lists

			result, err := ops.checkStackIntegrity(t.Context(), tc.stack)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.wantErr)
			}
			assert.Equal(t, tc.want, result)
		})
	}
}
