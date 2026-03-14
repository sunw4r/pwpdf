//go:build darwin

package platform

type darwinShellIntegrationManager struct{}

func newShellIntegrationManager() ShellIntegrationManager {
	return &darwinShellIntegrationManager{}
}

func (m *darwinShellIntegrationManager) Install(_ string) (string, error) {
	return "Finder integration is not installed automatically on macOS. Use Open With or drag a PDF onto pwpdf.app.", nil
}

func (m *darwinShellIntegrationManager) Uninstall(_ string) (string, error) {
	return "There is no automated macOS shell integration to remove.", nil
}
