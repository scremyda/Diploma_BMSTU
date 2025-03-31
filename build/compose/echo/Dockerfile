FROM golang:1.24-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/echo.bin ./echo/

FROM alpine:3.21

COPY --from=builder /app/echo.bin /app/echo

WORKDIR /app

CMD ["/app/echo"]