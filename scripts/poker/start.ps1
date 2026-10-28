$ErrorActionPreference = 'Stop'
Push-Location (Split-Path $PSScriptRoot -Parent)
try {
    docker compose up --build -d poker
    if ($LASTEXITCODE -ne 0) { throw 'Startup failed. Start Docker Desktop in Linux containers mode.' }
    Write-Host 'Build status: docker compose logs -f poker'
    Write-Host 'After [4/4] appears, open http://localhost:6080/vnc.html?autoconnect=1&resize=scale'
} finally { Pop-Location }
