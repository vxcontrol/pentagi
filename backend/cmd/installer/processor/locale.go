package processor

// Docker operations messages
const (
	MsgPullingImage                = "Pulling image: %s"
	MsgImagePullCompleted          = "Completed pulling %s"
	MsgImagePullFailed             = "Failed to pull image %s: %v"
	MsgRemovingWorkerContainers    = "Removing worker containers"
	MsgStoppingContainer           = "Stopping container %s"
	MsgRemovingContainer           = "Removing container %s"
	MsgContainerRemoved            = "Removed container %s"
	MsgNoWorkerContainersFound     = "No worker containers found"
	MsgWorkerContainersRemoved     = "Removed %d worker containers"
	MsgRemovingImage               = "Removing image: %s"
	MsgImageRemoved                = "Successfully removed image %s"
	MsgImageNotFound               = "Image %s not found (already removed)"
	MsgWorkerImagesRemoveCompleted = "Worker images removal completed"
	MsgEnsuringDockerNetworks      = "Ensuring docker networks exist"
	MsgDockerNetworkExists         = "Docker network exists: %s"
	MsgCreatingDockerNetwork       = "Creating docker network: %s"
	MsgDockerNetworkCreated        = "Docker network created: %s"
	MsgDockerNetworkCreateFailed   = "Failed to create docker network %s: %v"
	MsgRecreatingDockerNetwork     = "Recreating docker network with compose labels: %s"
	MsgDockerNetworkRemoved        = "Docker network removed: %s"
	MsgDockerNetworkRemoveFailed   = "Failed to remove docker network %s: %v"
	MsgDockerNetworkInUse          = "Docker network %s is in use by containers; cannot recreate"
)

// File system operations messages
const (
	MsgExtractingDockerCompose          = "Extracting docker-compose.yml"
	MsgExtractingLangfuseCompose        = "Extracting docker-compose-langfuse.yml"
	MsgExtractingObservabilityCompose   = "Extracting docker-compose-observability.yml"
	MsgExtractingObservabilityDirectory = "Extracting observability directory"
	MsgSkippingExternalLangfuse         = "Skipping external Langfuse deployment"
	MsgSkippingExternalObservability    = "Skipping external Observability deployment"
	MsgPatchingComposeFile              = "Patching docker-compose file: %s"
	MsgComposePatchCompleted            = "Docker-compose file patching completed"
	MsgCleaningUpStackFiles             = "Cleaning up stack files for %s"
	MsgStackFilesCleanupCompleted       = "Stack files cleanup completed"
	MsgEnsurngStackIntegrity            = "Ensuring %s stack integrity"
	MsgVerifyingStackIntegrity          = "Verifying %s stack integrity"
	MsgStackIntegrityVerified           = "Stack %s integrity verified"
	MsgUpdatingExistingFile             = "Updating existing file: %s"
	MsgCreatingMissingFile              = "Creating missing file: %s"
	MsgFileIntegrityValid               = "File integrity valid: %s"
	MsgSkippingModifiedFile             = "Skipping modified files: %s"
	MsgDirectoryCheckedWithModified     = "Directory checked with modified files present: %s"
)

// Update operations messages
const (
	MsgCheckingUpdates            = "Checking for updates"
	MsgDownloadingInstaller       = "Downloading installer update"
	MsgInstallerDownloadCompleted = "Installer download completed"
	MsgUpdatingInstaller          = "Updating installer"
	MsgRemovingInstaller          = "Removing installer"
	MsgInstallerUpdateCompleted   = "Installer update completed"
	MsgVerifyingBinaryChecksum    = "Verifying binary checksum"
	MsgReplacingInstallerBinary   = "Replacing installer binary"
)

// Remove operations messages
const (
	MsgRemovingStack          = "Removing stack: %s"
	MsgStackRemovalCompleted  = "Stack removal completed for %s"
	MsgPurgingStack           = "Purging stack: %s"
	MsgStackPurgeCompleted    = "Stack purge completed for %s"
	MsgExecutingDockerCompose = "Executing docker-compose command: %s"
	MsgDockerComposeCompleted = "Docker-compose command completed"
	MsgFactoryResetStarting   = "Starting factory reset"
	MsgFactoryResetCompleted  = "Factory reset completed"
	MsgRestoringDefaultEnv    = "Restoring default .env from embedded"
	MsgDefaultEnvRestored     = "Default .env restored"
)

type Subsystem string

const (
	SubsystemDocker     Subsystem = "docker"
	SubsystemCompose    Subsystem = "compose"
	SubsystemFileSystem Subsystem = "file-system"
	SubsystemUpdate     Subsystem = "update"
)

type SubsystemOperationMessage struct {
	Enter string
	Exit  string
	Error string
}

var SubsystemOperationMessages = map[Subsystem]map[ProcessorOperation]SubsystemOperationMessage{
	SubsystemCompose: {
		ProcessorOperationStart: SubsystemOperationMessage{
			Enter: "Starting %s compose stack",
			Exit:  "Compose stack %s was started",
			Error: "Failed to start %s compose stack",
		},
		ProcessorOperationStop: SubsystemOperationMessage{
			Enter: "Stopping %s compose stack",
			Exit:  "Compose stack %s was stopped",
			Error: "Failed to stop %s compose stack",
		},
		ProcessorOperationRestart: SubsystemOperationMessage{
			Enter: "Restarting %s compose stack",
			Exit:  "Compose stack %s was restarted",
			Error: "Failed to restart %s compose stack",
		},
		ProcessorOperationUpdate: SubsystemOperationMessage{
			Enter: "Updating %s compose stack",
			Exit:  "Compose stack %s was updated",
			Error: "Failed to update %s compose stack",
		},
		ProcessorOperationDownload: SubsystemOperationMessage{
			Enter: "Downloading %s compose stack",
			Exit:  "Compose stack %s was downloaded",
			Error: "Failed to download %s compose stack",
		},
		ProcessorOperationRemove: SubsystemOperationMessage{
			Enter: "Removing %s compose stack",
			Exit:  "Compose stack %s was removed",
			Error: "Failed to remove %s compose stack",
		},
		ProcessorOperationPurge: SubsystemOperationMessage{
			Enter: "Purging %s compose stack",
			Exit:  "Compose stack %s was purged",
			Error: "Failed to purge %s compose stack",
		},
	},
}
