# pwpdf

![Go](https://img.shields.io/badge/go-1.26%2B-00ADD8?logo=go&logoColor=white)
![Wails](https://img.shields.io/badge/gui-wails-v2-ff875f)
![pdfcpu](https://img.shields.io/badge/pdf-pdfcpu-1f2937)
![License](https://img.shields.io/badge/license-MIT-22c55e)

**The Overkill PDF Password Protector**

`pwpdf` is a compact desktop + CLI utility that adds password protection to existing PDF files without rebuilding pages, rasterizing content, or touching the original document unless you explicitly ask for overwrite behavior in the CLI.

It is built in Go, uses `pdfcpu` for encryption, ships with a Wails desktop app, and leans into a compact old-school desktop aesthetic instead of pretending to be a web app in a window.

## Preview

Drop your screenshots and gifs under `docs/images/` and reference them with relative paths.

Recommended pattern:

```md
![Main Window](docs/images/main-window.png)
![Drag and Drop Flow](docs/images/drag-drop.gif)
```

That is the simplest and most portable approach for GitHub READMEs. It keeps assets versioned with the repo and avoids broken external links later.

## What It Does

- Encrypts existing PDF files with `pdfcpu`
- Preserves layout and appearance by modifying PDF structure instead of recreating pages
- Supports both GUI and CLI workflows
- Accepts drag-and-drop in the desktop app
- Uses native open/save dialogs in GUI mode
- Prompts securely for a password in the terminal when needed
- Defaults output names to `*_encrypted.pdf`
- Includes optional shell integration helpers for Windows and Linux

## Why It Exists

Most people do not need a full PDF suite to put a password on one file.

They need:

- pick a PDF
- type a password
- save the protected copy
- move on

That is the whole point of `pwpdf`.

## How It Works

`pwpdf` shares one encryption workflow between the CLI and the GUI:

1. Validate that the input exists and is really a PDF
2. Resolve the default output path
3. Validate the password
4. Encrypt the file with `pdfcpu` using standard PDF encryption
5. Validate that the resulting file exists and is actually protected

Important detail: the app does **not** rasterize pages, flatten the document into images, or rebuild layout manually. The goal is to keep the document visually identical and only add password protection.

## GUI

The desktop app is built with Wails and is intentionally focused:

1. Drop a PDF into the window or click `Select PDF`
2. Enter the password twice
3. Choose where the encrypted copy should be saved
4. Done

GUI highlights:

- drag and drop
- native file picker
- native save dialog
- password confirmation
- preloaded file support when the app is launched with a PDF path
- compact layout that adapts for smaller screens

You can also launch the GUI with a file already selected:

```bash
go run . /path/to/file.pdf
```

## CLI

Basic usage:

```bash
pwpdf encrypt input.pdf
pwpdf encrypt input.pdf --password "MySecret123"
pwpdf encrypt input.pdf --password "MySecret123" --output output.pdf
pwpdf encrypt input.pdf --password "MySecret123" --output input.pdf --overwrite
```

You can run it from source as well:

```bash
go run . encrypt input.pdf
go run . encrypt input.pdf --password "MySecret123"
```

Supported commands:

```text
pwpdf
pwpdf /path/to/file.pdf
pwpdf encrypt input.pdf [options]
pwpdf install-shell-integration
pwpdf uninstall-shell-integration
pwpdf version
```

`encrypt` options:

- `--password`, `-p`: user password
- `--output`, `-o`: output file path
- `--owner-password`: optional owner password
- `--overwrite`: allow overwriting the output file, and the input file if you point output back to it

If `--password` is omitted, the CLI prompts for it securely in the terminal.

## Shell Integration

### Windows

Install a right-click context-menu entry for PDF files:

```bash
pwpdf install-shell-integration
```

That adds `Protect PDF with pwpdf` and opens the GUI with the selected PDF.

### Linux

Install a local `.desktop` entry so the app can appear in `Open With` flows:

```bash
pwpdf install-shell-integration
```

### macOS

There is no invasive Finder integration step. The practical path is:

- `Open With`
- drag a PDF onto the app
- launch the app with a file path

## Build

### Requirements

- Go `1.26.1` or newer
- Node.js `18+`
- Wails CLI
- On macOS: Xcode Command Line Tools

Install Wails CLI:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Linux desktop builds also need the GTK/WebKit dependencies required by Wails.

On macOS, install the Apple toolchain with:

```bash
xcode-select --install
```

Then verify the local environment with:

```bash
wails doctor
```

### Development

```bash
make dev
```

or:

```bash
./scripts/dev.sh
```

### Tests

```bash
make test
```

### Release Build

```bash
make build
```

or:

```bash
./scripts/build.sh
```

PowerShell:

```powershell
.\scripts\build.ps1
```

On macOS, the direct local build path is:

```bash
./scripts/build.sh
scripts/build-all.sh --selector macos
```

macOS builds export both artifacts:

- `pwpdf.app`
- `pwpdf` as a standalone CLI binary copied out of `pwpdf.app/Contents/MacOS/pwpdf`

## Multi-Platform Builds

There is a matrix helper for release builds:

```bash
scripts/build-all.sh --selector all
make build-all SELECTOR=all
```

Artifacts are written to `dist/<platform>/`.

Supported selectors:

- `all`
- `linux`
- `windows`
- `macos`
- exact targets like `linux/amd64`, `linux/arm64`, `windows/amd64`, `windows/arm64`, `darwin/universal`

Notes:

- On macOS, use `scripts/build-all.sh --selector macos` or `scripts/build-all.sh --selector darwin/universal` to build the local `.app` and a standalone `pwpdf` CLI binary without Docker
- Linux foreign-architecture builds can fall back to Docker automatically
- Windows targets can be built through Wails from Linux in this repository workflow
- macOS desktop packaging still needs a macOS machine or runner for the real `.app` output

## Project Layout

```text
.
├── build/
├── docs/
├── frontend/
├── internal/
│   ├── application/
│   ├── cli/
│   ├── files/
│   ├── gui/
│   ├── pdf/
│   ├── platform/
│   ├── validation/
│   └── version/
├── scripts/
├── main.go
└── wails.json
```

If you want the architecture write-up instead of the marketing version, read [docs/architecture.md](docs/architecture.md).

## Security Notes

- `pwpdf` uses standard PDF encryption via `pdfcpu`
- Viewer behavior still depends on the PDF security model being respected by the PDF reader
- PDF passwords are useful, but they are not a substitute for storage-layer or endpoint security controls

## Limitations

- It protects existing PDFs; it is not a PDF editor
- Already encrypted PDFs need to be decrypted before being protected again
- Shell integration is intentionally minimal and optional
- macOS packaging still requires a macOS environment

## FAQ

### Does it preserve the original PDF layout?

Yes. That is one of the core requirements of the project. The app does not recreate pages as images and does not intentionally alter fonts, page sizes, or content positioning.

### Can it overwrite the original file?

Not by default.

- GUI mode always writes a new file
- CLI mode writes a new file unless you explicitly opt into overwrite behavior

### Why both GUI and CLI?

Because both are useful:

- GUI for normal users
- CLI for automation, terminal workflows, shell integration, and quick one-off usage

## License

MIT. See [LICENSE](LICENSE).
