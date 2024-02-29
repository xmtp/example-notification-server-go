# BUILD IMAGE --------------------------------------------------------
FROM golang:1.20-alpine as builder

# Get build tools and required header files
RUN apk add --no-cache build-base

WORKDIR /app
COPY . .

# Build the final node binary
ARG GIT_COMMIT=unknown
ARG XMTP_GO_CLIENT_VERSION=unknown
RUN go build \
    -o bin/notifications-server \
    -ldflags="-X 'main.GitCommit=$GIT_COMMIT' -X 'main.XMTPGoClientVersion=$XMTP_GO_CLIENT_VERSION'" \
    cmd/server/main.go

# ACTUAL IMAGE -------------------------------------------------------

FROM alpine:3.12

LABEL maintainer="engineering@xmtp.com"
LABEL source="https://github.com/xmtp/example-notification-server-go"
LABEL description="XMTP Example Notification Server"

# color, nocolor, json
ENV GOLOG_LOG_FMT=nocolor

# go-waku default port
EXPOSE 5556

COPY --from=builder /app/bin/notifications-server /usr/bin/

ENTRYPOINT ["/usr/bin/notifications-server"]
# By default just show help if called without arguments
CMD ["--help"]
