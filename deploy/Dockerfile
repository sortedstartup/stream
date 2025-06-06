FROM gcr.io/distroless/base

# Accept environment name as build arg
ARG ENV_NAME=staging

# Copy the binary generated for the given environment
COPY ./backend/mono/stream-${ENV_NAME}-binary /app

EXPOSE 8080

# Set entrypoint
WORKDIR /
CMD ["/app"]
