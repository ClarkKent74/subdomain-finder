FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN swag init -g cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o subdomain-finder ./cmd/server

FROM screetsec/sudomy:v1.2.0-dev

WORKDIR /app

COPY --from=builder /app/subdomain-finder .

RUN ln -sf /usr/lib/sudomy/engine /app/engine && \
    ln -sf /usr/lib/sudomy/plugin /app/plugin && \
    ln -sf /usr/lib/sudomy/lib /app/lib && \
    ln -sf /usr/lib/sudomy/templates /app/templates

RUN echo 'SHODAN_API=""\nCENSYS_API=""\nCENSYS_SECRET=""\nVIRUSTOTAL=""\nBINARYEDGE=""\nSECURITY_TRAILS=""' \
    > /usr/lib/sudomy/sudomy.api

EXPOSE 8080

ENTRYPOINT ["/app/subdomain-finder"]
CMD []