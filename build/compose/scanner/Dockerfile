FROM golang:1.24-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/scanner.bin ./scanner/cmd/

FROM alpine:3.21

COPY --from=builder /app/scanner.bin /app/scanner

WORKDIR /app

CMD ["/app/scanner"]