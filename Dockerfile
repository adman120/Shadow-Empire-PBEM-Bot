## Build the app
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o shadow-empire-bot .

## Run the app
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/shadow-empire-bot .
RUN mkdir -p /app/data
VOLUME /app/data
ENV WATCH_DIRECTORY=/app/data

CMD ["./shadow-empire-bot"]
