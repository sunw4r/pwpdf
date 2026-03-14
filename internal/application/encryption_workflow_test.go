package application_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/sp/pwpdf/internal/application"
	"github.com/sp/pwpdf/internal/files"
	"github.com/sp/pwpdf/internal/pdf"
	"github.com/sp/pwpdf/internal/validation"
)

func TestEncryptionWorkflowEncryptsPDF(t *testing.T) {
	api.DisableConfigDir()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "sample.pdf")
	writeMinimalPDF(t, inputPath)

	workflow := newTestWorkflow()
	result, err := workflow.Encrypt(application.EncryptRequest{
		InputPath:    inputPath,
		UserPassword: "secret-123",
	})
	if err != nil {
		t.Fatalf("unexpected encryption error: %v", err)
	}

	if filepath.Base(result.OutputPath) != "sample_encrypted.pdf" {
		t.Fatalf("unexpected output filename: %s", filepath.Base(result.OutputPath))
	}

	if _, err := api.GetPermissionsFile(result.OutputPath, nil); err == nil {
		t.Fatal("expected the encrypted file to require a password")
	}

	verificationConfiguration := model.NewAESConfiguration("secret-123", "", 256)
	verificationConfiguration.Permissions = model.PermissionsAll
	if _, err := api.GetPermissionsFile(result.OutputPath, verificationConfiguration); err != nil {
		t.Fatalf("expected the encrypted file to open with the password: %v", err)
	}
}

func TestEncryptionWorkflowRejectsInputOverwriteWithoutPermission(t *testing.T) {
	api.DisableConfigDir()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "sample.pdf")
	writeMinimalPDF(t, inputPath)

	workflow := newTestWorkflow()
	_, err := workflow.Encrypt(application.EncryptRequest{
		InputPath:    inputPath,
		OutputPath:   inputPath,
		UserPassword: "secret-123",
	})
	if err == nil {
		t.Fatal("expected an error when trying to overwrite the input without permission")
	}
}

func newTestWorkflow() *application.EncryptionWorkflow {
	return application.NewEncryptionWorkflow(
		validation.NewPasswordValidator(),
		validation.NewFileValidator(),
		files.NewOutputPathResolver(),
		pdf.NewEncryptionService(),
	)
}

func writeMinimalPDF(t *testing.T, outputPath string) {
	t.Helper()

	stream := "BT\n/F1 18 Tf\n72 72 Td\n(Hello from pwpdf) Tj\nET\n"
	objects := []string{
		"1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n",
		"2 0 obj\n<< /Type /Pages /Count 1 /Kids [3 0 R] >>\nendobj\n",
		"3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 144] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n",
		fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", len(stream), stream),
		"5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n",
	}

	var buffer bytes.Buffer
	buffer.WriteString("%PDF-1.4\n")

	offsets := make([]int, len(objects)+1)
	for index, object := range objects {
		offsets[index+1] = buffer.Len()
		buffer.WriteString(object)
	}

	xrefOffset := buffer.Len()
	fmt.Fprintf(&buffer, "xref\n0 %d\n", len(objects)+1)
	buffer.WriteString("0000000000 65535 f \n")
	for index := 1; index <= len(objects); index++ {
		fmt.Fprintf(&buffer, "%010d 00000 n \n", offsets[index])
	}
	fmt.Fprintf(&buffer, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xrefOffset)

	if err := os.WriteFile(outputPath, buffer.Bytes(), 0o644); err != nil {
		t.Fatalf("unable to write minimal PDF: %v", err)
	}
}
