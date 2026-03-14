package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

type TerminalPrompter struct {
	stdin  *os.File
	output io.Writer
}

func NewTerminalPrompter(stdin *os.File, output io.Writer) *TerminalPrompter {
	return &TerminalPrompter{
		stdin:  stdin,
		output: output,
	}
}

func (p *TerminalPrompter) PromptPassword(prompt string) (string, error) {
	if p.stdin == nil {
		return "", errors.New("standard input is unavailable")
	}

	if !term.IsTerminal(int(p.stdin.Fd())) {
		return "", errors.New("password flag is required when standard input is not interactive")
	}

	if _, err := fmt.Fprint(p.output, prompt); err != nil {
		return "", err
	}

	passwordBytes, err := term.ReadPassword(int(p.stdin.Fd()))
	if _, writeErr := fmt.Fprintln(p.output); writeErr != nil && err == nil {
		err = writeErr
	}
	if err != nil {
		return "", fmt.Errorf("unable to read password: %w", err)
	}

	return string(passwordBytes), nil
}
