# Use an official Golang image as the base image
FROM golang:1.21-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files first to leverage Docker cache
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the entire project to the working directory
COPY . .

# Build the Go app
RUN go build -o /app/shiroxy cmd/shiroxy/main.go

# Expose the application port if necessary (modify accordingly)
EXPOSE 80
EXPOSE 443

# Command to run the Go app with the config file
CMD ["/app/shiroxy", "-c", "/app/defaults/shiroxy.conf.yaml"]
