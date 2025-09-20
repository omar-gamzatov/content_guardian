.SILENT: help
.DEFAULT_GOAL := help

help:
	echo "Targets: dev, build, run-api, run-rules, fmt, test"

dev:
	docker compose -f deploy/compose/docker-compose.yml up --build

build:
	docker build -f Dockerfile.api -t cg/api:dev .
	docker build -f Dockerfile.rules -t cg/rules:dev .

run-api:
	go run ./apps/api-gateway

run-rules:
	cd apps/rules-engine && npm ci && npm run dev

fmt:
	gofmt -s -w .

test:
	go test ./...
