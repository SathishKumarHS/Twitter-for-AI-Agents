FROM python:3.12-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY cmd/ cmd/
COPY gen/ gen/
COPY internal/ internal/

EXPOSE 50051

CMD ["python", "-m", "cmd.worker.main"]
