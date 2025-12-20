# Go Import Cycle Diagnostic Tool (PowerShell)

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Go Import Cycle Diagnostic Tool" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# 清理缓存
Write-Host "1. Cleaning Go cache..." -ForegroundColor Yellow
go clean -cache
Write-Host "   ✓ Cache cleaned" -ForegroundColor Green
Write-Host ""

# 检查导入关系
Write-Host "2. Checking import relationships..." -ForegroundColor Yellow
Write-Host ""

Write-Host "   Checking middleware imports:" -ForegroundColor White
Get-Content internal\middleware\middleware.go | Select-String -Pattern "^import" -Context 0,10 | Select-Object -First 1
Write-Host ""

Write-Host "   Checking if service imports middleware:" -ForegroundColor White
$serviceImportsMiddleware = Get-ChildItem internal\service\*.go | Select-String -Pattern "dashgo/internal/middleware"
if ($serviceImportsMiddleware) {
    Write-Host "   ❌ FOUND: service imports middleware (this causes cycle!)" -ForegroundColor Red
    $serviceImportsMiddleware
} else {
    Write-Host "   ✓ OK: service does not import middleware" -ForegroundColor Green
}
Write-Host ""

Write-Host "   Checking if service imports handler:" -ForegroundColor White
$serviceImportsHandler = Get-ChildItem internal\service\*.go | Select-String -Pattern "dashgo/internal/handler"
if ($serviceImportsHandler) {
    Write-Host "   ❌ FOUND: service imports handler (this causes cycle!)" -ForegroundColor Red
    $serviceImportsHandler
} else {
    Write-Host "   ✓ OK: service does not import handler" -ForegroundColor Green
}
Write-Host ""

# 尝试编译
Write-Host "3. Attempting to compile..." -ForegroundColor Yellow
Write-Host ""

$compileOutput = go build -o $env:TEMP\dashgo-server.exe .\cmd\server 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✓ Compilation successful!" -ForegroundColor Green
    Remove-Item $env:TEMP\dashgo-server.exe -ErrorAction SilentlyContinue
} else {
    Write-Host ""
    Write-Host "❌ Compilation failed. Error output:" -ForegroundColor Red
    Write-Host "-----------------------------------" -ForegroundColor Red
    $compileOutput
    Write-Host "-----------------------------------" -ForegroundColor Red
    Write-Host ""
    
    # 检查是否是循环导入错误
    if ($compileOutput -match "import cycle") {
        Write-Host "Detected import cycle. Analyzing..." -ForegroundColor Yellow
        Write-Host ""
        $compileOutput | Select-String -Pattern "import cycle"
    }
}

Write-Host ""
Write-Host "4. Checking AuthService implementation..." -ForegroundColor Yellow
$hasGetUserFromToken = Get-Content internal\service\auth.go | Select-String -Pattern "func \(s \*AuthService\) GetUserFromToken"
if ($hasGetUserFromToken) {
    Write-Host "   ✓ AuthService has GetUserFromToken method" -ForegroundColor Green
} else {
    Write-Host "   ❌ AuthService missing GetUserFromToken method" -ForegroundColor Red
}

Write-Host ""
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Diagnostic complete" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
