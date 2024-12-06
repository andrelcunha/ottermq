# Stage 1: Build Stage
FROM golang:1.23-alpine AS builder

# Install build tools
RUN apk add --no-cache make git

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app using the Makefile
RUN make build

####################################################################
# Stage 2: Final Stage
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# RUN apk add --no-cache bash

# Copy the built binaries from the builder stage
COPY --from=builder /app/bin/ottermq .
COPY --from=builder /app/bin/ottermq-webadmin .

# Expose ports 8081 for the web admin and 5672 for the broker
EXPOSE 8081
EXPOSE 5672

# Command to run both broker and web admin binaries
CMD ["sh", "-c", "./ottermq & ./ottermq-webadmin"]
# CMD ["tail", "-f", "/dev/null"]
