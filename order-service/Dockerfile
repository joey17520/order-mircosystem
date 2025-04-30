FROM golang:1.23.8-alpine AS builder
LABEL authors="Joey"

WORKDIR /app
ENV TZ=Asia/Shanghai

COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .
RUN go build -o order-service ./cmd/server/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/order-service ./order-service
COPY --from=builder /app/config ./config
EXPOSE 50051
EXPOSE 8081

CMD ["./order-service"]
