FROM python:3.11-slim AS base

# Системные зависимости (spaCy, presidio, torch могут требовать gcc, build-essential и т.д.)
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential curl git wget \
    && rm -rf /var/lib/apt/lists/*

# Создаем директорию для cache
RUN mkdir -p /home/appuser/.cache/torch/hub/checkpoints

# Скачиваем файл напрямую при сборке
RUN wget -O /home/appuser/.cache/torch/hub/checkpoints/multilingual_debiased-0b549669.ckpt \
    "https://github.com/unitaryai/detoxify/releases/download/v0.4-alpha/multilingual_debiased-0b549669.ckpt"

# Рабочая директория
WORKDIR /app

# Ускоряем pip и отключаем кэш
ENV PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

# Устанавливаем зависимости
# В requirements.txt уже пинним версии (см. ниже)
COPY requirements.txt /app/requirements.txt
RUN pip install --upgrade pip && \
    pip install -r /app/requirements.txt

# Заранее скачиваем spaCy модель, чтобы не делать это при старте контейнера
# Модель маленькая для снижения размера (en_core_web_sm)
RUN python -m spacy download en_core_web_sm-3.8.0
RUN python -m spacy download en_core_web_lg-3.8.0

# Создаём непривилегированного пользователя
RUN useradd -m -u 10001 appuser
USER appuser

# Копируем приложение
COPY --chown=appuser:appuser app.py /app/app.py

# Опционально: прогреваем модели на этапе сборки (уменьшит первый latency)
# Команда ниже закомментирована, потому что прогрев в CI может занять время.
# RUN python - <<'PY'\nfrom app import tox\nprint('Detoxify loaded')\nPY

EXPOSE 8000

# HEALTHCHECK для docker-compose (альтернатива — проверка оркестратором)
# В slim-образе нет curl; используем python + requests
# Добавим tiny http checker:
RUN pip install --no-cache-dir requests
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD python -c "import requests; import os; import sys; \
    url=f'http://127.0.0.1:{os.getenv(\"PORT\", \"8000\")}/docs'; \
    sys.exit(0 if requests.get(url, timeout=3).status_code<500 else 1)"

# Переменные окружения
ENV HOST=0.0.0.0 \
    PORT=8000 \
    UVICORN_WORKERS=1 \
    UVICORN_LOG_LEVEL=info

# Запуск uvicorn
CMD ["bash", "-lc", "exec uvicorn app:app --host ${HOST} --port ${PORT} --workers ${UVICORN_WORKERS} --log-level ${UVICORN_LOG_LEVEL}"]
