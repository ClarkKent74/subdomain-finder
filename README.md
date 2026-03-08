# Subdomain Finder

Сервис поиска поддоменов. Принимает домен и алгоритм сканирования, возвращает список найденных поддоменов.

## Алгоритмы

| Алгоритм | Описание |
|---|---|
| `passive` | Ищет поддомены в публичных Certificate Transparency логах через crt.sh. Не создаёт трафик к цели. |
| `bruteforce` | Перебирает популярные имена поддоменов через DNS запросы. Поддерживает загрузку своего словаря. |
| `zonetransfer` | Запрашивает полную копию DNS зоны через AXFR. Работает только на неправильно настроенных серверах. |

## Быстрый старт

Нужен только Docker.

```bash
# 1. Клонировать репозиторий
git clone https://github.com/ClarkKent74/subdomain-finder
cd subdomain-finder

# 2. Создать конфиг
cp .env.example .env

# 3. Запустить
sudo docker compose up -d
```

Сервис запустится на `http://localhost:8080`.

Swagger UI: `http://localhost:8080/swagger/index.html`

## API

### POST /findDomains

Создаёт задачу на поиск поддоменов. Возвращает `201` сразу — сканирование выполняется асинхронно.

**Параметры (multipart/form-data):**

| Параметр | Тип | Обязательный | Описание |
|---|---|---|---|
| `domain` | string | да | Целевой домен |
| `algorithm` | string | да | Алгоритм: `passive`, `bruteforce`, `zonetransfer` |
| `wordlist` | file | нет | Файл словаря для bruteforce (одно слово на строку) |

**Примеры:**

```bash
# Пассивный поиск
curl -X POST http://localhost:8080/findDomains \
  -F "domain=example.com" \
  -F "algorithm=passive"

# Брутфорс со встроенным словарём
curl -X POST http://localhost:8080/findDomains \
  -F "domain=example.com" \
  -F "algorithm=bruteforce"

# Брутфорс со своим словарём
curl -X POST http://localhost:8080/findDomains \
  -F "domain=example.com" \
  -F "algorithm=bruteforce" \
  -F "wordlist=@/path/to/mylist.txt"

# Zone Transfer
curl -X POST http://localhost:8080/findDomains \
  -F "domain=example.com" \
  -F "algorithm=zonetransfer"
```

**Ответ `201`:**
```json
{
  "domain": "example.com",
  "algorithm": "passive",
  "status": "pending"
}
```

---

### GET /getResult

Возвращает статус и результаты задачи.

**Параметры (query):**

| Параметр | Тип | Описание |
|---|---|---|
| `domain` | string | Целевой домен |
| `algorithm` | string | Алгоритм |

**Пример:**

```bash
curl "http://localhost:8080/getResult?domain=example.com&algorithm=passive"
```

**Ответ `202` — сканирование ещё выполняется:**
```json
{
  "domain": "example.com",
  "algorithm": "passive",
  "status": "running"
}
```

**Ответ `200` — сканирование завершено:**
```json
{
  "domain": "example.com",
  "algorithm": "passive",
  "status": "done",
  "results": [
    "api.example.com",
    "dev.example.com",
    "mail.example.com"
  ]
}
```

**Ответ `200` — сканирование завершено с ошибкой:**
```json
{
  "domain": "example.com",
  "algorithm": "zonetransfer",
  "status": "failed",
  "error": "все NS серверы отклонили zone transfer"
}
```

---

### Коды ответов

| Код | Описание |
|---|---|
| `201` | Задача создана |
| `200` | Результат готов |
| `202` | Сканирование ещё выполняется |
| `400` | Неверные параметры |
| `404` | Задача не найдена |
| `409` | Такая задача уже выполняется |
| `429` | Превышен rate limit или очередь переполнена |

## Словарь для брутфорса

Файл словаря — обычный текстовый файл, одно слово на строку. Строки начинающиеся с `#` и пустые строки игнорируются.

```text
# Мой словарь
www
api
dev
staging
admin
```

Если словарь не передан — используется встроенный список из ~70 популярных имён.

## Конфигурация

Все параметры задаются через `.env` файл. Полный список с описанием — в `.env.example`.

| Переменная | Дефолт | Описание |
|---|---|---|
| `PORT` | `8080` | Порт сервера |
| `REDIS_URL` | `redis://localhost:6379` | Адрес Redis |
| `WORKER_POOL_SIZE` | `10` | Количество воркеров |
| `WORKER_QUEUE_SIZE` | `100` | Размер очереди задач |
| `WORKER_SCAN_TIMEOUT` | `5m` | Таймаут одного сканирования |
| `STORE_TASK_TTL` | `24h` | Время хранения результатов |
| `STORE_MAX_TASKS` | `10000` | Максимум задач в хранилище |
| `RATE_LIMIT_RPM` | `20` | Запросов с одного IP в минуту |
| `DNS_RESOLVER` | `8.8.8.8:53` | DNS резолвер для брутфорса |
| `DNS_WORKER_COUNT` | `50` | Параллельных DNS запросов |
| `SCANNER_PASSIVE_TIMEOUT` | `30s` | Таймаут запроса к crt.sh |
| `SCANNER_CB_THRESHOLD` | `5` | Ошибок до открытия circuit breaker |
| `SCANNER_CB_TIMEOUT` | `30s` | Время ожидания перед повторной попыткой |


Лицензия
Проект распространяется под лицензией GNU General Public License v3.0.