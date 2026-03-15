package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sunw4r/pwpdf/internal/application"
	"github.com/sunw4r/pwpdf/internal/platform"
	"github.com/sunw4r/pwpdf/internal/version"
)

type CommandHandler struct {
	workflow                *application.EncryptionWorkflow
	shellIntegrationManager platform.ShellIntegrationManager
	buildInfo               version.Info
	stdin                   *os.File
	stdout                  io.Writer
	stderr                  io.Writer
}

func NewCommandHandler(
	workflow *application.EncryptionWorkflow,
	shellIntegrationManager platform.ShellIntegrationManager,
	buildInfo version.Info,
	stdin *os.File,
	stdout io.Writer,
	stderr io.Writer,
) *CommandHandler {
	return &CommandHandler{
		workflow:                workflow,
		shellIntegrationManager: shellIntegrationManager,
		buildInfo:               buildInfo,
		stdin:                   stdin,
		stdout:                  stdout,
		stderr:                  stderr,
	}
}

func (h *CommandHandler) Run(args []string) int {
	if len(args) == 0 {
		h.printHelp()
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		h.printHelp()
		return 0
	case "version", "-v", "--version":
		fmt.Fprintln(h.stdout, h.buildInfo.String())
		return 0
	case "encrypt":
		return h.runEncrypt(args[1:])
	case "install-shell-integration":
		return h.runInstallShellIntegration()
	case "uninstall-shell-integration":
		return h.runUninstallShellIntegration()
	default:
		fmt.Fprintf(h.stderr, "Unknown command: %s\n\n", args[0])
		h.printHelp()
		return 2
	}
}

func (h *CommandHandler) runEncrypt(args []string) int {
	encryptFlags := flag.NewFlagSet("encrypt", flag.ContinueOnError)
	encryptFlags.SetOutput(h.stderr)

	var password string
	var outputPath string
	var ownerPassword string
	var overwrite bool

	encryptFlags.StringVar(&password, "password", "", "User password for the encrypted PDF")
	encryptFlags.StringVar(&password, "p", "", "User password for the encrypted PDF")
	encryptFlags.StringVar(&outputPath, "output", "", "Output file path")
	encryptFlags.StringVar(&outputPath, "o", "", "Output file path")
	encryptFlags.StringVar(&ownerPassword, "owner-password", "", "Owner password for the encrypted PDF")
	encryptFlags.BoolVar(&overwrite, "overwrite", false, "Allow overwriting an existing output file or the input file")
	encryptFlags.Usage = func() {
		fmt.Fprintln(h.stderr, "Usage: pwpdf encrypt input.pdf [options]")
		fmt.Fprintln(h.stderr)
		fmt.Fprintln(h.stderr, "Options:")
		encryptFlags.PrintDefaults()
	}

	normalizedArgs, err := normalizeEncryptArgs(args)
	if err != nil {
		fmt.Fprintln(h.stderr, err)
		return 2
	}

	if err := encryptFlags.Parse(normalizedArgs); err != nil {
		return 2
	}

	if encryptFlags.NArg() != 1 {
		fmt.Fprintln(h.stderr, "The encrypt command requires exactly one input PDF.")
		fmt.Fprintln(h.stderr)
		encryptFlags.Usage()
		return 2
	}

	inputPath := encryptFlags.Arg(0)
	if password == "" {
		promptedPassword, err := NewTerminalPrompter(h.stdin, h.stderr).PromptPassword("Password: ")
		if err != nil {
			fmt.Fprintln(h.stderr, err)
			return 1
		}
		password = promptedPassword
	}

	result, err := h.workflow.Encrypt(application.EncryptRequest{
		InputPath:                    inputPath,
		OutputPath:                   outputPath,
		UserPassword:                 password,
		OwnerPassword:                ownerPassword,
		AllowExistingOutputOverwrite: overwrite,
		AllowInputOverwrite:          overwrite,
	})
	if err != nil {
		fmt.Fprintln(h.stderr, err)
		return 1
	}

	fmt.Fprintln(h.stdout, result.Message)
	fmt.Fprintf(h.stdout, "Output: %s\n", result.OutputPath)
	return 0
}

func normalizeEncryptArgs(args []string) ([]string, error) {
	var normalizedArgs []string
	var positionalArgs []string

	for index := 0; index < len(args); index++ {
		argument := args[index]

		switch {
		case argument == "--":
			positionalArgs = append(positionalArgs, args[index+1:]...)
			index = len(args)
		case isEncryptFlagWithInlineValue(argument):
			normalizedArgs = append(normalizedArgs, argument)
		case encryptFlagRequiresValue(argument):
			if index+1 >= len(args) {
				return nil, fmt.Errorf("flag needs an argument: %s", argument)
			}

			normalizedArgs = append(normalizedArgs, argument, args[index+1])
			index++
		case encryptFlagIsBoolean(argument) || strings.HasPrefix(argument, "-"):
			normalizedArgs = append(normalizedArgs, argument)
		default:
			positionalArgs = append(positionalArgs, argument)
		}
	}

	return append(normalizedArgs, positionalArgs...), nil
}

func encryptFlagRequiresValue(argument string) bool {
	switch argument {
	case "--password", "-p", "--output", "-o", "--owner-password":
		return true
	default:
		return false
	}
}

func isEncryptFlagWithInlineValue(argument string) bool {
	return strings.HasPrefix(argument, "--password=") ||
		strings.HasPrefix(argument, "-p=") ||
		strings.HasPrefix(argument, "--output=") ||
		strings.HasPrefix(argument, "-o=") ||
		strings.HasPrefix(argument, "--owner-password=")
}

func encryptFlagIsBoolean(argument string) bool {
	return argument == "--overwrite"
}

func (h *CommandHandler) runInstallShellIntegration() int {
	executablePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(h.stderr, "Unable to resolve the executable path: %v\n", err)
		return 1
	}

	message, err := h.shellIntegrationManager.Install(executablePath)
	if err != nil {
		fmt.Fprintln(h.stderr, err)
		return 1
	}

	fmt.Fprintln(h.stdout, message)
	return 0
}

func (h *CommandHandler) runUninstallShellIntegration() int {
	executablePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(h.stderr, "Unable to resolve the executable path: %v\n", err)
		return 1
	}

	message, err := h.shellIntegrationManager.Uninstall(executablePath)
	if err != nil {
		fmt.Fprintln(h.stderr, err)
		return 1
	}

	fmt.Fprintln(h.stdout, message)
	return 0
}

func (h *CommandHandler) printHelp() {
	fmt.Fprintln(h.stdout, "pwpdf adds password protection to existing PDF files.")
	fmt.Fprintln(h.stdout)
	fmt.Fprintln(h.stdout, "Usage:")
	fmt.Fprintln(h.stdout, "  pwpdf")
	fmt.Fprintln(h.stdout, "  pwpdf /path/to/file.pdf")
	fmt.Fprintln(h.stdout, "  pwpdf encrypt input.pdf [options]")
	fmt.Fprintln(h.stdout, "  pwpdf install-shell-integration")
	fmt.Fprintln(h.stdout, "  pwpdf uninstall-shell-integration")
	fmt.Fprintln(h.stdout, "  pwpdf version")
	fmt.Fprintln(h.stdout)
	fmt.Fprintln(h.stdout, "Options:")
	fmt.Fprintln(h.stdout, "  -h, --help       Show help")
	fmt.Fprintln(h.stdout, "  -v, --version    Show version")
}
