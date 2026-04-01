FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/rustler ./cmd/rustler/
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata curl
COPY --from=builder /bin/rustler /usr/local/bin/rustler
ENV PORT="8980" DATA_DIR="/data"
EXPOSE 8980
HEALTHCHECK --interval=30s --timeout=5s CMD curl -sf http://localhost:8980/health || exit 1
ENTRYPOINT ["rustler"]
