#Requires -RunAsAdministrator

param(
    [switch]$Force
)

$ErrorActionPreference = "Stop"
$InstallDir = "$env:ProgramFiles\smtp-cli"
$Binary = "$InstallDir\smtp-cli.exe"

Write-Host ""
Write-Host "=== smtp-cli Installer ===" -ForegroundColor Cyan
Write-Host ""

# Step 1: Check Go is available
Write-Host "[1/4] Checking Go compiler..." -ForegroundColor Yellow
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: Go is not installed or not in PATH." -ForegroundColor Red
    Write-Host "Install Go first: https://go.dev/dl/" -ForegroundColor Red
    exit 1
}
$goVersion = go version
Write-Host "  $goVersion" -ForegroundColor Green

# Step 2: Build binary
Write-Host "[2/4] Building smtp-cli..." -ForegroundColor Yellow
if ((Test-Path $Binary) -and -not $Force) {
    Write-Host "  Binary already exists (use -Force to rebuild)." -ForegroundColor Green
} else {
    if ($Force) {
        Write-Host "  Force rebuild requested..." -ForegroundColor Yellow
    }
    $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
    Push-Location $scriptDir
    try {
        go build -o $Binary ./cmd/smtp-cli
        if ($LASTEXITCODE -ne 0) {
            Write-Host "ERROR: Build failed." -ForegroundColor Red
            exit 1
        }
        Write-Host "  Built successfully: $Binary" -ForegroundColor Green
    } finally {
        Pop-Location
    }
}

# Step 3: Verify binary works
Write-Host "[3/4] Verifying binary..." -ForegroundColor Yellow
$helpOutput = & $Binary help 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Binary doesn't run correctly." -ForegroundColor Red
    Write-Host $helpOutput -ForegroundColor Red
    exit 1
}
Write-Host "  Binary is working." -ForegroundColor Green

# Step 4: Add to PATH
Write-Host "[4/4] Configuring system PATH..." -ForegroundColor Yellow
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
if ($CurrentPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", "Machine")
    Write-Host "  Added $InstallDir to system PATH." -ForegroundColor Green
    Write-Host "  NOTE: Restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
} else {
    Write-Host "  Already in system PATH." -ForegroundColor Green
}

# Done
Write-Host ""
Write-Host "=== Installation Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Usage:" -ForegroundColor Cyan
Write-Host "  smtp-cli help"
Write-Host "  smtp-cli send --to user@example.com --subject ""Hello"" --body ""Message"""
Write-Host ""
