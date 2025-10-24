FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o leetbot ./cmd/bot
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/leetbot .
CMD ["./leetbot"]
