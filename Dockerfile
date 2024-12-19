# Stage 1: Build Stage
FROM golang:1.23-alpine AS builder

# Install build tools
RUN apk add --no-cache make git gcc g++

ENV CGO_ENABLED 1

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Download swaggo for generating swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Generate swagger docs
RUN make docs

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
COPY --from=builder /app/web/static ./web/static
COPY --from=builder /app/web/templates ./web/templates

# Expose ports 3000 for the web admin and 5672 for the broker
EXPOSE 3000
EXPOSE 5672

# Command to run both broker and web admin binaries
CMD ["sh", "-c", "./ottermq & ./ottermq-webadmin"]
# CMD ["tail", "-f", "/dev/null"]
