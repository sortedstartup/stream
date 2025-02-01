# Stage 1: Build
FROM golang:1.23.4-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /build

# # Copy go mod and sum files
# COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
# RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o monoservice ./mono

RUN apk --update add ca-certificates

# Stage 2: Runtime
FROM scratch AS runtime

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /build/monoservice /service/

# Command to run the executable
CMD ["/service/monoservice"]