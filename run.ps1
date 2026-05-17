# run.ps1 — Load config from .env and run qagent
# Usage:  .\run.ps1 .\testdata\math\math.go

param(
    [Parameter(Mandatory=$true)]
    [string]$File,
    [int]$MaxHeals = 2
)

# Load all QAGENT_* vars from .env into this session
if (Test-Path .env) {
    Get-Content .env | Where-Object { $_ -match "^\w" } | ForEach-Object {
        $parts = $_.Split("=", 2)
        Set-Item -Path "Env:$($parts[0])" -Value $parts[1]
    }
} else {
    Write-Error ".env file not found."
    exit 1
}

.\bin\qagent.exe --file $File --max-heals $MaxHeals
