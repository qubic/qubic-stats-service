FROM golang:1.22 AS builder
ENV CGO_ENABLED=0

WORKDIR /build
COPY .. /build

RUN go build -o "/build/qubic-stats-processor"


FROM alpine:latest
COPY --from=builder /build/qubic-stats-processor /app/processor
RUN chmod +x /app/processor

EXPOSE 80

WORKDIR /app

ENTRYPOINT ["./processor"]