# Telemetry Collection Script for QAgent
# Runs 50+ iterations across diverse test corpus
# Collects data to ~/.qagent/runs.jsonl for analysis

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  QAgent Telemetry Collection Script" -ForegroundColor Cyan
Write-Host "  Phase 4: Heal Success Rate Analysis" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

$totalRuns = 0

# Get all Go test files from testdata (excluding _test.go files)
$testFiles = @(
    "testdata/graph/poland_ball.go",
    "testdata/tree/party.go",
    "testdata/dp/fibonacci_mod.go",
    "testdata/buggy/buggy_math.go",
    "testdata/parser/string_parser.go",
    "testdata/calculator/complex_calc.go",
    "testdata/collections/ops.go",
    "testdata/validator/input.go",
    "testdata/statemachine/state.go",
    "testdata/math/math.go",
    "testdata/strutil/strings_util.go",
    "testdata/httpclient/http_client.go"
)

Write-Host "Test Files to Run:" -ForegroundColor Yellow
$testFiles | ForEach-Object { Write-Host "  - $_" }
Write-Host ""

Write-Host "Configuration:" -ForegroundColor Yellow
Write-Host "  - Iterations per file: 5-7"
Write-Host "  - Total expected runs: 60-84"
Write-Host '  - Coverage tracking: ENABLED (--coverage)'
Write-Host "  - Output: ~/.qagent/runs.jsonl"
Write-Host ""

$startTime = Get-Date
Write-Host "Starting telemetry collection at $($startTime.ToString('HH:mm:ss'))" -ForegroundColor Cyan
Write-Host ""

# Run iterations
foreach ($file in $testFiles) {
    Write-Host "-------------------------------------------" -ForegroundColor Cyan
    Write-Host "File: $file" -ForegroundColor Green
    Write-Host "-------------------------------------------" -ForegroundColor Cyan
    
    # Run 5-7 iterations per file (randomly chosen)
    $iterations = Get-Random -Minimum 5 -Maximum 8
    
    for ($i = 1; $i -le $iterations; $i++) {
        Write-Host "  [Iteration $i/$iterations]" -ForegroundColor Cyan
        
        # Run QAgent with coverage flag
        & .\run.ps1 $file --coverage 2>&1 | Out-Null
        
        $totalRuns++
        Write-Host "    Run #$totalRuns completed" -ForegroundColor Gray
        
        # Give system a brief pause between runs
        Start-Sleep -Milliseconds 500
    }
    
    Write-Host ""
}

$endTime = Get-Date
$duration = $endTime - $startTime

Write-Host ""
Write-Host "============================================" -ForegroundColor Green
Write-Host "  Telemetry Collection Complete!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""

Write-Host "Summary:" -ForegroundColor Yellow
Write-Host "  - Total Runs: $totalRuns"
Write-Host "  - Duration: $($duration.TotalSeconds.ToString('F1')) seconds"
Write-Host "  - Output Location: ~/.qagent/runs.jsonl"
Write-Host ""

Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host "  1. Telemetry has been collected to ~/.qagent/runs.jsonl"
Write-Host "  2. Run this command to view raw data:"
Write-Host "     cat ~/.qagent/runs.jsonl | jq ."
Write-Host "  3. Analyze error distribution:"
Write-Host "     cat ~/.qagent/runs.jsonl | jq '.error_type' | sort | uniq -c"
Write-Host "  4. Share the runs.jsonl file for analysis"
Write-Host ""

Write-Host "Telemetry collection finished at $($endTime.ToString('yyyy-MM-dd HH:mm:ss'))" -ForegroundColor Gray
