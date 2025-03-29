FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .

ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=1

RUN apk add --no-cache gcc musl-dev sqlite-dev
RUN apk add --no-cache tzdata ca-certificates sqlite

RUN cd server && go mod download && go mod tidy

RUN cd server && go test -v ./...

RUN cd server/cmd && go build -o http-download-server .

FROM alpine:latest

WORKDIR /app

RUN mkdir -p /app/downloads

COPY --from=builder /app/server/cmd/http-download-server /app/
COPY --from=builder /app/ui /app/ui

EXPOSE 8000

ENV GIN_MODE=release

CMD ["./http-download-server", "-dsn", "file:/app/data.db"]
