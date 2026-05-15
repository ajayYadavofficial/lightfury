FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o lightfury ./cmd/main.go


FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/lightfury .

EXPOSE 8080

CMD ["./lightfury"]
