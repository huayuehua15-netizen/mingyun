#!/bin/sh
# 用 Docker 跑 Go 工具链，保证任何机器上行为一致。
# 用法:
#   ./scripts/dev.sh test
#   ./scripts/dev.sh build
#   ./scripts/dev.sh paipan -summary -coins 111,110,100,000,111,100
#   ./scripts/dev.sh paipan -lines "少阳,少阴,老阳,少阴,少阳,老阴"
#   ./scripts/dev.sh calendar -date 2026-11-25 -time 14:30 -lon 113.26
#   ./scripts/dev.sh shell

set -e

task="${1:-test}"
if [ $# -gt 0 ]; then shift; fi

root=$(cd "$(dirname "$0")/.." && pwd)
image="golang:1.25-alpine"
build_script="cd /src/apps/api && mkdir -p /src/bin && CGO_ENABLED=0 go build -trimpath -o /src/bin/mingyun ./cmd/mingyun"

run_args="run --rm \
  -e GOPROXY=https://goproxy.cn,direct \
  -e GOTOOLCHAIN=local \
  -v $root:/src \
  -v mingyun-gomod:/go/pkg/mod"

build() {
  docker $run_args -w /src "$image" sh -c "$build_script"
}

case "$task" in
  test)
    docker $run_args -w /src "$image" sh -c \
      "cd /src/paipan && go vet ./... && go test -count=1 ./... && cd /src/apps/api && go vet ./... && go test -count=1 ./..."
    ;;
  build)
    build
    echo "已生成 bin/mingyun"
    ;;
  paipan)
    build
    docker $run_args --entrypoint /src/bin/mingyun "$image" paipan "$@"
    ;;
  calendar)
    build
    docker $run_args --entrypoint /src/bin/mingyun "$image" calendar "$@"
    ;;
  shell)
    exec docker $run_args -it -w /src "$image" sh
    ;;
  *)
    echo "未知任务: $task（可用: test build paipan calendar shell）" >&2
    exit 2
    ;;
esac
