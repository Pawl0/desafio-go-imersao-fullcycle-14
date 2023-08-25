
FROM golang:1.16 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

FROM debian:buster-slim

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
