FROM golang:1.23-alpine AS builder
WORKDIR $GOPATH/src/app
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags="-w -s" -o /tmp/app

FROM alpine
RUN mkdir /app
WORKDIR /app

RUN addgroup --system --gid 1001 appgroup && \
    adduser -S -u 1001 -G appgroup appuser && \
    chown -R appuser:appgroup /app

COPY --chown=1001:1001 --from=builder /tmp/app app

ADD ./static static/

USER appuser

CMD ["/app/app"]