FROM golang:1.17.1-alpine AS builder
WORKDIR $GOPATH/src/app
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /tmp/app

FROM alpine
COPY --from=builder /tmp/app /app

RUN mkdir -p /config

RUN groupadd -g 2000 AppGroup && \
    useradd -m -u 2001 -g AppGroup AppUser && \
    chown -R AppUser /config

USER AppUser
CMD ["/app"]