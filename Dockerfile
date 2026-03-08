FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN swag init -g cmd/server/main.go

RUN CGO_ENABLED=0 GOOS=linux go build -o subdomain-finder ./cmd/server

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/subdomain-finder .

EXPOSE 8080

CMD ["./subdomain-finder"]