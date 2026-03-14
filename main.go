package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailslinux "github.com/wailsapp/wails/v2/pkg/options/linux"
	wailsmac "github.com/wailsapp/wails/v2/pkg/options/mac"
	wailswindows "github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/sp/pwpdf/internal/application"
	"github.com/sp/pwpdf/internal/cli"
	"github.com/sp/pwpdf/internal/files"
	"github.com/sp/pwpdf/internal/gui"
	"github.com/sp/pwpdf/internal/pdf"
	"github.com/sp/pwpdf/internal/platform"
	"github.com/sp/pwpdf/internal/validation"
	"github.com/sp/pwpdf/internal/version"
)

//go:embed all:frontend/dist
var frontendAssets embed.FS

//go:embed build/appicon.png
var applicationIcon []byte

func main() {
	api.DisableConfigDir()

	buildInfo := version.BuildInfo()
	passwordValidator := validation.NewPasswordValidator()
	fileValidator := validation.NewFileValidator()
	outputPathResolver := files.NewOutputPathResolver()
	encryptionService := pdf.NewEncryptionService()
	workflow := application.NewEncryptionWorkflow(
		passwordValidator,
		fileValidator,
		outputPathResolver,
		encryptionService,
	)

	executionPlan := application.DetectExecution(os.Args[1:])
	if executionPlan.Mode == application.ModeCLI {
		commandHandler := cli.NewCommandHandler(
			workflow,
			platform.NewShellIntegrationManager(),
			buildInfo,
			os.Stdin,
			os.Stdout,
			os.Stderr,
		)
		os.Exit(commandHandler.Run(executionPlan.CommandArgs))
	}

	desktopApp := gui.NewApp(
		workflow,
		passwordValidator,
		buildInfo,
		executionPlan.PreselectedInputPath,
	)

	if err := runDesktopApplication(desktopApp); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runDesktopApplication(desktopApp *gui.App) error {
	return wails.Run(&options.App{
		Title:     "pwpdf",
		Width:     580,
		Height:    560,
		MinWidth:  336,
		MinHeight: 350,
		AssetServer: &assetserver.Options{
			Assets: frontendAssets,
		},
		BackgroundColour: options.NewRGB(7, 3, 17),
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
		},
		OnStartup: func(ctx context.Context) {
			desktopApp.Startup(ctx)
		},
		OnDomReady: func(ctx context.Context) {
			desktopApp.OnDomReady(ctx)
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "github.com/sp/pwpdf",
			OnSecondInstanceLaunch: desktopApp.HandleSecondInstanceLaunch,
		},
		Windows: &wailswindows.Options{
			ResizeDebounceMS:     16,
			WebviewGpuIsDisabled: false,
		},
		Mac: &wailsmac.Options{
			About: &wailsmac.AboutInfo{
				Title: "pwpdf",
				Icon:  applicationIcon,
			},
		},
		Linux: &wailslinux.Options{
			WebviewGpuPolicy: wailslinux.WebviewGpuPolicyOnDemand,
			ProgramName:      "pwpdf",
			Icon:             applicationIcon,
		},
		Bind: []interface{}{
			desktopApp,
		},
	})
}
