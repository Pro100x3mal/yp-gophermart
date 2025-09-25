# Gophermart — накопительная система лояльности

## Описание
HTTP‑сервис для регистрации пользователей, приема номеров заказов, учёта баллов лояльности, списаний и получения статусов начислений из внешней системы «accrual». Архитектура: handlers → services → repositories, зависимости внедряются через конструкторы.

- Хранилище: PostgreSQL
- Сжатие HTTP: поддержка gzip
- Аутентификация: JWT (cookie access_token)
- Валидация номеров заказов: алгоритм Луна
- Батч‑обработка заказов и ограничение RPS к accrual

## Сборка и запуск

### Вариант 1: go build
- сборка:
```shell script
go build -o ./cmd/gophermart/gophermart ./cmd/gophermart
```

- запуск:
```shell script
./cmd/gophermart/gophermart
```

### Вариант 2: go run
- запуск:
```shell script
go run ./cmd/gophermart
```

## Конфигурация

Ниже приведены параметры конфигурации с типами и значениями по умолчанию. Значения можно задавать как флагами командной строки, так и переменными окружения. Переменные окружения имеют приоритет над флагами.

### Флаги командной строки и переменные окружения

| Флаг | Переменная окружения | Тип | Значение по умолчанию | Описание |
|---|---|---|---|---|
| -l | LOG_LEVEL | string | info | Уровень логирования (info, debug, warn, error) |
| -a | RUN_ADDRESS | string | localhost:8080 | Адрес HTTP‑сервера (host:port) |
| -d | DATABASE_URI | string | — | DSN PostgreSQL (обязателен) |
| -r | ACCRUAL_SYSTEM_ADDRESS | string | — | Адрес внешней системы начислений |
| -s | SECRET | string | development-secret-change-me | Секретный ключ для JWT |
| -b | BATCH_SIZE | int | 10 | Размер батча запросов к accrual |
| -n | RATE_LIMIT | int | 5 | Лимит запросов/сек к accrual |
| -t | TOKEN_TTL | int (часы) | 24 | Время жизни JWT |
| -i | POLL_INTERVAL | int (секунды) | 1 | Интервал опроса accrual воркером (сек) |

Пример запуска с флагами:
```shell script
./gophermart -l info -a localhost:8080 -d "postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable" -r "http://localhost:8081" -s "change-me" -t 24
```

Пример запуска с переменными окружения:
```shell script
LOG_LEVEL=info \
RUN_ADDRESS=localhost:8080 \
DATABASE_URI="postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable" \
ACCRUAL_SYSTEM_ADDRESS="http://localhost:8081" \
SECRET="change-me" \
TOKEN_TTL=24 \
./gophermart
```

