# Use a minimal Alpine Linux-based image as the base image
FROM golang:1.20-alpine AS builder

# Set the working directory in the container
WORKDIR /app

# Copy the bot source code to the container
COPY . .

# Build the bot application
RUN go build -o /out/bot ./cmd/bot/main.go

# Create a new stage to keep the final image small
FROM alpine:latest

# Set the working directory in the container
WORKDIR /app

# Copy the built bot application from the previous stage
COPY --from=builder /out/bot .

# Start the bot when the container runs
CMD ["./bot"]