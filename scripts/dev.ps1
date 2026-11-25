#!/usr/bin/env pwsh
# 用 Docker 跑 Go 工具链，避免本机安装 Go 或被应用控制策略拦截。
# 用法:
#   .\scripts\dev.ps1 test
#   .\scripts\dev.ps1 build
#   .\scripts\dev.ps1 paipan -summary -coins 111,110,100,000,111,100
#   .\scripts\dev.ps1 paipan -lines "少阳,少阴,老阳,少阴,少阳,老阴"
#   .\scripts\dev.ps1 calendar -date 2026-11-25 -time 14:30 -lon 113.26
#   .\scripts\dev.ps1 shell

param(
    [ValidateSet("test", "build", "paipan", "calendar", "shell")]
    [string]$Task = "test",
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Rest
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$image = "golang:1.25-alpine"
$buildScript = "cd /src/apps/api && mkdir -p /src/bin && CGO_ENABLED=0 go build -trimpath -o /src/bin/mingyun ./cmd/mingyun"

$docker = "docker"
if (-not (Get-Command $docker -ErrorAction SilentlyContinue)) {
    $fallback = Join-Path $env:LOCALAPPDATA "Programs\DockerDesktop\resources\bin\docker.exe"
    if (Test-Path $fallback) { $docker = $fallback } else { throw "找不到 docker 命令" }
}

$runArgs = @(
    "run", "--rm",
    "-e", "GOPROXY=https://goproxy.cn,direct",
    "-e", "GOTOOLCHAIN=local",
    "-v", "${root}:/src",
    "-v", "mingyun-gomod:/go/pkg/mod"
)

function Invoke-Build {
    & $docker @runArgs -w /src $image sh -c $buildScript
    if ($LASTEXITCODE -ne 0) { throw "构建失败" }
}

# PowerShell 会把 "a,b,c" 拆成三个参数，这里把 -lines / -coins 后面的值重新接回逗号串。
function Merge-ListArgs([string[]]$InputArgs) {
    $out = New-Object System.Collections.Generic.List[string]
    $i = 0
    while ($i -lt $InputArgs.Count) {
        $a = $InputArgs[$i]
        if ($a -eq "-lines" -or $a -eq "-coins") {
            $vals = New-Object System.Collections.Generic.List[string]
            $j = $i + 1
            while ($j -lt $InputArgs.Count -and -not $InputArgs[$j].StartsWith("-")) {
                $vals.Add($InputArgs[$j])
                $j++
            }
            $out.Add($a)
            $out.Add(($vals -join ","))
            $i = $j
            continue
        }
        $out.Add($a)
        $i++
    }
    return $out.ToArray()
}

switch ($Task) {
    "test" {
        & $docker @runArgs -w /src $image sh -c "cd /src/paipan && go vet ./... && go test -count=1 ./... && cd /src/apps/api && go vet ./... && go test -count=1 ./..."
        exit $LASTEXITCODE
    }
    "build" {
        Invoke-Build
        Write-Host "已生成 bin/mingyun"
        exit 0
    }
    "paipan" {
        Invoke-Build
        $merged = Merge-ListArgs $Rest
        & $docker @runArgs --entrypoint /src/bin/mingyun $image paipan @merged
        exit $LASTEXITCODE
    }
    "calendar" {
        Invoke-Build
        $merged = Merge-ListArgs $Rest
        & $docker @runArgs --entrypoint /src/bin/mingyun $image calendar @merged
        exit $LASTEXITCODE
    }
    "shell" {
        & $docker @runArgs -it -w /src $image sh
        exit $LASTEXITCODE
    }
}
