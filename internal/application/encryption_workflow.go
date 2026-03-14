package application

import (
	"fmt"
	"path/filepath"

	"github.com/sp/pwpdf/internal/files"
	"github.com/sp/pwpdf/internal/pdf"
	"github.com/sp/pwpdf/internal/validation"
)

type EncryptionWorkflow struct {
	passwordValidator  *validation.PasswordValidator
	fileValidator      *validation.FileValidator
	outputPathResolver *files.OutputPathResolver
	encryptionService  *pdf.EncryptionService
}

func NewEncryptionWorkflow(
	passwordValidator *validation.PasswordValidator,
	fileValidator *validation.FileValidator,
	outputPathResolver *files.OutputPathResolver,
	encryptionService *pdf.EncryptionService,
) *EncryptionWorkflow {
	return &EncryptionWorkflow{
		passwordValidator:  passwordValidator,
		fileValidator:      fileValidator,
		outputPathResolver: outputPathResolver,
		encryptionService:  encryptionService,
	}
}

func (w *EncryptionWorkflow) PrepareInput(inputPath string) (*PreparedDocument, error) {
	absoluteInputPath, err := filepath.Abs(inputPath)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve the input path: %w", err)
	}

	if err := w.fileValidator.ValidateInputPDF(absoluteInputPath); err != nil {
		return nil, err
	}

	if err := w.encryptionService.ValidateInputPDF(absoluteInputPath); err != nil {
		return nil, err
	}

	defaultOutputPath := w.outputPathResolver.DefaultOutputPath(absoluteInputPath)

	return &PreparedDocument{
		InputPath:             absoluteInputPath,
		FileName:              filepath.Base(absoluteInputPath),
		DefaultOutputPath:     defaultOutputPath,
		DefaultOutputFilename: filepath.Base(defaultOutputPath),
		Directory:             filepath.Dir(absoluteInputPath),
	}, nil
}

func (w *EncryptionWorkflow) Encrypt(request EncryptRequest) (*EncryptResult, error) {
	preparedDocument, err := w.PrepareInput(request.InputPath)
	if err != nil {
		return nil, err
	}

	if err := w.passwordValidator.Validate(request.UserPassword); err != nil {
		return nil, err
	}

	resolvedOutputPath, err := w.outputPathResolver.ResolveOutputPath(
		preparedDocument.InputPath,
		request.OutputPath,
		request.AllowExistingOutputOverwrite,
		request.AllowInputOverwrite,
	)
	if err != nil {
		return nil, err
	}

	if err := w.fileValidator.ValidateOutputPDF(resolvedOutputPath); err != nil {
		return nil, err
	}

	encryptionResult, err := w.encryptionService.Encrypt(pdf.EncryptionRequest{
		InputPath:     preparedDocument.InputPath,
		OutputPath:    resolvedOutputPath,
		UserPassword:  request.UserPassword,
		OwnerPassword: request.OwnerPassword,
	})
	if err != nil {
		return nil, err
	}

	return &EncryptResult{
		OutputPath:     encryptionResult.OutputPath,
		OutputFilename: filepath.Base(encryptionResult.OutputPath),
		Message:        "Encrypted PDF saved successfully.",
	}, nil
}
