FROM golang:1.17.1-alpine AS builder
WORKDIR $GOPATH/src/app
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /tmp/app

FROM alpine
RUN apk add --no-cache bash
COPY --from=builder /tmp/app /app

CMD ["/app"]