Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if (-not (Get-Command wails -ErrorAction SilentlyContinue)) {
    Write-Error "Wails CLI is required. Install it with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
}

$buildArgs = @("build", "-clean")
$isLinux = [System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Linux)

if ($isLinux) {
    $pkgConfig = Get-Command pkg-config -ErrorAction SilentlyContinue
    if ($pkgConfig) {
        & $pkgConfig.Source --exists webkit2gtk-4.1
        if ($LASTEXITCODE -eq 0) {
            $buildArgs += @("-tags", "webkit2_41")
        }
    }
}

go run ./scripts/syncicons
go test ./...
wails @buildArgs
