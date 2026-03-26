package processor

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"pentagi/cmd/installer/checker"
	"pentagi/pkg/version"
)

const updateServerURL = "https://update.pentagi.com"

type updateOperationsImpl struct {
	processor *processor
}

func newUpdateOperations(p *processor) updateOperations {
	return &updateOperationsImpl{processor: p}
}

func (u *updateOperationsImpl) checkUpdates(ctx context.Context, state *operationState) (*checker.CheckUpdatesResponse, error) {
	u.processor.appendLog(MsgCheckingUpdates, ProductStackInstaller, state)

	request := u.buildUpdateCheckRequest()
	serverURL := u.getUpdateServerURL()
	proxyURL := u.getProxyURL()

	response, err := u.callUpdateServer(ctx, serverURL, proxyURL, request)
	if err != nil {
		return nil, fmt.Errorf("failed to check updates: %w", err)
	}

	return response, nil
}

func (u *updateOperationsImpl) downloadInstaller(ctx context.Context, state *operationState) error {
	u.processor.appendLog(MsgDownloadingInstaller, ProductStackInstaller, state)

	downloadURL, err := u.getInstallerDownloadURL(ctx)
	if err != nil {
		return err
	}

	tempFile, err := u.downloadBinaryToTemp(ctx, downloadURL)
	if err != nil {
		return err
	}
	defer os.Remove(tempFile)

	u.processor.appendLog(MsgVerifyingBinaryChecksum, ProductStackInstaller, state)
	if err := u.verifyBinaryChecksum(tempFile); err != nil {
		return err
	}

	// TODO: copy binary to current update directory

	u.processor.appendLog(MsgInstallerUpdateCompleted, ProductStackInstaller, state)
	return fmt.Errorf("not implemented")
}

func (u *updateOperationsImpl) updateInstaller(ctx context.Context, state *operationState) error {
	u.processor.appendLog(MsgUpdatingInstaller, ProductStackInstaller, state)

	downloadURL, err := u.getInstallerDownloadURL(ctx)
	if err != nil {
		return err
	}

	tempFile, err := u.downloadBinaryToTemp(ctx, downloadURL)
	if err != nil {
		return err
	}
	defer os.Remove(tempFile)

	u.processor.appendLog(MsgVerifyingBinaryChecksum, ProductStackInstaller, state)
	if err := u.verifyBinaryChecksum(tempFile); err != nil {
		return err
	}

	// TODO: replace installer binary after communication with current installer process
	u.processor.appendLog(MsgReplacingInstallerBinary, ProductStackInstaller, state)
	if err := u.replaceInstallerBinary(tempFile); err != nil {
		return err
	}

	u.processor.appendLog(MsgInstallerUpdateCompleted, ProductStackInstaller, state)
	return fmt.Errorf("not implemented")
}

func (u *updateOperationsImpl) removeInstaller(ctx context.Context, state *operationState) error {
	u.processor.appendLog(MsgRemovingInstaller, ProductStackInstaller, state)

	// TODO: remove installer binary

	return fmt.Errorf("not implemented")
}

func (u *updateOperationsImpl) buildUpdateCheckRequest() checker.CheckUpdatesRequest {
	currentVersion := version.GetBinaryVersion()
	if versionVar, exists := u.processor.state.GetVar("PENTAGI_VERSION"); exists {
		currentVersion = versionVar.Value
	}

	return checker.CheckUpdatesRequest{
		InstallerOsType:        runtime.GOOS,
		InstallerVersion:       currentVersion,
		GraphitiConnected:      u.processor.checker.GraphitiConnected,
		GraphitiExternal:       u.processor.checker.GraphitiExternal,
		GraphitiInstalled:      u.processor.checker.GraphitiInstalled,
		LangfuseConnected:      u.processor.checker.LangfuseConnected,
		LangfuseExternal:       u.processor.checker.LangfuseExternal,
		LangfuseInstalled:      u.processor.checker.LangfuseInstalled,
		ObservabilityConnected: u.processor.checker.ObservabilityConnected,
		ObservabilityExternal:  u.processor.checker.ObservabilityExternal,
		ObservabilityInstalled: u.processor.checker.ObservabilityInstalled,
	}
}

func (u *updateOperationsImpl) callUpdateServer(
	ctx context.Context,
	serverURL, proxyURL string,
	request checker.CheckUpdatesRequest,
) (*checker.CheckUpdatesResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if proxyURL != "" {
		proxyURLParsed, err := url.Parse(proxyURL)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURLParsed),
			}
		}
	}

	return u.callExistingUpdateChecker(ctx, serverURL, client, request)
}

func (u *updateOperationsImpl) getInstallerDownloadURL(ctx context.Context) (string, error) {
	response, err := u.checkUpdates(ctx, &operationState{})
	if err != nil {
		return "", err
	}

	if response.InstallerIsUpToDate {
		return "", fmt.Errorf("no update available")
	}

	return "https://update.pentagi.com/installer", nil
}

func (u *updateOperationsImpl) downloadBinaryToTemp(ctx context.Context, downloadURL string) (string, error) {
	client := &http.Client{
		Timeout: 300 * time.Second,
	}

	resp, err := client.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %s", resp.Status)
	}

	tempFile, err := os.CreateTemp("", "pentagi-update-*.bin")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to write downloaded binary: %w", err)
	}

	return tempFile.Name(), nil
}

func (u *updateOperationsImpl) verifyBinaryChecksum(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open binary for verification: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return nil
}

func (u *updateOperationsImpl) replaceInstallerBinary(newBinaryPath string) error {
	currentBinary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current binary path: %w", err)
	}

	backupPath := currentBinary + ".backup"
	if err := u.copyFile(currentBinary, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	if err := u.copyFile(newBinaryPath, currentBinary); err != nil {
		u.copyFile(backupPath, currentBinary)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	if err := os.Chmod(currentBinary, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	os.Remove(backupPath)
	return nil
}

func (u *updateOperationsImpl) getUpdateServerURL() string {
	if serverVar, exists := u.processor.state.GetVar("UPDATE_SERVER_URL"); exists && serverVar.Value != "" {
		return serverVar.Value
	}

	return "https://update.pentagi.com"
}

func (u *updateOperationsImpl) getProxyURL() string {
	if proxyVar, exists := u.processor.state.GetVar("HTTP_PROXY"); exists {
		return proxyVar.Value
	}

	return ""
}

func (u *updateOperationsImpl) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}

func (u *updateOperationsImpl) callExistingUpdateChecker(
	ctx context.Context,
	url string,
	client *http.Client,
	request checker.CheckUpdatesRequest,
) (*checker.CheckUpdatesResponse, error) {
	return &checker.CheckUpdatesResponse{
		InstallerIsUpToDate:     true,
		PentagiIsUpToDate:       true,
		GraphitiIsUpToDate:      true,
		LangfuseIsUpToDate:      true,
		ObservabilityIsUpToDate: true,
		WorkerIsUpToDate:        true,
	}, nil
}
