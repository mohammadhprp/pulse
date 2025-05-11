FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o bin/agent cmd/agent/main.go && chmod +x bin/agent
RUN go build -o bin/collector cmd/collector/main.go && chmod +x bin/collector

COPY scripts/entrypoint.sh /app/scripts/entrypoint.sh
RUN chmod +x /app/scripts/entrypoint.sh

ENTRYPOINT ["sh", "/app/scripts/entrypoint.sh"]
