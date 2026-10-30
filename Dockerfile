# Printrment server image (Linux x86-64), for Docker / TrueNAS.
# Built by GitHub Actions (.github/workflows/docker.yml) and published to
# ghcr.io/clarkyau/printrment.

FROM golang:1.27.3-bookworm AS build
WORKDIR /src
COPY . .
# Everything the build needs is in the repo (third_party/), so no downloads.
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOFLAGS=-mod=mod GOPROXY=off
RUN go vet . && go test . && \
    go build -trimpath -ldflags "-s -w" -o /out/printrment . && \
    mkdir -p /out/data

# Nothing but the app: no shell, no package manager. The app carries its own
# FFmpeg (checked against its SHA-256 before each run).
FROM scratch
COPY --from=build /out/printrment /printrment
COPY --from=build --chown=568:568 /out/data /data
ENV PRINTRMENT_SERVER=1 \
    PRINTRMENT_DATA=/data \
    HOME=/data \
    TZ=UTC
# 568 = the "apps" user TrueNAS runs apps as
USER 568:568
VOLUME /data
EXPOSE 8787 8788
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 CMD ["/printrment", "--healthcheck"]
ENTRYPOINT ["/printrment"]
