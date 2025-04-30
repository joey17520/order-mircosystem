FROM golang:1.23.8-alpine AS builder
LABEL authors="Joey"

WORKDIR /app
ENV TZ=Asia/Shanghai

COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .
RUN go build -o api-service ./cmd/server/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/api-service ./api-service
COPY --from=builder /app/config ./config
EXPOSE 8080

CMD ["./api-service"]
