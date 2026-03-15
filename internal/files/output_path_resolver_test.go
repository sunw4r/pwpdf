package files_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sunw4r/pwpdf/internal/files"
)

func TestOutputPathResolverBuildsDefaultName(t *testing.T) {
	resolver := files.NewOutputPathResolver()

	outputPath := resolver.DefaultOutputPath("/tmp/report.pdf")
	expectedPath := filepath.Join("/tmp", "report_encrypted.pdf")

	if outputPath != expectedPath {
		t.Fatalf("expected %q, got %q", expectedPath, outputPath)
	}
}

func TestOutputPathResolverRejectsExistingOutputWithoutOverwrite(t *testing.T) {
	resolver := files.NewOutputPathResolver()
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "sample.pdf")
	outputPath := filepath.Join(tempDir, "existing.pdf")

	if err := os.WriteFile(outputPath, []byte("data"), 0o644); err != nil {
		t.Fatalf("unable to create existing output file: %v", err)
	}

	if _, err := resolver.ResolveOutputPath(inputPath, outputPath, false, false); err == nil {
		t.Fatal("expected an error when the output file already exists")
	}
}

func TestOutputPathResolverAppendsPDFExtension(t *testing.T) {
	resolver := files.NewOutputPathResolver()
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "sample.pdf")
	outputPath := filepath.Join(tempDir, "secured")

	resolvedPath, err := resolver.ResolveOutputPath(inputPath, outputPath, false, false)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}

	if filepath.Ext(resolvedPath) != ".pdf" {
		t.Fatalf("expected output path to end with .pdf, got %q", resolvedPath)
	}
}
