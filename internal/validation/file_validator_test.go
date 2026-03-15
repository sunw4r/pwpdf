package validation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sunw4r/pwpdf/internal/validation"
)

func TestFileValidatorRejectsMissingInputFile(t *testing.T) {
	validator := validation.NewFileValidator()

	if err := validator.ValidateInputPDF("/tmp/does-not-exist.pdf"); err == nil {
		t.Fatal("expected an error for a missing input file")
	}
}

func TestFileValidatorRejectsWrongInputExtension(t *testing.T) {
	validator := validation.NewFileValidator()
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "notes.txt")

	if err := os.WriteFile(inputPath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("unable to create test file: %v", err)
	}

	if err := validator.ValidateInputPDF(inputPath); err == nil {
		t.Fatal("expected an error for a non-PDF input file")
	}
}
