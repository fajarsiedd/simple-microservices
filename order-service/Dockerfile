FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /app/order-service ./main.go

FROM scratch

WORKDIR /app

COPY --from=builder /app/order-service .

COPY --from=builder /app/.env .

EXPOSE 8080

CMD ["./order-service"]