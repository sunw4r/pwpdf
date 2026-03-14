//go:build windows

package platform

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const windowsShellKeyPath = `Software\Classes\SystemFileAssociations\.pdf\shell\pwpdf`

type windowsShellIntegrationManager struct{}

func newShellIntegrationManager() ShellIntegrationManager {
	return &windowsShellIntegrationManager{}
}

func (m *windowsShellIntegrationManager) Install(applicationPath string) (string, error) {
	absoluteApplicationPath, err := filepath.Abs(applicationPath)
	if err != nil {
		return "", fmt.Errorf("unable to resolve the application path: %w", err)
	}

	shellKey, _, err := registry.CreateKey(registry.CURRENT_USER, windowsShellKeyPath, registry.SET_VALUE)
	if err != nil {
		return "", fmt.Errorf("unable to create the Windows shell integration key: %w", err)
	}
	defer shellKey.Close()

	if err := shellKey.SetStringValue("", "Protect PDF with pwpdf"); err != nil {
		return "", fmt.Errorf("unable to configure the context menu label: %w", err)
	}
	if err := shellKey.SetStringValue("Icon", absoluteApplicationPath); err != nil {
		return "", fmt.Errorf("unable to configure the context menu icon: %w", err)
	}

	commandKey, _, err := registry.CreateKey(registry.CURRENT_USER, windowsShellKeyPath+`\command`, registry.SET_VALUE)
	if err != nil {
		return "", fmt.Errorf("unable to create the Windows shell command key: %w", err)
	}
	defer commandKey.Close()

	commandValue := fmt.Sprintf(`"%s" "%%1"`, absoluteApplicationPath)
	if err := commandKey.SetStringValue("", commandValue); err != nil {
		return "", fmt.Errorf("unable to configure the Windows shell command: %w", err)
	}

	return "Installed Windows context menu integration.", nil
}

func (m *windowsShellIntegrationManager) Uninstall(_ string) (string, error) {
	if err := deleteRegistryKeyIfPresent(windowsShellKeyPath + `\command`); err != nil {
		return "", fmt.Errorf("unable to remove the Windows command key: %w", err)
	}

	if err := deleteRegistryKeyIfPresent(windowsShellKeyPath); err != nil {
		return "", fmt.Errorf("unable to remove the Windows shell key: %w", err)
	}

	return "Removed Windows context menu integration.", nil
}

func deleteRegistryKeyIfPresent(path string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, path, registry.QUERY_VALUE)
	if err != nil {
		return nil
	}
	key.Close()

	return registry.DeleteKey(registry.CURRENT_USER, path)
}
