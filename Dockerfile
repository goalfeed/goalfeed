# Use the official Go image as the base image
FROM golang:1.21 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files and download the dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o goalfeed .

# Use a lightweight image for the final image
FROM alpine:latest

# Copy the binary from the builder image
COPY --from=builder /app/goalfeed /goalfeed

# Command to run the application
CMD ["/goalfeed"]
