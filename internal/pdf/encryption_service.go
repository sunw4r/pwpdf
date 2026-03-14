package pdf

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdfcpucore "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type EncryptionService struct{}

func NewEncryptionService() *EncryptionService {
	return &EncryptionService{}
}

func (s *EncryptionService) ValidateInputPDF(inputPath string) error {
	if _, err := api.ReadContextFile(inputPath); err != nil {
		return s.normalizePDFError(err)
	}

	return nil
}

func (s *EncryptionService) Encrypt(request EncryptionRequest) (*EncryptionResult, error) {
	encryptionConfiguration := model.NewAESConfiguration(
		request.UserPassword,
		effectiveOwnerPassword(request),
		256,
	)
	encryptionConfiguration.Permissions = model.PermissionsAll

	if err := api.EncryptFile(request.InputPath, request.OutputPath, encryptionConfiguration); err != nil {
		return nil, fmt.Errorf("unable to encrypt the PDF: %w", s.normalizePDFError(err))
	}

	verificationConfiguration := model.NewAESConfiguration(request.UserPassword, "", 256)
	verificationConfiguration.Permissions = model.PermissionsAll

	if err := api.ValidateFile(request.OutputPath, verificationConfiguration); err != nil {
		return nil, fmt.Errorf("the encrypted PDF could not be validated: %w", s.normalizePDFError(err))
	}

	if _, err := api.GetPermissionsFile(request.OutputPath, nil); err == nil {
		return nil, errors.New("encryption verification failed because the output PDF can still be opened without a password")
	}

	if _, err := api.GetPermissionsFile(request.OutputPath, verificationConfiguration); err != nil {
		return nil, fmt.Errorf("encryption verification failed: %w", s.normalizePDFError(err))
	}

	outputInfo, err := os.Stat(request.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("unable to inspect the encrypted PDF: %w", err)
	}

	return &EncryptionResult{
		OutputPath: request.OutputPath,
		Algorithm:  "AES-256",
		SizeBytes:  outputInfo.Size(),
	}, nil
}

func (s *EncryptionService) normalizePDFError(err error) error {
	if errors.Is(err, pdfcpucore.ErrWrongPassword) || strings.Contains(err.Error(), "correct password") {
		return errors.New("the selected PDF is already password protected")
	}

	return err
}

func effectiveOwnerPassword(request EncryptionRequest) string {
	if request.OwnerPassword != "" {
		return request.OwnerPassword
	}

	return request.UserPassword
}
