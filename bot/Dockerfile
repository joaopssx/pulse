FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONPATH=/app

RUN groupadd -r pulsebot \
    && useradd -r -u 10001 -g pulsebot -m -d /home/pulsebot pulsebot

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY src ./src

RUN chown -R pulsebot:pulsebot /app

USER pulsebot

EXPOSE 5000

CMD ["python", "-m", "src.main"]
