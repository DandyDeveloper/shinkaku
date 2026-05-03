# ---- build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Download dependencies first (layer-cached unless go.mod/go.sum change)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 — modernc.org/sqlite is pure-Go, no C toolchain needed
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/shinkaku ./cmd/server
RUN mkdir -p /data

# ---- runtime stage ----
# distroless/static has no shell, no package manager — minimal attack surface
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/shinkaku /shinkaku
COPY --chown=nonroot:nonroot --from=builder /data /data

EXPOSE 8080

ENTRYPOINT ["/shinkaku"]
