package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type OutputPathResolver struct{}

func NewOutputPathResolver() *OutputPathResolver {
	return &OutputPathResolver{}
}

func (r *OutputPathResolver) DefaultOutputPath(inputPath string) string {
	fileExtension := filepath.Ext(inputPath)
	fileNameWithoutExtension := strings.TrimSuffix(filepath.Base(inputPath), fileExtension)
	return filepath.Join(filepath.Dir(inputPath), fileNameWithoutExtension+"_encrypted.pdf")
}

func (r *OutputPathResolver) ResolveOutputPath(
	inputPath string,
	requestedOutputPath string,
	allowExistingOutputOverwrite bool,
	allowInputOverwrite bool,
) (string, error) {
	outputPath := requestedOutputPath
	if strings.TrimSpace(outputPath) == "" {
		outputPath = r.DefaultOutputPath(inputPath)
	}

	outputPath = appendPDFExtensionIfMissing(outputPath)

	absoluteInputPath, err := filepath.Abs(inputPath)
	if err != nil {
		return "", fmt.Errorf("unable to resolve the input path: %w", err)
	}

	absoluteOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return "", fmt.Errorf("unable to resolve the output path: %w", err)
	}

	isSamePath := pathsReferToSameLocation(absoluteInputPath, absoluteOutputPath)
	if isSamePath && !allowInputOverwrite {
		return "", errors.New("refusing to overwrite the input PDF; choose a different output path")
	}

	if !isSamePath {
		if _, err := os.Stat(absoluteOutputPath); err == nil && !allowExistingOutputOverwrite {
			return "", fmt.Errorf("the output file already exists: %s", absoluteOutputPath)
		} else if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("unable to inspect the output path: %w", err)
		}
	}

	return absoluteOutputPath, nil
}

func appendPDFExtensionIfMissing(path string) string {
	if filepath.Ext(path) == "" {
		return path + ".pdf"
	}
	return path
}

func pathsReferToSameLocation(left, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)

	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}

	return left == right
}
