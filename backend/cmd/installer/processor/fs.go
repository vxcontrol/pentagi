package processor

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"pentagi/cmd/installer/files"

	"gopkg.in/yaml.v3"
)

const (
	observabilityDirectory  = "observability"
	pentagiExampleConfigLLM = "example.provider.yml"
)

var filesToExcludeFromVerification = []string{
	"observability/otel/config.yml",
	"observability/grafana/config/grafana.ini",
	pentagiExampleConfigLLM,
}

var allStacks = []ProductStack{
	ProductStackPentagi,
	ProductStackGraphiti,
	ProductStackLangfuse,
	ProductStackObservability,
}

type fileSystemOperationsImpl struct {
	processor *processor
}

func newFileSystemOperations(p *processor) fileSystemOperations {
	return &fileSystemOperationsImpl{processor: p}
}

func (fs *fileSystemOperationsImpl) ensureStackIntegrity(ctx context.Context, stack ProductStack, state *operationState) error {
	fs.processor.appendLog(fmt.Sprintf(MsgEnsurngStackIntegrity, stack), stack, state)
	defer fs.processor.appendLog("", stack, state)

	switch stack {
	case ProductStackPentagi:
		if err := fs.ensureFileFromEmbed(composeFilePentagi, state); err != nil {
			return err
		}
		return fs.ensureFileFromEmbed(pentagiExampleConfigLLM, state)

	case ProductStackGraphiti:
		return fs.ensureFileFromEmbed(composeFileGraphiti, state)

	case ProductStackLangfuse:
		return fs.ensureFileFromEmbed(composeFileLangfuse, state)

	case ProductStackObservability:
		if err := fs.ensureFileFromEmbed(composeFileObservability, state); err != nil {
			return err
		}
		return fs.ensureDirectoryFromEmbed(observabilityDirectory, state)

	case ProductStackAll, ProductStackCompose:
		// process all stacks sequentially
		for _, s := range allStacks {
			if err := fs.ensureStackIntegrity(ctx, s, state); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("operation ensure integrity not applicable for stack %s", stack)
	}
}

func (fs *fileSystemOperationsImpl) verifyStackIntegrity(ctx context.Context, stack ProductStack, state *operationState) error {
	fs.processor.appendLog(fmt.Sprintf(MsgVerifyingStackIntegrity, stack), stack, state)
	defer fs.processor.appendLog("", stack, state)

	switch stack {
	case ProductStackPentagi:
		return fs.verifyFileIntegrity(composeFilePentagi, state)

	case ProductStackGraphiti:
		return fs.verifyFileIntegrity(composeFileGraphiti, state)

	case ProductStackLangfuse:
		return fs.verifyFileIntegrity(composeFileLangfuse, state)

	case ProductStackObservability:
		if err := fs.verifyFileIntegrity(composeFileObservability, state); err != nil {
			return err
		}
		return fs.verifyDirectoryIntegrity(observabilityDirectory, state)

	case ProductStackAll, ProductStackCompose:
		// process all stacks sequentially
		for _, s := range allStacks {
			if err := fs.verifyStackIntegrity(ctx, s, state); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("operation verify integrity not applicable for stack %s", stack)
	}
}

// checkStackIntegrity is a silent version of verifyStackIntegrity, used for getting files statuses
func (fs *fileSystemOperationsImpl) checkStackIntegrity(ctx context.Context, stack ProductStack) (FilesCheckResult, error) {
	result := make(FilesCheckResult)

	switch stack {
	case ProductStackPentagi:
		result[composeFilePentagi] = fs.checkFileIntegrity(composeFilePentagi)

	case ProductStackGraphiti:
		result[composeFileGraphiti] = fs.checkFileIntegrity(composeFileGraphiti)

	case ProductStackLangfuse:
		result[composeFileLangfuse] = fs.checkFileIntegrity(composeFileLangfuse)

	case ProductStackObservability:
		result[composeFileObservability] = fs.checkFileIntegrity(composeFileObservability)
		if r, err := fs.checkDirectoryIntegrity(observabilityDirectory); err != nil {
			return result, err
		} else {
			maps.Copy(result, r)
		}

	case ProductStackAll, ProductStackCompose:
		// process all stacks sequentially
		for _, s := range allStacks {
			if r, err := fs.checkStackIntegrity(ctx, s); err != nil {
				return result, err // early exit after first error
			} else {
				maps.Copy(result, r)
			}
		}

	default:
		return result, fmt.Errorf("operation check integrity not applicable for stack %s", stack)
	}

	return result, nil
}

func (fs *fileSystemOperationsImpl) cleanupStackFiles(ctx context.Context, stack ProductStack, state *operationState) error {
	workingDir := filepath.Dir(fs.processor.state.GetEnvPath())
	fs.processor.appendLog(fmt.Sprintf(MsgCleaningUpStackFiles, stack), stack, state)
	defer fs.processor.appendLog("", stack, state)

	var filesToRemove []string

	switch stack {
	case ProductStackPentagi:
		filesToRemove = append(filesToRemove, filepath.Join(workingDir, composeFilePentagi))
		filesToRemove = append(filesToRemove, filepath.Join(workingDir, pentagiExampleConfigLLM))

	case ProductStackGraphiti:
		filesToRemove = append(filesToRemove, filepath.Join(workingDir, composeFileGraphiti))

	case ProductStackLangfuse:
		filesToRemove = append(filesToRemove, filepath.Join(workingDir, composeFileLangfuse))

	case ProductStackObservability:
		filesToRemove = append(filesToRemove, filepath.Join(workingDir, composeFileObservability))
		filesToRemove = append(filesToRemove, filepath.Join(workingDir, observabilityDirectory))

	case ProductStackAll, ProductStackCompose:
		for _, s := range allStacks {
			if err := fs.cleanupStackFiles(ctx, s, state); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("operation cleanup not applicable for stack %s", stack)
	}

	for _, path := range filesToRemove {
		if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	fs.processor.appendLog(MsgStackFilesCleanupCompleted, stack, state)
	return nil
}

func (fs *fileSystemOperationsImpl) ensureFileFromEmbed(filename string, state *operationState) error {
	workingDir := filepath.Dir(fs.processor.state.GetEnvPath())
	targetPath := filepath.Join(workingDir, filename)

	if !state.force && fs.fileExists(targetPath) {
		fs.processor.appendLog(fmt.Sprintf(MsgFileIntegrityValid, filename), ProductStackAll, state)
		return nil
	}

	if fs.fileExists(targetPath) {
		fs.processor.appendLog(fmt.Sprintf(MsgUpdatingExistingFile, filename), ProductStackAll, state)
	} else {
		fs.processor.appendLog(fmt.Sprintf(MsgCreatingMissingFile, filename), ProductStackAll, state)
	}

	return fs.processor.files.Copy(filename, workingDir, true)
}

func (fs *fileSystemOperationsImpl) ensureDirectoryFromEmbed(dirname string, state *operationState) error {
	workingDir := filepath.Dir(fs.processor.state.GetEnvPath())
	targetPath := filepath.Join(workingDir, dirname)

	if !state.force && fs.directoryExists(targetPath) {
		return fs.verifyDirectoryContentIntegrity(dirname, targetPath, state)
	}

	if fs.directoryExists(targetPath) {
		fs.processor.appendLog(fmt.Sprintf(MsgUpdatingExistingFile, dirname), ProductStackAll, state)
	} else {
		fs.processor.appendLog(fmt.Sprintf(MsgCreatingMissingFile, dirname), ProductStackAll, state)
	}

	return fs.processor.files.Copy(dirname, workingDir, true)
}

func (fs *fileSystemOperationsImpl) checkFileIntegrity(filename string) files.FileStatus {
	workingDir := filepath.Dir(fs.processor.state.GetEnvPath())
	return fs.processor.files.Check(filename, workingDir)
}

func (fs *fileSystemOperationsImpl) checkDirectoryIntegrity(dirname string) (FilesCheckResult, error) {
	workingDir := filepath.Dir(fs.processor.state.GetEnvPath())
	targetPath := filepath.Join(workingDir, dirname)
	return fs.checkDirectoryContentIntegrity(dirname, targetPath)
}

func (fs *fileSystemOperationsImpl) checkDirectoryContentIntegrity(embedPath, targetPath string) (FilesCheckResult, error) {
	if !fs.processor.files.Exists(embedPath) {
		return FilesCheckResult{embedPath: files.FileStatusMissing}, nil
	}

	info, err := fs.processor.files.Stat(embedPath)
	if err != nil {
		return FilesCheckResult{}, fmt.Errorf("failed to stat embedded directory %s: %w", embedPath, err)
	}

	if !info.IsDir() {
		return FilesCheckResult{embedPath: fs.checkFileIntegrity(embedPath)}, nil
	}

	// get list of embedded files in directory
	embeddedFiles, err := fs.processor.files.List(embedPath)
	if err != nil {
		return FilesCheckResult{}, fmt.Errorf("failed to list embedded files in %s: %w", embedPath, err)
	}

	// check each embedded file exists and matches in target directory except excluded files
	result := make(FilesCheckResult)
	workingDir := filepath.Dir(targetPath)
	for _, embeddedFile := range embeddedFiles {
		status := fs.processor.files.Check(embeddedFile, workingDir)
		// skip integrity tracking for excluded files but ensure their presence
		if !fs.isExcludedFromVerification(embeddedFile) || status != files.FileStatusModified {
			result[embeddedFile] = status
		}
	}

	return result, nil
}

func (fs *fileSystemOperationsImpl) verifyFileIntegrity(filename string, state *operationState) error {
	workingDir := filepath.Dir(fs.processor.state.GetEnvPath())
	targetPath := filepath.Join(workingDir, filename)

	if !fs.fileExists(targetPath) {
		fs.processor.appendLog(fmt.Sprintf(MsgCreatingMissingFile, filename), ProductStackAll, state)
		return fs.processor.files.Copy(filename, workingDir, true)
	}

	if state.force {
		fs.processor.appendLog(fmt.Sprintf(MsgUpdatingExistingFile, filename), ProductStackAll, state)
		return fs.processor.files.Copy(filename, workingDir, true)
	}

	if err := fs.validateYamlFile(targetPath); err != nil {
		fs.processor.appendLog(fmt.Sprintf(MsgUpdatingExistingFile, filename), ProductStackAll, state)
		return fs.processor.files.Copy(filename, workingDir, true)
	}

	fs.processor.appendLog(fmt.Sprintf(MsgFileIntegrityValid, filename), ProductStackAll, state)
	return nil
}

func (fs *fileSystemOperationsImpl) verifyDirectoryIntegrity(dirname string, state *operationState) error {
	workingDir := filepath.Dir(fs.processor.state.GetEnvPath())
	targetPath := filepath.Join(workingDir, dirname)

	if !fs.directoryExists(targetPath) {
		fs.processor.appendLog(fmt.Sprintf(MsgCreatingMissingFile, dirname), ProductStackAll, state)
		return fs.processor.files.Copy(dirname, workingDir, true)
	}

	if state.force {
		// update directory content selectively, respecting excluded files
		fs.processor.appendLog(fmt.Sprintf(MsgUpdatingExistingFile, dirname), ProductStackAll, state)
		return fs.verifyDirectoryContentIntegrity(dirname, targetPath, state)
	}

	return fs.verifyDirectoryContentIntegrity(dirname, targetPath, state)
}

func (fs *fileSystemOperationsImpl) verifyDirectoryContentIntegrity(embedPath, targetPath string, state *operationState) error {
	if !fs.processor.files.Exists(embedPath) {
		return fmt.Errorf("embedded directory %s not found", embedPath)
	}

	info, err := fs.processor.files.Stat(embedPath)
	if err != nil {
		return fmt.Errorf("failed to stat embedded directory %s: %w", embedPath, err)
	}

	if !info.IsDir() {
		return fs.verifyFileIntegrity(embedPath, state)
	}

	// get list of embedded files in directory
	embeddedFiles, err := fs.processor.files.List(embedPath)
	if err != nil {
		return fmt.Errorf("failed to list embedded files in %s: %w", embedPath, err)
	}

	// verify each embedded file exists and matches in target directory
	// note: targetPath is the full path to the directory we're verifying
	// we need to check files relative to the parent of targetPath
	workingDir := filepath.Dir(targetPath)
	hasUnupdatedModified := false
	for _, embeddedFile := range embeddedFiles {
		// skip integrity tracking for excluded files but ensure their presence
		if fs.isExcludedFromVerification(embeddedFile) {
			status := fs.processor.files.Check(embeddedFile, workingDir)
			if status == files.FileStatusMissing {
				fs.processor.appendLog(fmt.Sprintf(MsgCreatingMissingFile, embeddedFile), ProductStackAll, state)
				if err := fs.processor.files.Copy(embeddedFile, workingDir, true); err != nil {
					return fmt.Errorf("failed to copy missing excluded file %s: %w", embeddedFile, err)
				}
			}
			// do not mark as modified and do not overwrite if present
			continue
		}
		// check file status using optimized hash comparison
		status := fs.processor.files.Check(embeddedFile, workingDir)

		switch status {
		case files.FileStatusMissing:
			fs.processor.appendLog(fmt.Sprintf(MsgCreatingMissingFile, embeddedFile), ProductStackAll, state)
			if err := fs.processor.files.Copy(embeddedFile, workingDir, true); err != nil {
				return fmt.Errorf("failed to copy missing file %s: %w", embeddedFile, err)
			}

		case files.FileStatusModified:
			if state.force {
				fs.processor.appendLog(fmt.Sprintf(MsgUpdatingExistingFile, embeddedFile), ProductStackAll, state)
				if err := fs.processor.files.Copy(embeddedFile, workingDir, true); err != nil {
					return fmt.Errorf("failed to update modified file %s: %w", embeddedFile, err)
				}
			} else {
				hasUnupdatedModified = true
				fs.processor.appendLog(fmt.Sprintf(MsgSkippingModifiedFile, embeddedFile), ProductStackAll, state)
			}

		case files.FileStatusOK:
			// file is valid, no action needed
		}
	}

	if hasUnupdatedModified {
		fs.processor.appendLog(fmt.Sprintf(MsgDirectoryCheckedWithModified, embedPath), ProductStackAll, state)
	} else {
		fs.processor.appendLog(fmt.Sprintf(MsgFileIntegrityValid, embedPath), ProductStackAll, state)
	}

	return nil
}

// isExcludedFromVerification returns true if the provided path should be excluded
// from integrity verification. The file should still exist on disk, but its
// content modifications must not trigger updates or verification failures.
func (fs *fileSystemOperationsImpl) isExcludedFromVerification(path string) bool {
	normalized := filepath.ToSlash(path)
	for _, excluded := range filesToExcludeFromVerification {
		if filepath.ToSlash(excluded) == normalized {
			return true
		}
	}
	return false
}

func (fs *fileSystemOperationsImpl) fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (fs *fileSystemOperationsImpl) directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (fs *fileSystemOperationsImpl) validateYamlFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var data interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	return nil
}
