# Use a base image that matches the binary's architecture
FROM alpine:latest

# Install necessary runtime dependencies and file command
RUN apk add --no-cache ca-certificates libc6-compat file

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory
WORKDIR /app

# Copy the binary from your build context
COPY discovr /app/discovr

# Set permissions and ownership
RUN chmod +x /app/discovr && \
    chown appuser:appgroup /app/discovr

# Debugging step: verify binary
RUN ls -l /app/discovr && \
    file /app/discovr

# Switch to non-root user
USER appuser

# Use a shell entrypoint to support argument passing
ENTRYPOINT ["/app/discovr"]

