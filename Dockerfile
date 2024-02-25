# First stage: build the application
FROM golang:1.22-alpine as builder

# Install build dependencies
RUN apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp .

# Second stage: setup the runtime environment
FROM alpine:3.13

# Install bash for wait-for-it.sh and timezone data
RUN apk add --no-cache bash tzdata

# Set the timezone (optional)
RUN cp /usr/share/zoneinfo/America/Sao_Paulo /etc/localtime \
    && echo "America/Sao_Paulo" > /etc/timezone

# Create app directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/myapp .

# Copy wait-for-it.sh into the image
COPY wait-for-it.sh wait-for-it.sh

# Make wait-for-it.sh executable
RUN chmod +x wait-for-it.sh

# Command to run the application using wait-for-it.sh
CMD ["./wait-for-it.sh", "db:5432", "--timeout=30", "--", "./myapp"]
