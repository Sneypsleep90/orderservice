# Wildberries Order Service

Микросервис для приёма и обработки заказов в стиле Wildberries, реализованный на **Go**, с хранением в **PostgreSQL**, асинхронной обработкой через **Kafka** и высокопроизводительным **in-memory** кэшем.

---

## 🚀 Быстрый старт

1. Запустите инфраструктуру:

```bash
docker-compose up -d
docker-compose ps
```

2. Подготовьте базу данных (пример для PostgreSQL):

```sql
CREATE USER myapp_user WITH PASSWORD 'myapp_password';
CREATE DATABASE myapp_db OWNER myapp_user;
GRANT ALL PRIVILEGES ON DATABASE myapp_db TO myapp_user;
```

3. Установите зависимости и запустите приложение:

```bash
go mod tidy
go run cmd/main.go
```

Откройте веб-интерфейс: `http://localhost:8081`

---

## 📁 Структура проекта

```
myapp/
├─ cmd/
│  └─ main.go                 # Точка входа приложения
├─ internal/
│  ├─ model/                  # Модели данных
│  ├─ handlers/               # HTTP handlers
│  ├─ service/                # Бизнес-логика
│  ├─ repository/             # Доступ к БД
│  ├─ kafka/                  # Kafka consumer
│  ├─ cache/                  # in-memory cache
│  └─ migrate/                # Запуск миграций
├─ migrations/                # SQL миграции
├─ web/                       # Веб-интерфейс
├─ go.mod
├─ .env
├─ docker-compose.yaml
└─ README.md
```

---

## 🧩 Основные сущности

**Order**

* `order_uid` — уникальный идентификатор
* `track_number` — номер отслеживания
* `customer_id` — идентификатор клиента
* `date_created` — дата создания

**Delivery**

* Контактные данные: `name`, `phone`, `email`
* Адрес: `city`, `region`, `address`

**Payment**

* `transaction`, `amount`, `currency`, `provider`

**Item**

* `chrt_id`, `name`, `brand`, `price`, `total_price`

---

## 🧰 Интерфейсы (public contracts)

### Repository

```go
type Repository interface {
    CreateOrder(order *model.Order) error
    GetOrderByUID(orderUID string) (*model.Order, error)
    GetAllOrders() ([]*model.Order, error)
    UpdateOrder(order *model.Order) error
    DeleteOrder(orderUID string) error
}
```

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

## 🌐 HTTP API

### Основные

* `GET /order/{order_uid}` — получить заказ по UID
* `GET /api/v1/orders` — получить все заказы (поддерживает пагинацию)
* `PUT /api/v1/orders/{order_uid}` — обновить заказ
* `DELETE /api/v1/orders/{order_uid}` — удалить заказ

### Служебные

* `GET /health` — проверка здоровья сервиса
* `GET /api/v1/cache/stats` — статистика кэша
* `POST /api/v1/cache/warmup` — прогрев кэша

### Пример запроса

```bash
curl http://localhost:8081/order/b563feb7b2b84b6test
```

---

## 📨 Kafka

**Создание топика** (в контейнере Kafka):

```bash
docker exec -it myapp_kafka kafka-topics --create --topic orders --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
```

**Отправка тестового сообщения**:

```bash
echo '{ "order_uid": "b563feb7b2b84b6test", ... }' | docker exec -i myapp_kafka kafka-console-producer --topic orders --bootstrap-server localhost:9092
```

В приложении реализован consumer, который асинхронно обрабатывает входящие сообщения и сохраняет заказы в БД с записью в кэш.

---

## ⚙️ Конфигурация (`.env`)

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=myapp_user
DB_PASSWORD=myapp_password
DB_NAME=myapp_db
DB_SSLMODE=disable

KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=orders
KAFKA_GROUP_ID=order-service

SERVER_PORT=8081
```

---

## 🔧 Особенности реализации

* Thread-safe in-memory кэш с возможностью прогрева (Warmup)
* Валидация входящих данных
* Graceful shutdown с ожиданием завершения обработки Kafka сообщений
* Транзакции для целостности при работе с БД
* Connection pooling и индексация для ускорения выборок
* Пагинация при получении списков заказов
* Асинхронная обработка сообщений Kafka

---

## 📊 Мониторинг и отладка

* Kafka UI: `http://localhost:8080`
* Статистика кэша: `curl http://localhost:8081/api/v1/cache/stats`
* Health check: `curl http://localhost:8081/health`

---

## 🛠 Разработка

### Миграции

```bash
touch migrations/000002_add_index.up.sql
touch migrations/000002_add_index.down.sql
```

### Тесты

```bash
go test ./...
go test -cover ./...
```

### Линтинг

```bash
golangci-lint run
```

---

## ⚠️ Troubleshooting

**Kafka**

* Проверить статус контейнеров: `docker-compose ps`
* Просмотреть логи: `docker-compose logs kafka`
* Список топиков:
  `docker exec -it myapp_kafka kafka-topics --list --bootstrap-server localhost:9092`

**Postgres**

* Подключиться в контейнере: `docker exec -it myapp_postgres psql -U myapp_user -d myapp_db`
* Проверить таблицы: `\dt`

**Кэш**

* Статистика: `curl http://localhost:8081/api/v1/cache/stats`
* Прогрев: `curl -X POST http://localhost:8081/api/v1/cache/warmup`

---

## ✅ Рекомендации по улучшению

* Добавить metrics (Prometheus) и дашборд (Grafana)
* CI/CD для автоматического прогона миграций и тестов
* Добавить схемы валидации сообщений Kafka (JSON Schema / Protobuf)
* Rate limiting и RBAC для HTTP API

---

## 📝 Лицензия

MIT — свободное использование, модификация и распространение.

