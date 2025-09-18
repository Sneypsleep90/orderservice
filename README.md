```markdown
# Wildberries Order Service

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?logo=postgresql&logoColor=white)
![Kafka](https://img.shields.io/badge/Kafka-Confluent-231F20?logo=apachekafka&logoColor=white)
![Prometheus](https://img.shields.io/badge/Prometheus-metrics-E6522C?logo=prometheus&logoColor=white)
![OpenTelemetry](https://img.shields.io/badge/Tracing-OpenTelemetry-6E44FF?logo=opentelemetry&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-compose-2496ED?logo=docker&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

Микросервис для обработки заказов в стиле Wildberries с использованием **Go**, **PostgreSQL**, **Kafka** и **in-memory кэширования**.

---

## 📚 Содержание

- [Архитектура проекта (структура файлов)](#-архитектура-проекта-структура-файлов)
- [Основные сущности](#-основные-сущности)
- [Интерфейсы](#-интерфейсы)
- [Запуск проекта](#-запуск-проекта)
- [API Endpoints](#-api-endpoints)
- [Веб-интерфейс](#-веб-интерфейс)
- [Конфигурация](#-конфигурация-configenv)
- [Особенности реализации](#-особенности-реализации)
- [Мониторинг](#-мониторинг)
- [Разработка (миграции, тесты, линтинг)](#-разработка)
- [Полное руководство](#-полное-руководство-как-поднять-проект-и-отправить-данные)

---

## 📦 Архитектура проекта (структура файлов)

```
myapp/
├─ .golangci.yml
├─ config.env
├─ docker-compose.yaml
├─ go.mod
├─ go.sum
├─ main
├─ README.md
├─ cmd/
│  ├─ main.go
│  └─ producer/
│     └─ main.go
├─ internal/
│  ├─ cache/
│  │  └─ cache.go
│  ├─ config/
│  │  └─ config.go
│  ├─ database/
│  │  └─ database.go
│  ├─ handlers/
│  │  ├─ handler.go
│  │  └─ handler_test.go
│  ├─ kafka/
│  │  └─ consumer.go
│  ├─ migrate/
│  │  └─ migrate.go
│  ├─ model/
│  │  └─ order.go
│  ├─ repository/
│  │  ├─ repository.go
│  │  └─ repository_test.go
│  └─ service/
│     ├─ metrics.go
│     ├─ service.go
│     ├─ service_test.go
│     └─ tracing.go
├─ migrations/
│  ├─ 000001_create_orders.down.sql
│  ├─ 000001_create_orders.up.sql
│  └─ 000002_add_unique_indexes.up.sql
└─ web/
   └─ index.html
```

---

## ⚙️ Основные сущности

### Order
- `order_uid` — уникальный идентификатор
- `track_number` — номер отслеживания
- `customer_id` — идентификатор клиента
- `date_created` — дата создания

### Delivery
- Контактные данные: `name`, `phone`, `email`
- Адрес: `city`, `region`, `address`

### Payment
- `transaction`, `amount`, `currency`, `provider`

### Item
- `chrt_id`, `name`, `brand`, `price`, `total_price`

---

## 🧩 Интерфейсы

### Repository
```go
type Repository interface {
    CreateOrder(order *model.Order) error
    GetOrderByUID(orderUID string) (*model.Order, error)
    GetAllOrders() ([]*model.Order, error)
    UpdateOrder(order *model.Order) error
    DeleteOrder(orderUID string) error
}
````

### Cache

```go
type Cache interface {
    Set(orderUID string, order *model.Order)
    Get(orderUID string) (*model.Order, bool)
    Delete(orderUID string)
    GetAll() map[string]*model.Order
    Clear()
    Size() int
}
```

### Service

```go
type Service interface {
    ProcessOrder(order *model.Order) error
    GetOrderByUID(orderUID string) (*model.Order, error)
    GetAllOrders() ([]*model.Order, error)
    UpdateOrder(order *model.Order) error
    DeleteOrder(orderUID string) error
    GetCacheStats() cache.CacheStats
    WarmupCache() error
}
```

---

## 🚀 Запуск проекта

### 1. Инфраструктура (Docker Compose)

```bash
docker-compose up -d
docker-compose ps
```

### 2. Настройка базы данных

```sql
CREATE USER myapp_user WITH PASSWORD 'myapp_password';
CREATE DATABASE myapp_db OWNER myapp_user;
GRANT ALL PRIVILEGES ON DATABASE myapp_db TO myapp_user;
```

### 3. Запуск приложения

```bash
go mod tidy
go run cmd/main.go
```

Миграции применяются автоматически при старте. Для ручного запуска отдельной команды можно выполнить:

```bash
go run cmd/main.go up
```

---

## 🌐 API Endpoints

### Основные

* `GET /order/{order_uid}` — получить заказ по UID
* `GET /api/v1/orders` — получить все заказы
* `PUT /api/v1/orders/{order_uid}` — обновить заказ
* `DELETE /api/v1/orders/{order_uid}` — удалить заказ

### Служебные

* `GET /health` — проверка здоровья
* `GET /api/v1/cache/stats` — статистика кэша
* `POST /api/v1/cache/warmup` — прогрев кэша
* `GET /metrics` — Prometheus-метрики

### Пример запроса

```bash
curl http://localhost:8081/order/b563feb7b2b84b6test
```

---

## 🖥 Веб-интерфейс

Открыть: [http://localhost:8081](http://localhost:8081)

---

## 📨 Kafka

### Создание топика

```bash
docker exec -it myapp_kafka \
  kafka-topics --create --topic orders --bootstrap-server localhost:9092 \
  --partitions 1 --replication-factor 1
```

### Отправка тестового сообщения

```bash
go run cmd/producer/main.go b563feb7b2b84b6test
```

---

## ⚙️ Конфигурация (`config.env`)

```env
DB_HOST=localhost
DB_PORT=5433
DB_USER=myapp_user
DB_PASSWORD=myapp_password
DB_NAME=myapp_db
DB_SSLMODE=disable

KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=orders
KAFKA_GROUP_ID=order-service
KAFKA_DLQ_TOPIC=orders-dlq

SERVER_PORT=8081

# Кэш: memory | lru
CACHE_TYPE=lru
CACHE_LRU_SIZE=1000

# Миграции
MIGRATIONS_PATH=./migrations
SKIP_MIGRATIONS=false
```

Приложение читает файл `config.env` через `godotenv` при старте.

---

## 📝 Особенности реализации

* Thread-safe кэш: in-memory и LRU (ограничение памяти)
* Валидация входящих данных с помощью `go-playground/validator`
* Транзакции для целостности данных; индексы, upsert-логика
* Kafka consumer с retry/backoff и DLQ (dead-letter queue)
* Prometheus-метрики (`/metrics`), healthcheck `/health`
* Прогрев кэша при старте, graceful shutdown

---

## 📊 Мониторинг

* Kafka UI: [http://localhost:8080](http://localhost:8080)
* Статистика кэша: `curl http://localhost:8081/api/v1/cache/stats`
* Health check: `curl http://localhost:8081/health`
* Метрики: `curl http://localhost:8081/metrics` (Prometheus)

---

## 🛠 Разработка

### Миграции

```bash
# Автоматически применяются при старте приложения
go run cmd/main.go up
```

### Тесты

Быстрый прогон всех тестов:

```bash
go test ./...
```

С покрытием/детально:

```bash
go test -race -v -cover ./...
```

HTML-отчёт покрытия:

```bash
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

Без кеша компиляции:

```bash
go test -count=1 ./...
```

### Трейсинг

Включён OpenTelemetry с stdout‑экспортёром. Трейсы выводятся в stdout приложения. Для интеграции с внешними бэкендами (OTLP/Jaeger) замените экспортёр в `internal/service/tracing.go`.

### Линтинг и форматирование

```bash
# goimports: автоупорядочивание импортов
go install golang.org/x/tools/cmd/goimports@latest
$(go env GOPATH)/bin/goimports -w .

# golangci-lint (локально при наличии)
golangci-lint run
```

### Генерация тестовых данных

Для генерации данных в продюсере используется пакет `gofakeit`.

```bash
go run cmd/producer/main.go <order_uid>
```

## 📚 Полное руководство: как поднять проект и отправить данные

Ниже — пошаговая инструкция «с нуля» до получения заказа через HTTP.

1) Запустите Docker Desktop

- Убедитесь, что Docker запущен локально.

2) Поднимите инфраструктуру (Postgres, Zookeeper, Kafka, Kafka UI)

```bash
cd /Users/<ваш_пользователь>/myapp
docker-compose up -d
docker-compose ps
```

- Если порт 5432 занят на хосте, в compose уже проброшен порт Postgres на 5433. Проверьте `config.env` → `DB_PORT=5433`.

3) Проверьте/настройте конфигурацию приложения

- Файл: `config.env` (загружается автоматически)
- Минимально проверьте:

```env
DB_HOST=localhost
DB_PORT=5433
DB_USER=myapp_user
DB_PASSWORD=myapp_password
DB_NAME=myapp_db

KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=orders
KAFKA_GROUP_ID=order-service
KAFKA_DLQ_TOPIC=orders-dlq

SERVER_PORT=8081
CACHE_TYPE=lru
```

4) Установите зависимости и запустите приложение

```bash
go mod tidy
go run cmd/main.go
```

- Приложение при старте применит миграции и прогреет кэш.
- Эндпоинт здоровья: `http://localhost:8081/health`

5) Создайте топик Kafka (если авто‑создание отключено)

```bash
docker exec -it myapp_kafka \
  kafka-topics --create --topic orders --bootstrap-server localhost:9092 \
  --partitions 1 --replication-factor 1
```

6) Отправьте заказ (продюсер на Go с генерацией данных)

```bash
go run cmd/producer/main.go b563feb7b2b84b6test
```

- Продюсер сгенерирует валидные данные и отправит их в топик `orders`.

7) Проверьте, что заказ обработан

```bash
curl http://localhost:8081/order/b563feb7b2b84b6test
```

- В ответе должен прийти JSON заказа. Если 404 — дождитесь обработки consumer'ом (обычно <1–2 сек) и повторите запрос.

8) Посмотрите дополнительную диагностику/метрики

- Здоровье: `curl http://localhost:8081/health`
- Кэш: `curl http://localhost:8081/api/v1/cache/stats`
- Метрики: `curl http://localhost:8081/metrics`
- Kafka UI: `http://localhost:8080`

9) Типичные проблемы и решения

- Порт 5432 занят → используйте `DB_PORT=5433` (по умолчанию уже так настроено), compose пробрасывает `5433:5432`.
- «Order not found» сразу после отправки → подождите 1–2 сек, consumer обработает сообщение и выполнит upsert в БД; повторите GET.
- Ошибки миграций про `ON CONFLICT` → убедитесь, что применены миграции индексов (`migrations/000002_add_unique_indexes.up.sql`). Приложение применяет их автоматически на старте.
- Kafka не отвечает → проверьте `docker-compose ps`, логи `docker-compose logs kafka`.

10) Полезные команды

```bash
# Проверить последние заказы в БД
docker exec -i myapp_postgres \
  psql -U myapp_user -d myapp_db -c \
  "SELECT order_uid, date_created FROM orders ORDER BY date_created DESC LIMIT 5;"

# Принудительно прогреть кэш
curl -X POST http://localhost:8081/api/v1/cache/warmup
```

---

## ⚠️ Troubleshooting

### Kafka

* Проверить статус: `docker-compose ps`
* Логи: `docker-compose logs kafka`
* Список топиков: `docker exec -it myapp_kafka kafka-topics --list --bootstrap-server localhost:9092`

### База данных

* Подключение: `docker exec -it myapp_postgres psql -U myapp_user -d myapp_db`
* Порт на хосте: `5433`
* Проверка миграций: `\dt`

### Кэш

* Статистика: `curl http://localhost:8081/api/v1/cache/stats`
* Прогрев кэша: `curl -X POST http://localhost:8081/api/v1/cache/warmup`

### DLQ/Retry

* Основной топик: `orders`
* DLQ-топик: `orders-dlq`
* При ошибках потребления сообщения попадают в DLQ после нескольких попыток


# orderservice
# orderservice
# orderservice
