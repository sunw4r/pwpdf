package gui

import (
	"context"
	"fmt"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/sp/pwpdf/internal/application"
	"github.com/sp/pwpdf/internal/validation"
	"github.com/sp/pwpdf/internal/version"
)

type App struct {
	ctx               context.Context
	workflow          *application.EncryptionWorkflow
	passwordValidator *validation.PasswordValidator
	buildInfo         version.Info
	initialInputPath  string
	queuedSelection   *application.PreparedDocument
	queuedError       string
	domReady          bool
}

func NewApp(
	workflow *application.EncryptionWorkflow,
	passwordValidator *validation.PasswordValidator,
	buildInfo version.Info,
	initialInputPath string,
) *App {
	return &App{
		workflow:          workflow,
		passwordValidator: passwordValidator,
		buildInfo:         buildInfo,
		initialInputPath:  initialInputPath,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) OnDomReady(_ context.Context) {
	a.domReady = true
	a.flushQueuedNotification()
}

func (a *App) GetApplicationState() ApplicationState {
	state := ApplicationState{
		Version: a.buildInfo.Version,
	}

	if strings.TrimSpace(a.initialInputPath) == "" {
		return state
	}

	preparedDocument, err := a.workflow.PrepareInput(a.initialInputPath)
	if err != nil {
		state.PendingError = err.Error()
		a.initialInputPath = ""
		return state
	}

	state.PendingSelection = preparedDocument
	a.initialInputPath = ""
	return state
}

func (a *App) SelectPDF() (*application.PreparedDocument, error) {
	selectedPath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select a PDF to protect",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "PDF Files (*.pdf)",
				Pattern:     "*.pdf",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to open the file picker: %w", err)
	}

	if selectedPath == "" {
		return nil, nil
	}

	return a.workflow.PrepareInput(selectedPath)
}

func (a *App) SetInputFile(inputPath string) (*application.PreparedDocument, error) {
	return a.workflow.PrepareInput(inputPath)
}

func (a *App) EncryptPDF(form EncryptionForm) (*OperationResult, error) {
	if err := a.passwordValidator.ValidateConfirmation(form.Password, form.PasswordConfirmation); err != nil {
		return nil, err
	}

	preparedDocument, err := a.workflow.PrepareInput(form.InputPath)
	if err != nil {
		return nil, err
	}

	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Save encrypted PDF",
		DefaultDirectory: preparedDocument.Directory,
		DefaultFilename:  preparedDocument.DefaultOutputFilename,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "PDF Files (*.pdf)",
				Pattern:     "*.pdf",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to open the save dialog: %w", err)
	}

	if savePath == "" {
		return &OperationResult{Cancelled: true}, nil
	}

	result, err := a.workflow.Encrypt(application.EncryptRequest{
		InputPath:                    preparedDocument.InputPath,
		OutputPath:                   savePath,
		UserPassword:                 form.Password,
		OwnerPassword:                form.OwnerPassword,
		AllowExistingOutputOverwrite: true,
		AllowInputOverwrite:          false,
	})
	if err != nil {
		return nil, err
	}

	return &OperationResult{
		OutputPath:     result.OutputPath,
		OutputFilename: result.OutputFilename,
		Message:        result.Message,
	}, nil
}

func (a *App) HandleSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	inputPath := firstUsableArgument(secondInstanceData.Args)
	if inputPath == "" {
		return
	}

	preparedDocument, err := a.workflow.PrepareInput(inputPath)
	if err != nil {
		a.queueOrEmitError(err.Error())
		return
	}

	a.queueOrEmitSelection(preparedDocument)
}

func (a *App) flushQueuedNotification() {
	if a.queuedSelection != nil {
		runtime.EventsEmit(a.ctx, EventPreparedDocument, a.queuedSelection)
		a.queuedSelection = nil
	}

	if a.queuedError != "" {
		runtime.EventsEmit(a.ctx, EventErrorMessage, a.queuedError)
		a.queuedError = ""
	}
}

func (a *App) queueOrEmitSelection(selection *application.PreparedDocument) {
	if a.domReady && a.ctx != nil {
		runtime.EventsEmit(a.ctx, EventPreparedDocument, selection)
		return
	}

	a.queuedSelection = selection
}

func (a *App) queueOrEmitError(message string) {
	if a.domReady && a.ctx != nil {
		runtime.EventsEmit(a.ctx, EventErrorMessage, message)
		return
	}

	a.queuedError = message
}

func firstUsableArgument(arguments []string) string {
	for _, argument := range arguments {
		if argument == "" || strings.HasPrefix(argument, "-") {
			continue
		}
		return argument
	}

	return ""
}
