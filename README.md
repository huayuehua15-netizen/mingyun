# 命云 MingYun

[![CI](https://github.com/huayuehua15-netizen/mingyun/actions/workflows/ci.yml/badge.svg)](https://github.com/huayuehua15-netizen/mingyun/actions/workflows/ci.yml)

开源的 AI 术数排盘平台。支持六爻、大六壬、紫微斗数三种起盘方式，盘面由确定性引擎计算，解读由可自主配置的大模型完成。

**技术栈**：Vue 3 + TypeScript / Go 1.25 / PostgreSQL / Redis / Docker Compose

## 当前状态

M1 进行中。历法层与六爻引擎 v0.1.0 已完成并在 Docker 内验证通过，HTTP 服务仍是骨架。

| 模块 | 状态 |
|---|---|
| `paipan/calendar` 历法层（干支、日柱、旬空、时辰、真太阳时、节气、月将） | ✅ 可用，测试通过 |
| `paipan/schema` 数据契约 | ✅ 可用 |
| `paipan/liuyao` 六爻（八宫、纳甲、六亲、六神、世应、伏神、旺衰、变爻回头生克） | ✅ v0.1.0，测试通过 |
| `apps/api` 单二进制（`serve` / `calendar` / `paipan` 子命令） | ✅ 可跑 |
| HTTP API：排盘、盘面详情、列表、历法、能力清单 | ✅ 可跑（内存存储） |
| `Dockerfile` 多阶段构建（22.9MB，健康检查通过） | ✅ 可构建 |
| 摇卦熵源、时间起卦 | ⏳ M1 收尾 |
| `paipan/daliuren` 大六壬 | ⏳ M3 |
| `paipan/ziwei` 紫微斗数 | ⏳ M4 |
| `apps/web` 前端 | ⏳ M2 |

## 目录

| 路径 | 说明 |
|---|---|
| `paipan/` | 排盘引擎，独立 Go module，可被第三方 `go get` |
| `paipan/calendar/` | 三术共用的历法计算 |
| `paipan/liuyao/` | 六爻装卦与排盘 |
| `paipan/schema/` | Chart / Fact / Method 等契约类型 |
| `apps/api/` | Go 服务，单二进制含 `serve` / `calendar` / `paipan` 子命令 |
| `scripts/` | 用 Docker 跑 Go 工具链的封装脚本 |
| `docs/` | 开发文档（先读 `docs/README.md`） |

## 快速开始

### 方式一：Docker（推荐，不污染本机）

需要 Docker Desktop。脚本会自动拉取 `golang:1.25-alpine`，并把模块缓存挂到 `mingyun-gomod` 卷里。

```
# Windows
pwsh .\scripts\dev.ps1 test

# Linux / macOS
./scripts/dev.sh test
```

排一张六爻盘（铜钱：1 为背，自下而上六组）：

```
pwsh .\scripts\dev.ps1 paipan -summary -coins 111,110,100,000,111,100
```

手动录入爻形（自下而上：少阳 / 少阴 / 老阳 / 老阴）：

```
pwsh .\scripts\dev.ps1 paipan -lines "少阳,少阴,老阳,少阴,少阳,老阴"
```

加 `-question` 带上所问之事，去掉 `-summary` 则输出完整盘面 JSON（含 `payload` 与给 AI 的 `facts`）：

```
pwsh .\scripts\dev.ps1 paipan -coins 111,110,100,000,111,100 -question "这次考试能否通过"
```

查看历法上下文（干支、节气、月将、旬空）：

```
pwsh .\scripts\dev.ps1 calendar -date 2026-11-25 -time 14:30 -lon 113.26
```

`dev.ps1 shell` / `dev.sh shell` 可进入容器内的 shell。

构建生产镜像（多阶段，产物约 23MB）：

```
docker build -t mingyun:latest .
docker run --rm mingyun:latest version
docker run --rm -p 8080:8080 -e APP_SECRET=0123456789abcdef0123456789abcdef mingyun:latest serve
```

### 试试 API

排一张六爻盘（服务端用 crypto/rand 摇卦，响应里带可复核的 `entropy`）：

```
curl -s -X POST localhost:8080/api/v1/charts/liuyao/cast \
  -H 'Content-Type: application/json' \
  -d '{"mode":"coin","question":"这次考试能否通过"}'
```

手动录入（六爻自下而上）：

```
curl -s -X POST localhost:8080/api/v1/charts/liuyao/cast \
  -H 'Content-Type: application/json' \
  -d '{"mode":"manual","lines":["少阳","少阳","少阳","少阳","少阳","少阳"]}'
```

取完整盘面（含全部 facts）与列表：

```
curl -s "localhost:8080/api/v1/charts/{chart_id}"
curl -s "localhost:8080/api/v1/charts?limit=10"
```

其它只读接口：`GET /api/v1/meta/capabilities`、`GET /api/v1/meta/calendar?at=2026-11-25T14:30:00%2B08:00&lon=113.26`。

响应统一信封 `{code, message, data, request_id, server_time}`，错误码见 `docs/05`。当前盘面存在内存里，重启即丢——Postgres 在 M2 接入。

### 方式二：本机 Go

需要 Go 1.25+。

```
cd paipan && go test ./...
cd ../apps/api && go test ./... && go build -o ../bin/mingyun ./cmd/mingyun
../bin/mingyun paipan -summary -coins 111,110,100,000,111,100
```

启动 HTTP 服务（当前只有健康检查）：

```
./bin/mingyun serve -addr :8080
curl localhost:8080/healthz
```

> 二进制的 `calendar` / `paipan` 已内嵌时区库（`time/tzdata`），容器或精简系统里没有 tzdata 也能解析 `Asia/Shanghai`。

## 模块说明

`paipan` 与 `apps/api` 是两个独立 module，根目录用 `go.work` 串联。发布 `paipan` 时按 `paipan/v0.x.y` 打 tag，主模块去掉本地 `replace` 即可。

## 文档

完整的开发文档在 `docs/`：产品需求、系统架构、排盘算法规格、数据模型、API 契约、AI 解读工程、前端规范、部署运维、合规风险、路线图、Go 上手路线。

## 许可

MIT，详见 `LICENSE`。

本项目仅供传统文化学习、研究与娱乐使用，不构成任何专业建议。部署为公开服务时，运营者需自行承担所在地区的合规义务，详见 `docs/09-合规与风险.md`。
