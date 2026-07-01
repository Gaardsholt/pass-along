FROM golang:1.26.4-alpine@sha256:3ad57304ad93bbec8548a0437ad9e06a455660655d9af011d58b993f6f615648 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags="-w -s"

FROM alpine:3.24.0@sha256:a2d49ea686c2adfe3c992e47dc3b5e7fa6e6b5055609400dc2acaeb241c829f4
WORKDIR /app

RUN addgroup --system --gid 1001 appgroup && \
    adduser -S -u 1001 -G appgroup appuser && \
    chown -R appuser:appgroup /app

COPY --chown=1001:1001 --from=builder /app/pass-along pass-along
COPY --chown=1001:1001 ./static static/

USER appuser

CMD ["/app/pass-along"]
