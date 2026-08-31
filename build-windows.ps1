$ErrorActionPreference = 'Stop'
Push-Location "$PSScriptRoot\desktop"
try {
    go test ./...
    go build -o pdanet-host.exe .\cmd\pdanet-host
} finally { Pop-Location }
Write-Host "Built desktop\pdanet-host.exe"
