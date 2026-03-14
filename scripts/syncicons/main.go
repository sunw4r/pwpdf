package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/leaanthony/winicon"
)

func main() {
	projectRoot, err := os.Getwd()
	if err != nil {
		fail("resolve project root", err)
	}

	sourcePath := filepath.Join(projectRoot, "build", "appicon.png")
	targetPath := filepath.Join(projectRoot, "build", "windows", "icon.ico")

	sourceBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		fail("read appicon.png", err)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		fail("create icon output directory", err)
	}

	targetFile, err := os.Create(targetPath)
	if err != nil {
		fail("create icon.ico", err)
	}
	defer targetFile.Close()

	if err := winicon.GenerateIcon(bytes.NewReader(sourceBytes), targetFile, []int{256, 128, 64, 48, 32, 16}); err != nil {
		fail("generate icon.ico", err)
	}
}

func fail(action string, err error) {
	fmt.Fprintf(os.Stderr, "unable to %s: %v\n", action, err)
	os.Exit(1)
}
