# ---- Etapa de build ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o iam-service ./cmd/api

# ---- Etapa final ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/iam-service .

EXPOSE 8001

ENTRYPOINT ["./iam-service"]