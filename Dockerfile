# ---------- Build stage ----------
FROM golang:1.23.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o fashionart .

# ---------- Runtime stage ----------
FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/fashionart .

COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

RUN mkdir -p /app/upload

EXPOSE 8080

CMD ["./fashionart"]