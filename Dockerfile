# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG BUILD_VERSION=container
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X xdns-go/internal/version.BuildVersion=${BUILD_VERSION}" -o /out/XDNS-server ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata su-exec libcap && addgroup -S xdns && adduser -S -G xdns xdns
WORKDIR /data
COPY --from=build /out/XDNS-server /usr/local/bin/XDNS-server
# The server binds UDP/TCP :53 (privileged) but the entrypoint drops to the
# non-root XDNS user via su-exec before it binds. A file capability survives
# the setuid + execve that su-exec performs, so grant only CAP_NET_BIND_SERVICE on
# the binary itself. Docker's default bounding set already includes it (compose
# also cap_add's it), so the dropped user can bind :53 without the whole process
# running as root. Without this, binding :53 fails with EACCES and the container
# crash-loops.
RUN setcap 'cap_net_bind_service=+ep' /usr/local/bin/XDNS-server
COPY server_config.toml.simple /opt/XDNS/server_config.toml.simple
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod 0755 /usr/local/bin/docker-entrypoint.sh && chown -R XDNS:XDNS /data
EXPOSE 53/udp 53/tcp 9090/tcp
VOLUME ["/data"]
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget -qO- http://127.0.0.1:9090/healthz || exit 1
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
