# Use an official Golang image as the base image
FROM golang:1.24

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files first to leverage Docker cache
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the entire project to the working directory
COPY . .

# Build the Go application
RUN go build -o /app/build/shiroxy ./cmd/shiroxy  # Adjust the build path and main package path as needed

# Expose the necessary ports
EXPOSE 80
EXPOSE 443
EXPOSE 2210

# Command to run the Go app with the config file
CMD ["/app/build/shiroxy", "-c", "/app/defaults/shiroxy.conf.yaml"]
