# Architecture

## Overview

`pwpdf` is organized around one shared encryption workflow that is reused by both the CLI and the Wails desktop app.

The goal is simple:

- validate the input PDF
- validate passwords and output paths
- encrypt with `pdfcpu`
- verify the output can only be opened with the password

## Major Components

### `internal/application`

Contains the shared orchestration layer.

- `EncryptionWorkflow.PrepareInput` validates the selected file and computes the default output name
- `EncryptionWorkflow.Encrypt` applies password validation, output-path resolution, and the actual PDF encryption service

### `internal/pdf`

Wraps `pdfcpu`.

- validates that the input file is a readable, unencrypted PDF
- encrypts using AES-256
- validates the encrypted output
- verifies the output cannot be opened without a password

### `internal/cli`

Parses commands and flags, prompts securely for the password when needed, and calls the shared workflow.

### `internal/gui`

Hosts the Wails backend API exposed to the frontend.

- opens native file and save dialogs
- receives drag-and-drop or shell-opened file paths
- validates password confirmation for the GUI flow
- calls the shared workflow

### `frontend`

Vanilla TypeScript UI focused on a compact, native-feeling desktop flow.

- drag-and-drop target powered by Wails runtime hooks
- small password form
- success and error states

### `internal/platform`

Contains minimal shell integration helpers.

- Windows: per-user context menu entry for PDF files
- Linux: `.desktop` entry for `Open With`
- macOS: documented fallback only

## Flow

### CLI

1. `main.go` detects CLI mode from the arguments
2. `internal/cli` parses the command
3. `EncryptionWorkflow` prepares and encrypts the PDF
4. the CLI prints the result and exits with a status code

### GUI

1. `main.go` launches Wails
2. the frontend selects or receives a PDF path
3. `internal/gui` validates the selection and opens the native save dialog
4. `EncryptionWorkflow` encrypts the file
5. the frontend renders the success or error state

## Why `pdfcpu`

`pdfcpu` works directly on PDF structure and supports standard PDF encryption. That means `pwpdf` can protect the file without rasterizing pages or reconstructing the document visually.

## Why Wails

Wails keeps the project in Go for backend logic while still providing a lightweight desktop window, native dialogs, drag-and-drop support, and straightforward packaging compared with heavier desktop frameworks.
