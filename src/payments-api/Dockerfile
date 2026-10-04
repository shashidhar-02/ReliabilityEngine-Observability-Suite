FROM python:3.11-alpine AS builder

WORKDIR /app
RUN python -m venv /opt/venv
COPY requirements.lock .
RUN /opt/venv/bin/python -m pip install --no-cache-dir --only-binary=:all: \
    --require-hashes -r requirements.lock && \
    /opt/venv/bin/python -m pip uninstall -y pip setuptools wheel

FROM python:3.11-alpine

WORKDIR /app
RUN apk upgrade --no-cache && apk add --no-cache libstdc++ ca-certificates && \
    python -m pip uninstall -y pip setuptools wheel && \
    adduser -D -u 1000 appuser
COPY --from=builder /opt/venv /opt/venv
COPY main.py .
ENV PATH="/opt/venv/bin:$PATH" PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1
USER 1000:1000
EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
