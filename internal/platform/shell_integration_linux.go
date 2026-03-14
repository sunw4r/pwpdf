//go:build linux

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type linuxShellIntegrationManager struct{}

func newShellIntegrationManager() ShellIntegrationManager {
	return &linuxShellIntegrationManager{}
}

func (m *linuxShellIntegrationManager) Install(applicationPath string) (string, error) {
	absoluteApplicationPath, err := filepath.Abs(applicationPath)
	if err != nil {
		return "", fmt.Errorf("unable to resolve the application path: %w", err)
	}

	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to resolve the home directory: %w", err)
	}

	applicationsDirectory := filepath.Join(homeDirectory, ".local", "share", "applications")
	if err := os.MkdirAll(applicationsDirectory, 0o755); err != nil {
		return "", fmt.Errorf("unable to create the applications directory: %w", err)
	}

	desktopFilePath := filepath.Join(applicationsDirectory, "pwpdf.desktop")
	desktopFileContents := fmt.Sprintf(`[Desktop Entry]
Type=Application
Version=1.0
Name=pwpdf
GenericName=PDF Protector
Comment=Add password protection to PDF files
Exec=%q %%f
Terminal=false
Categories=Office;Utility;
MimeType=application/pdf;
StartupNotify=true
`, absoluteApplicationPath)

	if err := os.WriteFile(desktopFilePath, []byte(desktopFileContents), 0o644); err != nil {
		return "", fmt.Errorf("unable to write the desktop entry: %w", err)
	}

	if updateDesktopDatabasePath, err := exec.LookPath("update-desktop-database"); err == nil {
		_ = exec.Command(updateDesktopDatabasePath, applicationsDirectory).Run()
	}

	return fmt.Sprintf("Installed Linux desktop integration at %s", desktopFilePath), nil
}

func (m *linuxShellIntegrationManager) Uninstall(_ string) (string, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to resolve the home directory: %w", err)
	}

	applicationsDirectory := filepath.Join(homeDirectory, ".local", "share", "applications")
	desktopFilePath := filepath.Join(applicationsDirectory, "pwpdf.desktop")
	if err := os.Remove(desktopFilePath); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("unable to remove the desktop entry: %w", err)
	}

	if updateDesktopDatabasePath, err := exec.LookPath("update-desktop-database"); err == nil {
		_ = exec.Command(updateDesktopDatabasePath, applicationsDirectory).Run()
	}

	return "Removed Linux desktop integration.", nil
}
