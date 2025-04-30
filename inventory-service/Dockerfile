FROM golang:1.23.8-alpine AS builder
LABEL authors="Joey"

WORKDIR /app
ENV TZ=Asia/Shanghai

COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .
RUN go build -o inventory-service ./cmd/server/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/inventory-service ./inventory-service
COPY --from=builder /app/config ./config
EXPOSE 50052
EXPOSE 8082

CMD ["./inventory-service"]
