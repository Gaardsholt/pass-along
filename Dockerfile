FROM golang:1.25.1-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags="-w -s"

FROM alpine:3.22.1
WORKDIR /app

RUN addgroup --system --gid 1001 appgroup && \
    adduser -S -u 1001 -G appgroup appuser && \
    chown -R appuser:appgroup /app

COPY --chown=1001:1001 --from=builder /app/pass-along pass-along
COPY ./static static/

USER appuser

CMD ["/app/pass-along"]
