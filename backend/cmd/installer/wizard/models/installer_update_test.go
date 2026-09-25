package models

import (
	"errors"
	"strings"
	"testing"

	"pentagi/cmd/installer/processor"
)

func TestInstallerUpdate_InstallerUpdateLines_DescribesTheOfferBeforeAnyDownload(t *testing.T) {
	for _, tc := range []struct {
		name           string
		running        string
		offered        *processor.InstallerPackage
		err            error
		want, unwanted []string
	}{
		{
			name:    "an ordinary offer names the build and says the running installer is not replaced",
			running: "2.0.0",
			offered: &processor.InstallerPackage{Version: "2.1.0", OS: "linux", Arch: "amd64", Size: 25 << 20, Path: "/opt/pentagi/installer_2.1.0"},
			want: []string{"Running version: 2.0.0", "Offered version: 2.1.0", "linux/amd64", "25.0 MiB",
				"/opt/pentagi/installer_2.1.0", "Press Enter to download", "The installer you are running is not replaced"},
			unwanted: []string{"will exit", "the one already running", "already here"},
		},
		{
			name:    "an offer of the version already running says so",
			running: "2.1.0",
			offered: &processor.InstallerPackage{Version: "2.1.0", OS: "linux", Arch: "amd64", Size: 1024, Path: "installer_2.1.0"},
			want:    []string{"The offered build is the one already running"},
		},
		{
			name:    "a file already on disk is announced before enter is pressed",
			running: "2.0.0",
			offered: &processor.InstallerPackage{Version: "2.1.0", OS: "linux", Arch: "amd64", Size: 1024, Path: "installer_2.1.0", Downloaded: true},
			want:    []string{"A file of this name is already here"},
		},
		{
			name:     "a failed lookup shows the failure and nothing else",
			running:  "2.0.0",
			err:      errors.New("update server is unreachable"),
			want:     []string{"update server is unreachable"},
			unwanted: []string{"Press Enter to download", "Running version"},
		},
		{
			name:     "an unanswered lookup says it is asking",
			running:  "2.0.0",
			want:     []string{"Asking the update server"},
			unwanted: []string{"Press Enter to download"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			document := strings.Join(installerUpdateLines(tc.running, tc.offered, tc.err), "\n")
			for _, want := range tc.want {
				contains(t, document, want, tc.name)
			}
			for _, unwanted := range tc.unwanted {
				absent(t, strings.ToLower(document), strings.ToLower(unwanted), tc.name)
			}
		})
	}
}

func TestInstallerUpdate_InstallerMoveInstructions_HandsOverTheCommandThatInstallsTheBuild(t *testing.T) {
	for _, tc := range []struct {
		name           string
		offered        *processor.InstallerPackage
		want, unwanted []string
	}{
		{
			name: "on linux both quoted paths are in the command and no windows warning is shown",
			offered: &processor.InstallerPackage{Version: "2.1.0", OS: "linux", Arch: "amd64", Size: 1024,
				Path: "/opt/pentagi/installer_2.1.0", CurrentPath: "/usr/local/bin/pentagi-installer"},
			want:     []string{"/opt/pentagi/installer_2.1.0", `mv "/opt/pentagi/installer_2.1.0" "/usr/local/bin/pentagi-installer"`},
			unwanted: []string{"close this installer"},
		},
		{
			name: "on windows the command is move and the installer must be closed first",
			offered: &processor.InstallerPackage{Version: "2.1.0", OS: "windows", Arch: "amd64",
				Path: `C:\pentagi\installer_2.1.0.exe`, CurrentPath: `C:\Program Files\PentAGI\installer.exe`},
			want: []string{`move /Y "C:\pentagi\installer_2.1.0.exe" "C:\Program Files\PentAGI\installer.exe"`,
				"close this installer before running the command"},
		},
		{
			name:     "without the running path no command is invented",
			offered:  &processor.InstallerPackage{Version: "2.1.0", OS: "linux", Path: "/opt/pentagi/installer_2.1.0"},
			want:     []string{"/opt/pentagi/installer_2.1.0", "Move it over the installer you launched"},
			unwanted: []string{`mv "`},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			document := strings.Join(installerMoveInstructions(tc.offered), "\n")
			for _, want := range tc.want {
				contains(t, document, want, tc.name)
			}
			for _, unwanted := range tc.unwanted {
				absent(t, strings.ToLower(document), strings.ToLower(unwanted), tc.name)
			}
		})
	}
}
