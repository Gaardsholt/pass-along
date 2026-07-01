FROM golang:1.26.4-alpine@sha256:7a3e50096189ad57c9f9f865e7e4aa8585ed1585248513dc5cda498e2f41812c AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags="-w -s"

FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
WORKDIR /app

RUN addgroup --system --gid 1001 appgroup && \
    adduser -S -u 1001 -G appgroup appuser && \
    chown -R appuser:appgroup /app

COPY --chown=1001:1001 --from=builder /app/pass-along pass-along
COPY --chown=1001:1001 ./static static/

USER appuser

CMD ["/app/pass-along"]
