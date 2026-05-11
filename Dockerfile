# Build multi-étape [cite: 326]
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o forum ./cmd/forum/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates sqlite
WORKDIR /root/
COPY --from=builder /app/forum .
COPY --from=builder /app/web ./web
COPY --from=builder /app/certs ./certs
EXPOSE 443
CMD ["./forum"]
