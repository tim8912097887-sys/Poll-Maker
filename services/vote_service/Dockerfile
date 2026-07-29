FROM golang:1.26.4-alpine AS builder
# Install git and certificates (needed for fetching dependencies)
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create a non-privileged user for the production stage
ENV USER=appuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /app

# Leverage Docker cache by copying and downloading dependencies first
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the source code
COPY . .

# Build the binary statically
# - ldflags="-w -s": Omits debugging information and symbols to reduce size
# - CGO_ENABLED=0: Disables cgo for a fully static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s" \
    -o /go/bin/myapp ./cmd
# ==========================================
# STAGE 2: Final lightweight runtime
# ==========================================
FROM scratch

# Copy system files needed for TLS, timezones, and non-root user
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy our static binary from the builder stage
COPY --from=builder /go/bin/myapp /go/bin/myapp


# Use the non-root user we created in the builder stage
USER appuser:appuser

# Port documentation (change to match your app)
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/go/bin/myapp"]