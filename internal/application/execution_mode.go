package application

import "strings"

type ExecutionMode int

const (
	ModeGUI ExecutionMode = iota
	ModeCLI
)

type ExecutionPlan struct {
	Mode                 ExecutionMode
	CommandArgs          []string
	PreselectedInputPath string
}

func DetectExecution(args []string) ExecutionPlan {
	if len(args) == 0 {
		return ExecutionPlan{Mode: ModeGUI}
	}

	firstArgument := args[0]
	if shouldRunCLI(firstArgument) {
		return ExecutionPlan{
			Mode:        ModeCLI,
			CommandArgs: args,
		}
	}

	return ExecutionPlan{
		Mode:                 ModeGUI,
		PreselectedInputPath: firstArgument,
	}
}

func shouldRunCLI(argument string) bool {
	switch argument {
	case "encrypt", "help", "version", "install-shell-integration", "uninstall-shell-integration", "-h", "--help", "-v", "--version":
		return true
	}

	return strings.HasPrefix(argument, "-")
}
