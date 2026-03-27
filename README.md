# Subdomain Finder

Сервис поиска поддоменов для на основе [Sudomy](https://github.com/screetsec/Sudomy). Принимает домен, запускает сканирование через множество источников одновременно и возвращает список найденных поддоменов.

## Как это работает

Сервис использует Sudomy который параллельно опрашивает 22 источника:
crt.sh, DNSdumpster, HackerTarget, Certspotter, Webarchive, CommonCrawl, AlienVault, RapidDNS, UrlScan и другие.

При наличии API ключей подключаются платные источники: Shodan, VirusTotal, Censys, SecurityTrails, BinaryEdge.

## Быстрый старт

Нужен только Docker.

```bash
# 1. Клонировать репозиторий
git clone https://github.com/твой_ник/subdomain-finder.git
cd subdomain-finder

# 2. Создать конфиг
cp .env.example .env

# 3. Указать Redis URL для docker compose
# В .env изменить:
# REDIS_URL=redis://redis:6379

# 4. Запустить
sudo docker compose up -d
```

Сервис запустится на `http://localhost:8080`.

Swagger UI: `http://localhost:8080/swagger/index.html`

## Локальная разработка

```bash
go mod tidy

go install github.com/swaggo/swag/cmd/swag@latest
export PATH=$PATH:$(go env GOPATH)/bin
swag init -g cmd/server/main.go

sudo docker compose up redis -d
go run cmd/server/main.go
```

## API

### POST /findDomains

Создаёт задачу на поиск поддоменов. Возвращает `201` сразу — сканирование выполняется асинхронно.

**Параметры (query):**

| Параметр | Тип | Обязательный | Описание |
|---|---|---|---|
| `domain` | string | да | Целевой домен |

**Пример:**

```bash
curl -X POST "http://localhost:8080/findDomains?domain=example.com"
```

**Ответ `201`:**
```json
{
  "domain": "example.com",
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

**Пример:**

```bash
curl "http://localhost:8080/getResult?domain=example.com"
```

**Ответ `202` — сканирование ещё выполняется:**
```json
{
  "domain": "example.com",
  "status": "running"
}
```

**Ответ `200` — сканирование завершено:**
```json
{
  "domain": "example.com",
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
  "status": "failed",
  "error": "сканирование Sudomy прервано: context deadline exceeded"
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

## Конфигурация

Все параметры задаются через `.env` файл.

| Переменная | Дефолт | Описание |
|---|---|---|
| `PORT` | `8080` | Порт сервера |
| `REDIS_URL` | `redis://localhost:6379` | Адрес Redis |
| `WORKER_POOL_SIZE` | `10` | Количество воркеров |
| `WORKER_QUEUE_SIZE` | `100` | Размер очереди задач |
| `WORKER_SCAN_TIMEOUT` | `35m` | Таймаут одного сканирования |
| `STORE_TASK_TTL` | `24h` | Время хранения результатов |
| `STORE_MAX_TASKS` | `10000` | Максимум задач в хранилище |
| `RATE_LIMIT_RPM` | `20` | Запросов с одного IP в минуту |
| `SUDOMY_PATH` | `/usr/lib/sudomy/sudomy` | Путь к Sudomy |
| `SUDOMY_SCAN_TIMEOUT` | `30m` | Таймаут сканирования Sudomy |
| `SUDOMY_VIRUSTOTAL_KEY` | — | API ключ VirusTotal |
| `SUDOMY_SHODAN_KEY` | — | API ключ Shodan |
| `SUDOMY_CENSYS_KEY` | — | API ключ Censys |
| `SUDOMY_SECURITYTRAILS_KEY` | — | API ключ SecurityTrails |

## Лицензия

Проект распространяется под лицензией [GNU General Public License v3.0](LICENSE).

Ты можешь свободно использовать, изменять и распространять этот код при условии что производные работы также распространяются под GPL v3.0.
