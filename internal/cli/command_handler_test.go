package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"

	"github.com/sp/pwpdf/internal/application"
	"github.com/sp/pwpdf/internal/files"
	"github.com/sp/pwpdf/internal/pdf"
	"github.com/sp/pwpdf/internal/platform"
	"github.com/sp/pwpdf/internal/validation"
	"github.com/sp/pwpdf/internal/version"
)

func TestNormalizeEncryptArgsAcceptsInputBeforeFlags(t *testing.T) {
	args := []string{"sample.pdf", "--password", "secret-123", "--output", "locked.pdf"}

	normalizedArgs, err := normalizeEncryptArgs(args)
	if err != nil {
		t.Fatalf("unexpected normalize error: %v", err)
	}

	expectedArgs := []string{"--password", "secret-123", "--output", "locked.pdf", "sample.pdf"}
	if !reflect.DeepEqual(normalizedArgs, expectedArgs) {
		t.Fatalf("expected %v, got %v", expectedArgs, normalizedArgs)
	}
}

func TestCommandHandlerRunEncryptAcceptsDocumentedFlagOrdering(t *testing.T) {
	api.DisableConfigDir()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "sample.pdf")
	outputPath := filepath.Join(tempDir, "sample_locked.pdf")
	writeMinimalPDF(t, inputPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	handler := NewCommandHandler(
		newTestWorkflow(),
		platform.NewShellIntegrationManager(),
		version.Info{Version: "test"},
		os.Stdin,
		&stdout,
		&stderr,
	)

	exitCode := handler.Run([]string{
		"encrypt",
		inputPath,
		"--password", "secret-123",
		"--output", outputPath,
	})
	if exitCode != 0 {
		t.Fatalf("expected success exit code, got %d with stderr %q", exitCode, stderr.String())
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected encrypted output to be created: %v", err)
	}

	if !bytes.Contains(stdout.Bytes(), []byte("Encrypted PDF saved successfully.")) {
		t.Fatalf("expected success message in stdout, got %q", stdout.String())
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
