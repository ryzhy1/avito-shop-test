FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN go build -o /app/bin/app ./cmd/main.go

FROM alpine:3.18 AS runner

WORKDIR /app

COPY --from=builder /app/bin/app /app/app

COPY migrations /app/migrations

COPY --from=builder /go/bin/goose /usr/local/bin/goose

COPY .env.local /app/.env.local

CMD ["sh", "-c", "goose -dir /app/migrations postgres 'user=postgres password=postgres dbname=postgres host=db port=5432 sslmode=disable' up && ./app"]