package platform

type ShellIntegrationManager interface {
	Install(applicationPath string) (string, error)
	Uninstall(applicationPath string) (string, error)
}

func NewShellIntegrationManager() ShellIntegrationManager {
	return newShellIntegrationManager()
}
