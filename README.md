# Content Guardian

Сервис автоматической модерации контента для обработки текстовых данных, изображений, видео и аудио. Масштабируемый, безопасный и экономичный solution для модерации пользовательского контента.

## 🚀 Возможности (предполагаемые)

- **Текстовая модерация** - обнаружение нежелательного текстового контента
- **Анализ изображений** - модерация визуального контента
- **Обработка видео и аудио** - анализ медиафайлов
- **Гибкая система правил** - настраиваемые политики модерации
- **Human-in-the-Loop** - интеграция с модераторами для сложных случаев
- **Масштабируемость** - обработка больших объемов контента
- **Мониторинг и observability** - полный контроль над работой системы

## 🛠 Технологический стек (предаврительно)

- **Backend**: Go, Node.js
- **Базы данных**: PostgreSQL, Redis
- **Мессенджер**: NATS JetStream / Kafka
- **Хранилище**: S3-совместимое хранилище
- **Контейнеризация**: Docker, Kubernetes
- **Мониторинг**: Prometheus, Grafana, OpenTelemetry
- **Оркестрация**: Kubernetes

## 📦 Структура проекта

```
content_guardian/
├── apps/
│   ├── api-gateway/          # Go API Gateway
│   └── rules-engine/         # Node.js Rules Engine
├── infrastructure/
│   ├── docker-compose.yml    # Локальная разработка
│   └── kubernetes/           # K8s манифесты
├── docs/
│   └── api-spec/             # OpenAPI спецификации
└── scripts/                  # Вспомогательные скрипты
```

## 🚀 Быстрый старт

### Предварительные требования

- Docker 20.10+
- Docker Compose 2.0+
- Go 1.20+ (для разработки)
- Node.js 20+ (для разработки)

### Запуск в Docker

```bash
# Клонирование репозитория
git clone <repository-url>
cd content_guardian

# Запуск сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps
```

Сервисы будут доступны:
- API Gateway: http://localhost:8080
- Rules Engine: http://localhost:3000
- PostgreSQL: localhost:5432
- Redis: localhost:6379

### Локальная разработка

```bash
# Запуск только инфраструктуры
docker-compose up -d postgres redis nats

# Запуск API Gateway (Go)
cd apps/api-gateway
go run main.go

# Запуск Rules Engine (Node.js)
cd apps/rules-engine
npm install
npm run dev
```

## 📡 API Endpoints

### Модерация текста

```http
POST /v1/moderate/text
Content-Type: application/json

{
  "content": "Текст для модерации",
  "contentId": "uuid",
  "userId": "user-uuid",
  "policy": "strict"
}
```

Ответ:
```json
{
  "approved": true,
  "moderationId": "mod-uuid",
  "reasons": [],
  "score": 0.1
}
```

### Статус сервиса

```http
GET /healthz
```

## 🔧 Конфигурация

Настройки выполняются через environment variables:

```bash
# API Gateway
export API_PORT=8080
export DB_URL="postgresql://user:pass@postgres:5432/content_guardian"
export NATS_URL="nats://nats:4222"

# Rules Engine
export RULES_PORT=3000
export REDIS_URL="redis://redis:6379"
```

## 🧪 Тестирование

```bash
# Запуск тестов API Gateway
cd apps/api-gateway
go test ./...

# Запуск тестов Rules Engine
cd apps/rules-engine
npm test

# Интеграционные тесты
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

<!-- ## 📊 Мониторинг

- **Metrics**: Prometheus на порту 9090
- **Dashboard**: Grafana на порту 3001
- **Tracing**: Jaeger на порту 16686
- **Logs**: Loki на порту 3100

## 🔒 Безопасность

- Все данные в движении шифруются (TLS)
- Минимизация хранимых PII данных
- Ролевая модель доступа (RBAC)
- Аудит и логирование всех операций

## 🗺 Roadmap

1. **MVP (Текст)** - Базовая текстовая модерация
2. **Изображения** - Модерация визуального контента
3. **Видео/Аудио** - Обработка медиафайлов
4. **Политики** - Гибкая система правил
5. **Human-in-the-Loop** - Интеграция с модераторами
6. **Observability** - Мониторинг и метрики
7. **Оптимизация** - Снижение стоимости эксплуатации -->

<!-- ## 🤝 Участие в разработке

1. Форкните репозиторий
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Закоммитьте изменения (`git commit -m 'Add amazing feature'`)
4. Запушьте branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request -->

## 📝 Лицензия

Этот проект лицензирован под MIT License - см. файл [LICENSE](LICENSE) для деталей.
<!-- 
## 🆘 Поддержка

Если у вас возникли вопросы:
- Создайте [Issue](https://github.com/your-org/content_guardian/issues)
- Напишите на email: support@content-guardian.com
- Присоединяйтесь к нашему [Discord](https://discord.gg/content-guardian)

---

**Content Guardian** - защищаем ваше сообщество от нежелательного контента! -->
