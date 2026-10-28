$ErrorActionPreference = 'Stop'
Push-Location (Split-Path $PSScriptRoot -Parent)
try {
    docker compose --profile tools run --rm -e CGO_ENABLED=0 tools wails build -s -platform windows/amd64
    if ($LASTEXITCODE -ne 0) { throw 'Windows build failed. Check Docker Desktop and the build log.' }
    Write-Host 'Created: build/bin/poker-lab-lounge.exe'
} finally { Pop-Location }
