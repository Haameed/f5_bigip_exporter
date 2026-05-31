# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.25-alpine AS build

WORKDIR /src

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w \
      -X main.version=${VERSION} \
      -X main.commit=${COMMIT} \
      -X main.buildDate=${BUILD_DATE}" \
    -o /bin/bigip_exporter ./cmd/bigip_exporter

# ---- Runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.title="bigip_exporter" \
      org.opencontainers.image.description="Prometheus exporter for F5 BIG-IP devices" \
      org.opencontainers.image.source="https://github.com/Haameed/bigip_exporter" \
      org.opencontainers.image.licenses="MIT"

COPY --from=build /bin/bigip_exporter /bin/bigip_exporter

EXPOSE 9142
USER nonroot:nonroot

ENTRYPOINT ["/bin/bigip_exporter"]
CMD ["-config", "/etc/bigip_exporter/bigip-config.yaml"]
