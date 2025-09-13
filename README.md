```markdown
# Wildberries Order Service

Микросервис для обработки заказов в стиле Wildberries с использованием **Go**, **PostgreSQL**, **Kafka** и **in-memory кэширования**.

---

## 📦 Архитектура проекта

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

````

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
docker exec -it myapp_kafka kafka-topics --create --topic orders --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
```

### Отправка тестового сообщения

```bash
echo '{ "order_uid": "b563feb7b2b84b6test", ... }' | docker exec -i myapp_kafka kafka-console-producer --topic orders --bootstrap-server localhost:9092
```

---

## ⚙️ Конфигурация (.env)

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

## 📝 Особенности реализации

* Thread-safe in-memory кэш с прогревом
* Валидация данных и graceful shutdown
* Транзакции для целостности данных
* Connection pooling, индексы, пагинация
* Асинхронная обработка Kafka сообщений

---

## 📊 Мониторинг

* Kafka UI: [http://localhost:8080](http://localhost:8080)
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

### Kafka

* Проверить статус: `docker-compose ps`
* Логи: `docker-compose logs kafka`
* Список топиков: `docker exec -it myapp_kafka kafka-topics --list --bootstrap-server localhost:9092`

### База данных

* Подключение: `docker exec -it myapp_postgres psql -U myapp_user -d myapp_db`
* Проверка миграций: `\dt`

### Кэш

* Статистика: `curl http://localhost:8081/api/v1/cache/stats`
* Прогрев кэша: `curl -X POST http://localhost:8081/api/v1/cache/warmup`


# orderservice
# orderservice
