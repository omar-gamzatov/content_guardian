# Makefile — Content Guardian

SHELL := /bin/bash

# -------- Variables --------
PROJECT           ?= content-guardian
REGISTRY          ?= local
API_IMAGE         ?= $(REGISTRY)/guardian-api:dev
ML_IMAGE          ?= $(REGISTRY)/guardian-ml:dev
COMPOSE_FILE      ?= ./docker-compose.yml
COMPOSE           ?= docker compose -f $(COMPOSE_FILE)

GO                ?= go
PKG               ?= ./...
API_CMD           ?= ./cmd/api

# Runtime env (локальный запуск)
REDIS_ADDR        ?= localhost:6379
POSTGRES_DSN      ?= postgres://postgres:postgres@localhost:5432/guardian?sslmode=disable
NATS_URL          ?= nats://localhost:4222
ML_URL            ?= http://localhost:8000
POLICY_PATH       ?= ./configs/policies/default/policy.v1.yaml

# Migrations (golang-migrate)
DB_DSN            ?= $(POSTGRES_DSN)
MIGRATIONS_DIR    ?= ./migrations
MIGRATE_DOCKER    ?= docker run --rm --network host -v $(PWD)/migrations:/migrations migrate/migrate:latest
MIGRATE_BIN       ?= migrate

# Python lint/format (опционально)
RUFF              ?= ruff

# -------- Help --------
.PHONY: help
help: ## Показать доступные команды
	@echo "Usage: make <target>"
	@echo
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n"} /^[a-zA-Z0-9_%-]+:.*?##/ { printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# -------- Go: deps/build/test --------
.PHONY: tidy
tidy: ## go mod tidy
	$(GO) mod tidy

.PHONY: build-api
build-api: ## Сборка API бинарника
	$(GO) build -o ./bin/api $(API_CMD)

.PHONY: test
test: ## Запуск unit-тестов Go
	$(GO) test -race -count=1 $(PKG)

.PHONY: fmt
fmt: ## Форматирование кода (Go + Python, если есть)
	$(GO) fmt $(PKG)
	-$(RUFF) format ml || true

.PHONY: lint
lint: ## Линтинг (golangci-lint + ruff при наличии)
	-which golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, skipping"
	-$(RUFF) check ml || true

# -------- Local run (без Docker) --------
.PHONY: run-api
run-api: ## Локальный запуск API (без Docker)
	REDIS_ADDR=$(REDIS_ADDR) \
	POSTGRES_DSN='$(POSTGRES_DSN)' \
	NATS_URL='$(NATS_URL)' \
	ML_URL='$(ML_URL)' \
	POLICY_PATH='$(POLICY_PATH)' \
	$(GO) run $(API_CMD)

# -------- Docker images --------
.PHONY: docker-build-api
docker-build-api: ## Собрать Docker-образ API
	docker build -t $(API_IMAGE) ./cmd/api

.PHONY: docker-build-ml
docker-build-ml: ## Собрать Docker-образ ML
	docker build -t $(ML_IMAGE) ./ml

.PHONY: docker-push
docker-push: ## Запушить образы (нужен логин в реестр)
	docker push $(API_IMAGE)
	docker push $(ML_IMAGE)

# -------- Docker Compose env --------
.PHONY: up
up: ## Запуск локального стенда (db, redis, nats, ml, api)
	$(COMPOSE) up -d --build

.PHONY: down
down: ## Остановка и удаление контейнеров
	$(COMPOSE) down -v

.PHONY: logs
logs: ## Логи всех сервисов
	$(COMPOSE) logs -f --tail=200

.PHONY: ps
ps: ## Список сервисов
	$(COMPOSE) ps

.PHONY: restart
restart: ## Перезапуск API и ML
	$(COMPOSE) up -d --no-deps --build api ml

.PHONY: restart-api
restart-api: ## Перезапуск API
	$(COMPOSE) up -d --no-deps --build api

.PHONY: restart-ml
restart-ml: ## Перезапуск ML
	$(COMPOSE) up -d --no-deps --build ml

.PHONY: restart-web-client
restart-web-client: ## Перезапуск ML
	$(COMPOSE) up -d --no-deps --build web-client

# -------- DB: migrations --------
# Используем migrate CLI в контейнере (по умолчанию). Можно переключить на локальный бинарь MIGRATE_BIN.
.PHONY: migrate-up
migrate-up: ## Применить все миграции (up)
	$(MIGRATE_DOCKER) -path=/migrations -database '$(DB_DSN)' up

.PHONY: migrate-down
migrate-down: ## Откатить одну миграцию (down 1)
	$(MIGRATE_DOCKER) -path=/migrations -database '$(DB_DSN)' down 1

.PHONY: migrate-force
migrate-force: ## Принудительно выставить версию (MIGRATE_VERSION=?)
	@if [ -z "$(MIGRATE_VERSION)" ]; then echo "Set MIGRATE_VERSION=<int>"; exit 1; fi
	$(MIGRATE_DOCKER) -path=/migrations -database '$(DB_DSN)' force $(MIGRATE_VERSION)

.PHONY: migrate-create
migrate-create: ## Создать шаблон миграции (NAME=descr)
	@if [ -z "$(NAME)" ]; then echo "Set NAME=<snake_case_name>"; exit 1; fi
	@ts=$$(date +%Y%m%d%H%M%S); \
	touch $(MIGRATIONS_DIR)/$${ts}_$(NAME).up.sql; \
	touch $(MIGRATIONS_DIR)/$${ts}_$(NAME).down.sql; \
	echo "Created: $(MIGRATIONS_DIR)/$${ts}_$(NAME).up.sql and .down.sql"

# -------- Data: seed --------
.PHONY: seed
seed: ## Засеять тестовые данные (scripts/dev_seed.sh)
	./scripts/dev_seed.sh

# -------- OpenAPI --------
.PHONY: openapi-validate
openapi-validate: ## Валидация OpenAPI (требуется swagger-cli)
	-which swagger-cli >/dev/null 2>&1 && swagger-cli validate schemas/openapi.yaml || echo "swagger-cli not installed, skipping"

# -------- Clean --------
.PHONY: clean
clean: ## Очистка сборок
	rm -rf ./bin
