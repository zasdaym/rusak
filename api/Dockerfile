FROM golang:1.21.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o rusak-api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/rusak-api /app/rusak-api
ENTRYPOINT ["/app/rusak-api"]
