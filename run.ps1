# run.ps1 — Load config from .env and run qagent
# Usage:  .\run.ps1 .\testdata\math\math.go [--max-heals 3] [--coverage] [--quiet]

param(
    [Parameter(Mandatory=$true)]
    [string]$File,
    [Parameter(ValueFromRemainingArguments=$true)]
    [string[]]$RemainingArgs
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

# Build the argument list
$args = @("--file", $File)
if ($RemainingArgs) {
    $args += $RemainingArgs
}

& .\bin\qagent.exe @args
