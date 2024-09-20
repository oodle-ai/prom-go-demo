FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first (if you have them)
COPY go.mod go.sum* ./

# Install dependencies
RUN go mod download

# Now copy the rest of the source code
COPY app.go ./

# Build the application
RUN go build -o main .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE ${APP_PORT}
CMD ["./main"]
