package validation

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileValidator struct{}

func NewFileValidator() *FileValidator {
	return &FileValidator{}
}

func (v *FileValidator) ValidateInputPDF(inputPath string) error {
	if strings.TrimSpace(inputPath) == "" {
		return errors.New("an input PDF is required")
	}

	if !hasPDFExtension(inputPath) {
		return errors.New("the input file must have a .pdf extension")
	}

	info, err := os.Stat(inputPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input PDF does not exist: %s", inputPath)
		}
		return fmt.Errorf("unable to access the input PDF: %w", err)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("input path is not a file: %s", inputPath)
	}

	return nil
}

func (v *FileValidator) ValidateOutputPDF(outputPath string) error {
	if strings.TrimSpace(outputPath) == "" {
		return errors.New("an output path is required")
	}

	if !hasPDFExtension(outputPath) {
		return errors.New("the output file must have a .pdf extension")
	}

	outputDirectory := filepath.Dir(outputPath)
	info, err := os.Stat(outputDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("output directory does not exist: %s", outputDirectory)
		}
		return fmt.Errorf("unable to access the output directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("output directory is not a directory: %s", outputDirectory)
	}

	targetInfo, err := os.Stat(outputPath)
	if err == nil && !targetInfo.Mode().IsRegular() {
		return fmt.Errorf("output path is not a regular file: %s", outputPath)
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to inspect the output path: %w", err)
	}

	return nil
}

func hasPDFExtension(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".pdf")
}
