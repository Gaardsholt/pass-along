FROM golang:1.17.1-alpine AS builder
WORKDIR $GOPATH/src/app
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /tmp/app

FROM alpine
COPY --from=builder /tmp/app /app

RUN mkdir -p /config
RUN addgroup -S appgroup && adduser -S appuser -G appgroup && chown -R appuser /config

USER appuser
CMD ["/app"]