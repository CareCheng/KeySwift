# ==========================================
#      KeySwift 图片人机验证插件构建脚本
# ==========================================
#
# 用法:
#   .\build.ps1                 # 默认构建 Windows/amd64 并打包
#   .\build.ps1 -Linux          # 构建 Linux/amd64
#   .\build.ps1 -Mac            # 构建 macOS/amd64
#   .\build.ps1 -All -X64 -Arm  # 构建全部平台与架构
#   .\build.ps1 -Clean          # 清理构建产物
#   .\build.ps1 -SkipFrontend   # 跳过前端构建（无 Node/离线环境，仅构建后端二进制）
#
# 前端构建说明：
#   插件前端为独立 Next.js 工程（frontend/），与主程序同栈。
#   默认执行 npm ci 安装依赖并 next build 静态导出，产物落到 releases/<version>/frontend/widget.html。
#   -SkipFrontend 时跳过前端构建，保留既有 releases frontend 产物（若存在）。

param(
    [switch]$Windows,
    [switch]$Linux,
    [switch]$Mac,
    [switch]$All,
    [switch]$Arm,
    [switch]$X64,
    [switch]$Clean,
    [switch]$SkipPause,
    [switch]$SkipFrontend
)

$ErrorActionPreference = "Stop"
$ScriptPath = $PSScriptRoot
Set-Location $ScriptPath

function Write-Info { Write-Host "[INFO] $args" -ForegroundColor Cyan }
function Write-Success { Write-Host "[SUCCESS] $args" -ForegroundColor Green }
function Write-Warning { Write-Host "[WARNING] $args" -ForegroundColor Yellow }
function Write-Err { Write-Host "[ERROR] $args" -ForegroundColor Red }

$PluginID = "keyswift.image_captcha"
$Version = "1.0.0"
$BinaryBase = "keyswift-image-captcha"
$ReleaseDir = Join-Path $ScriptPath "releases\$Version"
$DistDir = Join-Path $ScriptPath "dist"
$PackageDir = Join-Path $DistDir "packages"

if ($Clean) {
    Write-Info "清理插件构建产物"
    if (Test-Path (Join-Path $ReleaseDir "bin")) { Remove-Item -Path (Join-Path $ReleaseDir "bin") -Recurse -Force }
    if (Test-Path (Join-Path $ReleaseDir "frontend")) { Remove-Item -Path (Join-Path $ReleaseDir "frontend") -Recurse -Force }
    if (Test-Path $DistDir) { Remove-Item -Path $DistDir -Recurse -Force }
    Set-Content -Path (Join-Path $ReleaseDir "checksums.json") -Value "{}" -NoNewline -Encoding UTF8
    Write-Success "清理完成"
    if (-not $SkipPause) { Read-Host "按回车退出" | Out-Null }
    exit 0
}

$TargetOS = @()
if ($All) {
    $TargetOS += "windows", "linux", "darwin"
} else {
    if ($Windows) { $TargetOS += "windows" }
    if ($Linux) { $TargetOS += "linux" }
    if ($Mac) { $TargetOS += "darwin" }
}
if ($TargetOS.Count -eq 0) { $TargetOS += "windows" }

$TargetArch = @()
if ($Arm) { $TargetArch += "arm64" }
if ($X64) { $TargetArch += "amd64" }
if ($TargetArch.Count -eq 0) { $TargetArch += "amd64" }

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  $PluginID 构建脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Err "未找到 Go 编译器"
    exit 1
}
Write-Info "Go: $(go version)"
Write-Info "整理插件依赖"
go mod tidy
if ($LASTEXITCODE -ne 0) { throw "go mod tidy 失败" }

foreach ($os in $TargetOS) {
    foreach ($arch in $TargetArch) {
        $PlatformDir = "${os}_${arch}"
        $BinaryName = $BinaryBase
        if ($os -eq "windows") { $BinaryName += ".exe" }
        $OutputDir = Join-Path $ReleaseDir "bin\$PlatformDir"
        $OutputFile = Join-Path $OutputDir $BinaryName

        Write-Info "构建 $os/$arch"
        if (Test-Path $OutputDir) { Remove-Item -Path $OutputDir -Recurse -Force }
        New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

        $env:GOOS = $os
        $env:GOARCH = $arch
        $env:CGO_ENABLED = "0"
        go build -ldflags="-s -w" -o $OutputFile ./src
        if ($LASTEXITCODE -ne 0) { throw "Go 编译失败: $os/$arch" }
        Write-Success "已生成: $OutputFile"
    }
}

Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue

Write-Info "构建插件前端（Next.js 静态导出）"
$FrontendSourceDir = Join-Path $ScriptPath "frontend"
$FrontendReleaseDir = Join-Path $ReleaseDir "frontend"
if (Test-Path $FrontendReleaseDir) { Remove-Item -Path $FrontendReleaseDir -Recurse -Force }
New-Item -ItemType Directory -Path $FrontendReleaseDir -Force | Out-Null

if ($SkipFrontend) {
    Write-Warning "已跳过前端构建（-SkipFrontend）；releases/frontend 将为空，需确保已有产物或后续补构建"
} elseif (-not (Test-Path (Join-Path $FrontendSourceDir "package.json"))) {
    Write-Err "未找到前端工程 frontend/package.json"
    exit 1
} else {
    if (-not (Get-Command npm -ErrorAction SilentlyContinue)) {
        Write-Err "未找到 npm，无法构建前端；可在具备 Node 环境时重新构建，或使用 -SkipFrontend 仅构建后端"
        exit 1
    }
    Write-Info "安装前端依赖（npm ci）"
    Push-Location $FrontendSourceDir
    try {
        npm ci
        if ($LASTEXITCODE -ne 0) { throw "npm ci 失败" }
        Write-Info "执行前端静态导出（next build）"
        npm run build
        if ($LASTEXITCODE -ne 0) { throw "next build 失败" }
    } finally {
        Pop-Location
    }

    # Next.js 静态导出默认产物在 frontend/out/，单页 widget 路由生成 out/widget.html
    $FrontendOutDir = Join-Path $FrontendSourceDir "out"
    $BuiltWidgetHtml = Join-Path $FrontendOutDir "widget.html"
    if (-not (Test-Path $BuiltWidgetHtml)) {
        Write-Err "未找到前端构建产物 out/widget.html"
        exit 1
    }
    # 主页面 widget.html 落到 releases/frontend/widget.html（manifest 与宿主 serve 端点均指向此路径）
    Copy-Item -Path $BuiltWidgetHtml -Destination (Join-Path $FrontendReleaseDir "widget.html") -Force
    # 仅拷贝 widget.html 引用的 _next 静态资源目录（JS/CSS），其余 Next 默认产物（404/_not-found）不纳入插件包
    $NextAssetDir = Join-Path $FrontendOutDir "_next"
    if (Test-Path $NextAssetDir) {
        Copy-Item -Path $NextAssetDir -Destination (Join-Path $FrontendReleaseDir "_next") -Recurse -Force
    }
    Write-Success "前端产物已输出到 $FrontendReleaseDir"
}

Write-Info "生成 checksums.json"
$checksums = [ordered]@{}
Get-ChildItem -Path $ReleaseDir -Recurse -File |
    Where-Object { $_.Name -ne "checksums.json" } |
    Sort-Object FullName |
    ForEach-Object {
        $relative = [System.IO.Path]::GetRelativePath($ReleaseDir, $_.FullName).Replace("\", "/")
        $checksums[$relative] = (Get-FileHash -Path $_.FullName -Algorithm SHA256).Hash.ToLowerInvariant()
    }
$checksums | ConvertTo-Json -Depth 8 | Set-Content -Path (Join-Path $ReleaseDir "checksums.json") -Encoding UTF8

Write-Info "生成可安装插件包"
New-Item -ItemType Directory -Path $PackageDir -Force | Out-Null
foreach ($os in $TargetOS) {
    foreach ($arch in $TargetArch) {
        $PackageName = "$PluginID-$Version-${os}_${arch}.ksplugin.zip"
        $PackagePath = Join-Path $PackageDir $PackageName
        $StageDir = Join-Path $DistDir "stage\${os}_${arch}"
        $StagePluginDir = Join-Path $StageDir $PluginID
        if (Test-Path $StageDir) { Remove-Item -Path $StageDir -Recurse -Force }
        New-Item -ItemType Directory -Path $StagePluginDir -Force | Out-Null
        Copy-Item -Path (Join-Path $ScriptPath "releases") -Destination $StagePluginDir -Recurse
        if (Test-Path $PackagePath) { Remove-Item -Path $PackagePath -Force }
        Compress-Archive -Path (Join-Path $StageDir "*") -DestinationPath $PackagePath -Force
        Write-Success "已打包: $PackagePath"
    }
}

Write-Success "插件构建完成"
if (-not $SkipPause) { Read-Host "按回车退出" | Out-Null }
