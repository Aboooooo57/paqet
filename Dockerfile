# Build Stage
FROM golang:1.25-bookworm AS builder

# Install build dependencies
ENV GOPROXY=https://goproxy.io,direct
RUN apt-get update && apt-get install -y libpcap-dev

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=1 is required for pcap
RUN CGO_ENABLED=1 GOOS=linux go build -o paqet ./cmd

# Runtime Stage
FROM debian:bookworm-slim

WORKDIR /app

# Install runtime dependencies
# libpcap0.8: required for packet capture
# iptables: for firewall rule manipulation
# iproute2: for ip/ss commands (debugging)
# libcap2-bin: for capsh (used in entrypoint)
RUN apt-get update && apt-get install -y \
    libpcap0.8 \
    iptables \
    iproute2 \
    libcap2-bin \
    dos2unix \
    && rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /app/paqet /usr/local/bin/paqet
RUN chmod +x /usr/local/bin/paqet

# Copy entrypoint script
COPY docker-entrypoint.sh /usr/local/bin/
RUN dos2unix /usr/local/bin/docker-entrypoint.sh && chmod +x /usr/local/bin/docker-entrypoint.sh

# Copy default config template
COPY config_template.yaml /app/config.yaml

# Entrypoint
ENTRYPOINT ["docker-entrypoint.sh"]

# Default command
CMD ["paqet", "run", "-c", "config.yaml"]
