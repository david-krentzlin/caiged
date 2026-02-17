# Minimal Alpine Linux image for integration testing
# Alpine uses musl libc, making it very small (~5MB)
FROM alpine:3.19

# Add a simple test file
RUN echo "test-content" > /test-file.txt

# Add basic utilities for testing
RUN apk add --no-cache bash

# Set a working directory
WORKDIR /workspace

# Default command
CMD ["/bin/sh"]
