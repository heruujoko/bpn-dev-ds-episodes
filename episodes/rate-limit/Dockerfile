# Build stage: static binary (no cgo), stripped, reproducible paths.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /demo .

# Runtime: scratch. The binary is static, uses plain TCP to Redis (no TLS, so no CA certs),
# needs no shell, and only reads stdin / writes stdout for the menu. Image = just the binary.
FROM scratch
COPY --from=build /demo /demo
ENTRYPOINT ["/demo"]
