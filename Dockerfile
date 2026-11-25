ARG GO_VERSION=1.25

FROM golang:${GO_VERSION}-alpine AS builder
ARG GOPROXY=https://goproxy.cn,direct
ARG VERSION=0.1.0-dev
ENV GOPROXY=${GOPROXY} \
    GOTOOLCHAIN=local \
    CGO_ENABLED=0
WORKDIR /src

# 先只拷 go.mod，让依赖下载单独成层，改代码不必重新拉依赖。
COPY go.work ./
COPY paipan/go.mod paipan/go.sum ./paipan/
COPY apps/api/go.mod apps/api/go.sum ./apps/api/
RUN cd /src/apps/api && go mod download && cd /src/paipan && go mod download

COPY paipan ./paipan
COPY apps/api ./apps/api
RUN cd /src/apps/api \
 && go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/mingyun ./cmd/mingyun

FROM alpine:3.21
ARG VERSION=0.1.0-dev
LABEL org.opencontainers.image.title="mingyun" \
      org.opencontainers.image.description="AI 术数排盘平台（六爻 / 大六壬 / 紫微斗数）" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.licenses="MIT"

RUN apk add --no-cache ca-certificates \
 && addgroup -g 10001 -S mingyun \
 && adduser -u 10001 -S -G mingyun -H mingyun

COPY --from=builder /out/mingyun /usr/local/bin/mingyun

USER mingyun
EXPOSE 8080
ENV APP_PORT=8080 \
    TZ=Asia/Shanghai

ENTRYPOINT ["mingyun"]
CMD ["serve"]
