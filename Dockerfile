FROM golang:1.19.0-alpine AS builder
WORKDIR $GOPATH/src/app
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags="-w -s" -o /tmp/app

FROM alpine
RUN mkdir /app
WORKDIR /app
COPY --from=builder /tmp/app app
ADD ./static static/

RUN addgroup -S appgroup && adduser -S appuser -G appgroup && chown -R appuser /app
USER appuser

CMD ["/app/app"]