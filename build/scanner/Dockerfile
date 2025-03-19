FROM golang:1.24-alpine3.21 AS builder

WORKDIR /app

COPY ../../go.mod ../../go.sum ./
RUN go mod download

COPY ../../scanner ./scanner

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/scanner ./scanner/cmd/main.go

FROM alpine:3.21

COPY --from=builder /app/scanner /app/scanner

WORKDIR /app

EXPOSE 8080

CMD ["/app/main"]