# ========== 构建阶段 ==========
FROM golang:1.25-alpine AS builder

WORKDIR /build

# 国内网络环境加速依赖下载，海外构建时此行可删除
ENV GOPROXY=https://goproxy.cn,direct

# 先只复制依赖清单并下载，利用 Docker 层缓存：
# 只要 go.mod / go.sum 没变，改业务代码不会触发重新下载依赖
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0：纯静态编译，二进制不依赖系统 C 库，才能在 alpine 中直接运行
# -ldflags "-s -w"：去掉符号表和调试信息，减小二进制体积
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o gateway ./cmd/gateway

# ========== 运行阶段 ==========
FROM alpine:3.20

# ca-certificates：HTTPS 调用上游模型（DeepSeek / GLM）必需，否则报 x509 证书错误
# tzdata：时区数据，MySQL DSN 中 loc=Local 依赖它
RUN apk add --no-cache ca-certificates tzdata

# WORKDIR 必须是 /app：代码中以相对路径 configs/config.yaml 加载配置，
# k8s 的 ConfigMap 挂载点也是 /app/configs，两者对齐
WORKDIR /app

COPY --from=builder /build/gateway .

EXPOSE 8080

ENTRYPOINT ["./gateway"]
