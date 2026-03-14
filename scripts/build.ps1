Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if (-not (Get-Command wails -ErrorAction SilentlyContinue)) {
    Write-Error "Wails CLI is required. Install it with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
}

go run ./scripts/syncicons
go test ./...
wails build -clean -tags webkit2_41
