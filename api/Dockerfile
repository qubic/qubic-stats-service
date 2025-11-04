FROM golang:1.25 AS builder
ENV CGO_ENABLED=0

WORKDIR /build
COPY .. /build

RUN go build -o "/build/qubic-stats-api"


FROM alpine:latest
COPY --from=builder /build/qubic-stats-api /app/api
RUN chmod +x /app/api

EXPOSE 80

WORKDIR /app

ENTRYPOINT ["./api"]