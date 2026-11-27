# 命云 MingYun

[![CI](https://github.com/huayuehua15-netizen/mingyun/actions/workflows/ci.yml/badge.svg)](https://github.com/huayuehua15-netizen/mingyun/actions/workflows/ci.yml)

开源的 AI 六术排盘平台：六爻、大六壬、紫微斗数、小六壬、梅花易数、奇门遁甲。盘面由确定性 Go 引擎计算，解读交给可自主配置的 OpenAI 兼容模型；没有配置模型时，站点仍可排盘，并给出只列事实、不判吉凶的离线摘要。

**技术栈**：Vue 3 + TypeScript / Go 1.25 / Docker Compose / GitHub Actions

## 当前状态

当前**六术主链路均已打通**：六爻铜钱摇卦 / 手动录入，大六壬月将加时起课，紫微斗数按公历或农历生辰排命盘；小六壬按农历月日时落六宫，梅花支持时间起卦与两组报数，奇门按节气与符头起时家转盘。每条链路都走同一套结构：Go 确定性排盘 → 扁平 facts → AI 四段流式解读 / 离线摘要；首轮解读完成后，可以沿同一张盘继续追问，不重新起盘。

大六壬的九宗门、紫微的辅星与流年、奇门的格局与十干克应仍有规则覆盖边界，详见下文“已知边界”。

| 模块 | 状态 |
|---|---|
| `paipan/calendar` 历法层（干支、节气、月将、旬空、真太阳时、子初换日） | ✅ |
| `paipan/liuyao` 六爻（八宫、纳甲、六亲、六神、世应、伏神、旺衰、动变、通用取用神） | ✅ v0.2.0 |
| `paipan/daliuren` 大六壬（天地盘、四课、三传九宗门、十二天将、六亲、神煞、课体） | ✅ v0.1.0 |
| `paipan/ziwei` 紫微斗数（命身宫、五行局、十四主星、常用辅星、生年四化、十二宫、大限） | ✅ v0.1.0 |
| `paipan/xiaoliuren` 小六壬（月日时起课、六宫属性、落宫与近事窗口） | ✅ v0.1.0 |
| `paipan/meihua` 梅花易数（时间卦 / 报数卦、本互变、动爻、体用生克） | ✅ v0.1.0 |
| `paipan/qimen` 奇门遁甲（节气定局、符头定元、地盘三奇六仪、值符值使、九星八门八神） | ✅ v0.1.0 |
| `apps/api` 单二进制（`serve` / `calendar` / `paipan`） | ✅ |
| HTTP API：六术排盘、盘面详情、列表、历法、能力清单、SSE 解读与同盘追问 | ✅ |
| `apps/web` 游戏化四场景（山门 / 案上起盘 / 卦历 / 六术）、六卷行卷、六套盘面、自动留档、完整盘面与解读详情、同盘追问、SSE 解读 | ✅ |
| AI 解读（OpenAI 兼容 SSE、按术数分流的提示词与校验、同一张盘内追问、离线兜底） | ✅ |
| Docker Compose（API + Nginx 静态站） | ✅ |
| PostgreSQL / Redis 持久化、账号与额度 | ⏳ |

> 盘面目前存在内存里，API 重启即丢。生产部署前请先接 Postgres，或把「重启丢盘」写进你的使用说明。

**已知边界**：大六壬三传的涉害深浅、伏吟中末传、返吟无克取驿马、别责 / 八专采用通行简化取法，`school` 会记录流派开关；九宗门“每门 ≥15 例”的人工核对集尚未补齐。紫微已实现十四主星、常用辅星、生年四化和大限，流年与部分辅星表尚未展开；与专业软件 ≥200 例的逐字段对齐也仍在补测。小六壬只做月日时起课，小吉取木 / 六合存在流派差异，`school` 已记录；梅花采用通行 8×8 卦名表，体用以“动爻所在卦为用”，报数固定两组或三组；奇门采用拆补法、符头定元、天禽寄坤、九星顺时针、八门顺时针布、八神阳顺阴逆，不含超神接气置闰、十干克应、格局与用神体系。六术都能稳定出盘并完成四段解读，但不要把当前版本当作专业排盘软件的完全替代。

## 页面

界面按 PC 游戏的方式分成四幕：山门、案上起盘、卦历、六术。首页把六术排成六卷行卷路线，六卷均已开卷；底部导航固定在手边，纸本、朱砂、水墨山景与盘面保持同一套视觉；没有渐变、发光和 3D。

山门与六卷行卷：

<img src="docs/preview/six-methods-home.png" width="760" alt="山门与六术说明">
<img src="docs/preview/six-methods-home-bottom.png" width="760" alt="六卷均已开卷">

案上起盘。六爻保留铜钱六摇与逐爻生长；大六壬按时刻起课，紫微按生辰排命盘；小六壬按农历月日时落六宫，梅花可时间起卦或报数起卦，奇门按节气起九宫：

<img src="docs/preview/final-cast.png" width="760" alt="六爻起卦与盘面">

<img src="docs/preview/daliuren-cast.png" width="520" alt="大六壬起课表单">
<img src="docs/preview/daliuren-board.png" width="520" alt="大六壬天地盘与三传">

<img src="docs/preview/ziwei-cast.png" width="520" alt="紫微生辰表单">
<img src="docs/preview/ziwei-board.png" width="520" alt="紫微十二宫盘">

<img src="docs/preview/xiaoliuren-board.png" width="520" alt="小六壬月日时三宫与六宫盘">
<img src="docs/preview/meihua-board.png" width="520" alt="梅花本卦、互卦、变卦与体用">
<img src="docs/preview/qimen-board.png" width="520" alt="奇门九宫、星门神与近事用神">

起盘成功即自动收入「卦历」，落一枚「录」印；解读完成后再补写解读与时间，给出再问 / 翻历史的下一步：

<img src="docs/preview/case-seal.png" width="760" alt="起盘自动留档">

卦历：六术共用同一条记录流。每次起盘自动留档（含未解读/失败的记录），保存完整盘面、关键事实与当时解读；支持搜索、按状态筛选、展开详情、翻回原盘、导出 JSON、删除与清空。只存在浏览器 localStorage，最多 100 则，不上传服务端：

<img src="docs/preview/six-methods-archive.png" width="760" alt="六术卦历列表">

同盘追问：首轮四段完成后，在盘面下方沿同一张盘继续问，不重新起盘。追问沿用原盘 facts 与已发生对话，仍做盘面事实硬校验；最多 6 轮，换问题会提示另起一卦/课/盘。追问记录写入同一条卦历：

<img src="docs/preview/followup.png" width="760" alt="同盘追问">
<img src="docs/preview/followup-history.png" width="760" alt="卦历中的追问记录">

六术：六爻、大六壬、紫微斗数、小六壬、梅花易数、奇门遁甲均已开卷，各自有独立的盘面与断法：

<img src="docs/preview/six-methods-codex.png" width="760" alt="六术典籍页">

六术的真实模型四段解读：

<img src="docs/preview/daliuren-reading.png" width="520" alt="大六壬四段解读">
<img src="docs/preview/ziwei-reading.png" width="520" alt="紫微斗数四段解读">
<img src="docs/preview/xiaoliuren-reading.png" width="520" alt="小六壬四段解读">
<img src="docs/preview/meihua-reading.png" width="520" alt="梅花易数四段解读">
<img src="docs/preview/qimen-reading.png" width="520" alt="奇门遁甲四段解读">

> 截图里的解读是真实模型 `sensenova-6.8-flash` 跑出的四段解读；未配置模型时会走离线摘要，四段结构保留，只列盘面事实、不判吉凶。

早期纸本界面的未占态与逐爻生长：

<img src="docs/preview/home.png" width="720" alt="未占态">
<img src="docs/preview/partial.png" width="720" alt="逐爻生长">

一次摇卦的三枚铜钱动画（GIF 为慢放）：

<img src="docs/preview/toss.gif" width="720" alt="抛币动画">

完整盘面与四段解读示例：

<img src="docs/preview/result.png" width="720" alt="完整盘面 · 离线摘要">
<img src="docs/preview/result-ai-demo.png" width="720" alt="四段解读 · 本地 mock 模型">

> `result-ai-demo.png` 是本地 mock 模型跑通的流式四段演示，不代表具体模型的固定文风。

## 快速开始

### Docker Compose（推荐）

```
cp .env.example .env
# 改 .env 里的 APP_SECRET；要接模型再填 MINGYUN_AI_*
docker compose up --build
```

打开 <http://localhost:8080>。Nginx 托管前端，并把 `/api/` 反代到 API 容器；API 不单独暴露端口。SSE 已关闭代理缓冲，解读是逐字出现的。

只想跑离线排盘，不填任何 `MINGYUN_AI_*` 即可。

> 如果 `docker compose build` 卡在 Docker Hub 的 TLS handshake timeout，在 `.env` 里把 `GO_IMAGE` / `ALPINE_IMAGE` / `NODE_IMAGE` / `NGINX_IMAGE` 换成可访问的镜像源（`.env.example` 里有 DaoCloud 示例），或先手动 `docker pull` 对应镜像并 `docker tag` 成默认名。Compose 已把四个基础镜像都做成了 build arg。

### 本地开发

需要 Go 1.25+、Node 22+。

```
# 终端 1：API
cd apps/api
go run ./cmd/mingyun serve -addr :8080

# 终端 2：前端
cd apps/web
npm ci
npm run dev
```

Vite 默认把 `/api` 代理到 `http://localhost:8080`；需要换目标时：

```
VITE_API_TARGET=http://localhost:18080 npm run dev
```

### 只跑 API / 命令行

```
cd apps/api
go test ./...
go build -o ../bin/mingyun ./cmd/mingyun

# 排一张六爻盘（1 为背，自下而上六组）
../bin/mingyun paipan -summary -coins 111,110,100,000,111,100 -question "这次考试能否通过"

# 起一课大六壬（月将加时，question 用于后续解读）
../bin/mingyun paipan -summary -method daliuren -date 2026-11-27 -time 17:48 -lon 113.39 -question "这桩合作能否谈成"

# 排一张紫微命盘（公历生辰；农历加 -calendar lunar，闰月加 -leap-month）
../bin/mingyun paipan -summary -method ziwei -gender male -calendar solar -birth-date 1990-05-15 -birth-time 10:30 -lon 113.39

# 小六壬：农历月日时落六宫
../bin/mingyun paipan -summary -method xiaoliuren -date 2026-11-27 -time 10:30 -lon 113.39 -question "这次面试结果如何"

# 梅花：时间起卦，或 -numbers 3,4 报数起卦
../bin/mingyun paipan -summary -method meihua -date 2026-11-27 -time 10:30 -lon 113.39 -question "这次比赛能否进决赛"
../bin/mingyun paipan -summary -method meihua -numbers 3,4 -question "这次比赛能否进决赛"

# 奇门：时家转盘，按节气与符头起局
../bin/mingyun paipan -summary -method qimen -date 2026-11-27 -time 06:41 -lon 113.39 -question "这次考试走势如何"

# 查历法
../bin/mingyun calendar -date 2026-11-25 -time 14:30 -lon 113.39
```

> 紫微默认启用真太阳时。1990-05-15 10:30 按东经 113.39° 校正后是 09:07，因为当时中国实行夏令时（UTC+9）；这是历史时区的正确结果，不是算错。

## 接入模型

服务端只读环境变量，不把 Key 下发到浏览器。任何 OpenAI 兼容的 `/chat/completions` 流式接口都能接。

| 变量 | 说明 | 默认 |
|---|---|---|
| `MINGYUN_AI_BASE_URL` | 例如 `https://api.deepseek.com/v1`、`http://host.docker.internal:11434/v1` | 空 |
| `MINGYUN_AI_API_KEY` | 本地模型可留空 | 空 |
| `MINGYUN_AI_MODEL` | 模型名，如 `deepseek-chat`、`qwen-plus` | 空 |
| `MINGYUN_AI_TIMEOUT` | 单次请求超时，Go duration | `90s` |
| `MINGYUN_AI_MAX_TOKENS` | 输出上限，最大 8192 | `1200` |
| `MINGYUN_AI_TEMPERATURE` | 0–2 | `0.5` |

`BASE_URL` 与 `MODEL` 同时非空才启用在线解读；否则走 `offline-rulebook@0.1` 离线摘要。

几个可直接用的组合：

```
# DeepSeek
MINGYUN_AI_BASE_URL=https://api.deepseek.com/v1
MINGYUN_AI_MODEL=deepseek-chat
MINGYUN_AI_API_KEY=sk-...

# 通义千问 DashScope 兼容模式
MINGYUN_AI_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
MINGYUN_AI_MODEL=qwen-plus

# 本地 Ollama（Docker 里访问宿主机）
MINGYUN_AI_BASE_URL=http://host.docker.internal:11434/v1
MINGYUN_AI_MODEL=qwen2.5:7b

# 商汤日日新 Token Plan（OpenAI 兼容）
MINGYUN_AI_BASE_URL=https://token.sensenova.cn/v1
MINGYUN_AI_MODEL=sensenova-6.8-flash
MINGYUN_AI_TEMPERATURE=0.3
```

Token Plan 的 `kimi-k3` / `glm-5.3` 在当前免费额度下容易触发 TPM/RPM 429；本项目实测用 `sensenova-6.8-flash` 跑通了解读与盘面校验，因此默认示例选它。

> 公开部署且配置了付费 Key 时，请在 Nginx 或网关上加限流/鉴权。当前 API 没有账号体系，解读接口是开放的。

## 解读是怎么做的

1. Go 引擎只产出 `facts`。六爻：问事类别、用神、卦名、宫位、世应、月建日辰、旬空、六爻纳甲六亲六神动静、动爻变爻、伏神；大六壬：月将、日干支、四课、三传、课体、十二天将、旬空与神煞；紫微：命身宫、五行局、十二宫主星与辅星、生年四化、大限、命宫三方四正；小六壬：农历月日时、三宫路径、落宫、六宫属性与近事窗口；梅花：起卦方式、本卦 / 互卦 / 变卦、动爻、体用与体用关系；奇门：阴阳遁与局数、节气与三元、四柱、旬首、值符值使、九宫星门神、空亡马星门迫与近事用神落宫。AI 不排盘、不补星、不补爻、不改卦。
2. 六爻取用神由 `paipan/liuyao/yongshen.go` 的规则表决定，不看模型心情：考试父母、求财妻财、事业官鬼、感情男取妻财/女取官鬼、健康世爻、出行世爻、失物妻财、官司官鬼、家人按所指六亲、测姓自测取世爻/问他人取应爻。类别可由 `topic` 指定，未指定时按问题关键词分类；感情未标性别时并列两种取法并标 `待确认`，不硬取。测姓字谜只从真实姓氏候选取，生肖本身不能当姓。

![通用取用神事实](docs/preview/yongshen-facts.png)

3. 提示词按术数分流。`SystemPromptFor(method)` 给六术各自独立的角色、断法框架和输出约束，避免六术共用一套六爻话术而串味；`FactsText` 在末尾写明“未列出的内容一律视为盘面未提供”。四段固定为：`盘面要点 / 断语 / 应期与走向 / 建议`。
4. `POST /api/v1/charts/{id}/interpret` 用 SSE 返回 `meta / delta / usage / done / error`；前端用 `fetch` + `ReadableStream` 解析，逐字渲染。紫微允许不填问题，服务端会补默认的“看整体格局”问题；大六壬、小六壬、奇门要求先写明所问之事。
5. 在线模型先完整生成，再由审计做盘面事实校验。六爻走原有卦名、爻位、六亲、世应、动爻与用神校验；大六壬检查课体、三传与天将是否越界；紫微按“已知星曜 × 四化”确定性扫描；小六壬校验六宫名称；梅花用 64 卦名与八卦校验，防止说出本卦、互卦、变卦之外的卦；奇门校验九星、八门、八神、阴阳遁、三元、马星、空亡与门迫。六术共用绝对化措辞拦截，任一项不过就整段拦截，不把编造内容流到浏览器。紫微盘面长，默认输出上限抬到 2400 token，大六壬 2200，梅花 2200，奇门 2600，小六壬 1800；仍受 8192 总上限约束。
6. 没有模型时，同一路由返回分术数的离线摘要：六爻按卦面、大六壬按课体三传、紫微按命身与十二宫、小六壬按三宫落点、梅花按本互变与体用、奇门按局数与九宫值符值使，四段结构保留，但明确不判吉凶、不推应期。
7. 追问走 `POST /api/v1/charts/{id}/messages`：不吃首轮四段格式，提示词改为“同一张盘内的追问”，仍注入完整 facts、首轮问题与最近对话；`AuditFollowup` 继续按原术数拦截编造内容。最多 6 轮；若换了问题或对象，提示另起一卦/课/盘。
8. 前端把追问写入 `ArchiveEntry.followups`，刷新后仍在卦历详情里；若 API 重启导致服务端内存丢盘，前端请求会带本地 `chart` 快照，服务端按 `chart.method` 从 payload 重建对应 facts 后再回答，六术均已覆盖。

```
curl -N -X POST "localhost:8080/api/v1/charts/{chart_id}/interpret" \
  -H 'Content-Type: application/json' \
  -d '{"question":"这次考试能否通过","style":"plain"}'
```

## API

响应统一信封 `{code, message, data, request_id, server_time}`，错误码见 `docs/05`。

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/v1/charts/liuyao/cast` | 六爻起卦：`mode=coin` 传 `coins` 六组三枚，或 `mode=manual` 传 `lines`；可选 `topic/gender` |
| POST | `/api/v1/charts/daliuren/cast` | 大六壬起课：`at/tz/lon/use_true_solar_time/day_boundary`，`question` 用于后续解读 |
| POST | `/api/v1/charts/ziwei/cast` | 紫微排盘：`gender/calendar_type/birth_date/birth_time/is_leap_month/leap_month_policy`，可选 `question` |
| POST | `/api/v1/charts/xiaoliuren/cast` | 小六壬起课：`at/tz/lon/use_true_solar_time/day_boundary`，`question` 用于后续解读 |
| POST | `/api/v1/charts/meihua/cast` | 梅花起卦：`mode=time`，或 `mode=numbers` 传 2–3 个正整数 `numbers` |
| POST | `/api/v1/charts/qimen/cast` | 奇门起局：`at/tz/lon/use_true_solar_time/day_boundary`，`question` 用于后续解读 |
| GET | `/api/v1/charts/{id}` | 完整盘面（含全部 facts） |
| GET | `/api/v1/charts?limit=10` | 盘面列表；可加 `method=liuyao|daliuren|ziwei|xiaoliuren|meihua|qimen` |
| GET | `/api/v1/meta/capabilities` | 术数能力清单 |
| GET | `/api/v1/meta/calendar` | 历法上下文 |
| POST | `/api/v1/charts/{id}/interpret` | SSE 首轮解读；紫微可不传问题，大六壬必须有问题 |
| POST | `/api/v1/charts/{id}/messages` | SSE 同盘追问：`message` + 最近 `history`；服务端内存无此盘时可带本地 `chart` 快照重建 facts |

排一张六爻盘：

```
curl -s -X POST localhost:8080/api/v1/charts/liuyao/cast \
  -H 'Content-Type: application/json' \
  -d '{"mode":"coin","question":"这次考试能否通过"}'
```

大六壬与紫微：

```
curl -s -X POST localhost:8080/api/v1/charts/daliuren/cast \
  -H 'Content-Type: application/json' \
  -d '{"question":"这桩合作能否谈成","at":"2026-11-27T17:48:00+08:00","lon":113.39}'

curl -s -X POST localhost:8080/api/v1/charts/ziwei/cast \
  -H 'Content-Type: application/json' \
  -d '{"gender":"male","calendar_type":"solar","birth_date":"1990-05-15","birth_time":"10:30","lon":113.39}'
```

小六壬、梅花与奇门：

```
curl -s -X POST localhost:8080/api/v1/charts/xiaoliuren/cast \
  -H 'Content-Type: application/json' \
  -d '{"question":"这次面试结果如何","at":"2026-11-27T10:30:00+08:00","lon":113.39}'

curl -s -X POST localhost:8080/api/v1/charts/meihua/cast \
  -H 'Content-Type: application/json' \
  -d '{"question":"这次比赛能否进决赛","mode":"numbers","numbers":[3,4]}'

curl -s -X POST localhost:8080/api/v1/charts/qimen/cast \
  -H 'Content-Type: application/json' \
  -d '{"question":"这次考试走势如何","at":"2026-11-27T06:41:00+08:00","lon":113.39}'
```

服务端用 `crypto/rand` 摇卦时，六爻响应会带 `entropy.backs` 与 `entropy_proof`，可复核随机记录；前端一爻一摇时，六组背数由浏览器 CSPRNG 产生，提交后 Go 用同一规则复算，避免前端直接决定卦象。大六壬、紫微、小六壬、梅花、奇门都是确定性排盘，不需要熵源。

## 目录

| 路径 | 说明 |
|---|---|
| `paipan/` | 排盘引擎，独立 Go module |
| `paipan/calendar/` | 六术共用历法 |
| `paipan/liuyao/` | 六爻装卦与排盘 |
| `paipan/daliuren/` | 大六壬起课、三传、天将与 facts |
| `paipan/ziwei/` | 紫微斗数定盘、星曜、四化与大限 |
| `paipan/xiaoliuren/` | 小六壬月日时起课与六宫 |
| `paipan/meihua/` | 梅花时间卦 / 报数卦、本互变与体用 |
| `paipan/qimen/` | 奇门时家转盘：定局、值符值使、九宫星门神 |
| `paipan/schema/` | Chart / Fact / Method 契约 |
| `apps/api/` | Go 服务：HTTP、熵源、AI、内存存储 |
| `apps/api/internal/ai/` | OpenAI 兼容客户端、提示词、离线摘要 |
| `apps/web/` | Vue 3 前端、Nginx 镜像 |
| `docs/` | 产品、架构、算法、API、AI、前端、部署、合规、路线图 |

## 文档

先读 `docs/README.md`。重点：

- `docs/03-排盘算法规格.md`：历法与六术排盘规则
- `docs/liuyao-rulebook.md`：通用取用神规则表与未完成边界
- `docs/05-API接口契约.md`：信封、错误码、六术请求字段与 SSE
- `docs/06-AI解读与提示词工程.md`：四段输出、按术数分流的提示词与校验
- `docs/09-合规与风险.md`：公开部署的合规边界

## 路线图

- M1：六爻排盘、API、前端、Docker、AI 解读（已完成）
- M2：六爻全链路、Postgres、账号与额度、分享页
- M3：大六壬（基础链路已提前落地，九宗门核对集与流派开关 UI 待补）
- M4：紫微斗数（基础链路已提前落地，辅星/流年与 ≥200 例对齐待补）
- M4.5：小六壬、梅花、奇门（基础链路已提前落地；奇门格局、十干克应与用神体系待补）
- M5：账号商业化、管理后台、审计日志
- M6：开源发布 v1.0
- v1.1 候选：古籍 RAG、流年报告、评测集

## 许可

MIT，详见 `LICENSE`。

本项目仅供传统文化学习、研究与娱乐使用，不构成医疗、投资、法律或婚姻决策建议。公开部署时，运营者需自行承担所在地区的合规义务，详见 `docs/09-合规与风险.md`。
