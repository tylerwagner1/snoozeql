FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO disabled
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -a -installsuffix cgo -o /snoozeql ./cmd/server

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary
COPY --from=builder /snoozeql .

# Create non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

USER appuser

# Expose the server port
EXPOSE 8080

# Run the application
CMD ["./snoozeql"]
