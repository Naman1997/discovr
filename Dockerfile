# Use a base image that matches the binary's architecture
FROM alpine:latest

# Install necessary runtime dependencies and file command
RUN apk add --no-cache ca-certificates libc6-compat file sudo libpcap-dev

# Create a root user
ARG NEW_USER=appuser

# Create the new user
RUN adduser -D -s /bin/sh ${NEW_USER}

# Grant sudo privileges without password
RUN echo "${NEW_USER} ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/${NEW_USER} \
    && chmod 0440 /etc/sudoers.d/${NEW_USER}

# Set the working directory
WORKDIR /app

# Copy the binary from your build context
COPY discovr /app/discovr

# Set permissions and ownership
RUN chmod +x /app/discovr && \
    chown appuser:appuser /app/discovr

# Debugging step: verify binary
RUN ls -l /app/discovr && \
    file /app/discovr

# Switch to non-root user
USER appuser

# Use a shell entrypoint to support argument passing
ENTRYPOINT ["sudo", "/app/discovr"]

