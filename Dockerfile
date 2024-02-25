# Use an official Golang Alpine image to create a build artifact.
FROM golang:1.22-alpine as builder

# Install git, required for fetching Go modules.
# Also, ca-certificates is a common requirement for applications making HTTPS requests.
RUN apk add --no-cache git ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Fetch dependencies.
# Using go mod with Go 1.11 modules.
RUN go mod tidy

# Build the Go app as a static binary.
RUN CGO_ENABLED=0 go build -o /rinha

# Start from a fresh Alpine image to create a smaller final image
FROM alpine:3.18.3

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the static executable.
COPY --from=builder /rinha /rinha

# Run the binary.
CMD ["/rinha"]
