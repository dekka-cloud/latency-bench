# The target, as a small static binary on a distroless base that runs as a non-root user.
FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/target ./cmd/target \
 && mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/target /target
# Shared results go to /data, owned by the user the target runs as, so a volume mounted
# there starts out writable.
COPY --from=build --chown=65532:65532 /out/data /data
EXPOSE 8080 8443 7001/udp
ENTRYPOINT ["/target"]
