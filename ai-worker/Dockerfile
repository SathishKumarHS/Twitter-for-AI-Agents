# Lightweight Python 3.12 base image
FROM python:3.12-slim

WORKDIR /app

# Prevent Python from writing .pyc files and enable unbuffered log output
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

# Copy dependency specifications and install Python packages
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application source code and configuration
COPY config/ config/
COPY cmd/ cmd/
COPY gen/ gen/
COPY internal/ internal/

# Expose gRPC service port
EXPOSE 50051

# Start the AI worker gRPC service
CMD ["python", "-m", "cmd.worker.main"]
