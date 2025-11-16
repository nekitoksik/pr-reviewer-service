# PR Reviewer Service

Сервис назначения ревьюверов для Pull Request'ов внутри команды.  
Позволяет управлять командами и участниками, автоматически назначать ревьюверов на PR, выполнять переназначение и получать список PR'ов по конкретному пользователю.

## Функциональность

- Управление командами:
  - создание/обновление команды с участниками (`POST /team/add`);
  - получение состава команды (`GET /team/get`)
- Управление пользователями:
  - установка флага активности `is_active` (`POST /users/setIsActive`);
  - получение PR'ов, где пользователь назначен ревьювером (`GET /users/getReview`)
- Работа с Pull Request:
  - создание PR c автоматическим назначением до двух активных ревьюверов из команды автора, исключая самого автора (`POST /pullRequest/create`);
  - merge PR c идемпотентным поведением (`POST /pullRequest/merge`);
  - переназначение одного ревьювера на случайного активного участника из команды заменяемого ревьювера (`POST /pullRequest/reassign`)
- Статистика:
  - `GET /stats` — агрегированная статистика по количеству PR и количеству назначений по ревьюверам
- Health-check:
  - `GET /health` — проверка живости сервиса

Все HTTP-ручки описаны в `openapi.yml` в корне проекта.

## Архитектура

Сервис реализован на Go и разделён на слои:

- `internal/db` — инициализация подключения к PostgreSQL, запуск миграций
- `internal/repo` — доступ к данным (Teams, Users, PullRequests, Stats) поверх `pgxpool`
- `internal/service` — бизнес-логика:
  - назначение и переназначение ревьюверов;
  - обработка статусов PR (OPEN/MERGED) и гарантия идемпотентного merge;
  - построение статистики
- `internal/transport/http` — HTTP-слой на gin: роутер, хендлеры, DTO, swagger
- `config` — загрузка конфигурации через `cleanenv` из переменных окружения

Данные хранятся в PostgreSQL в следующих таблицах:

- `teams(team_name)` — команды
- `users(user_id, username, team_name, is_active)` — пользователи и их активность
- `pull_requests(pull_request_id, pull_request_name, author_id, status, created_at, merged_at)` — PR и их статусы
- `pr_reviewers(pull_request_id, reviewer_id)` — связи PR–ревьюверы.[file:94]

## Запуск

### Требования

- Go (версия указана в `go.mod`);
- Docker и docker-compose;
- PostgreSQL (поднимается через docker-compose или локально).[file:94]

### Конфигурация

Основные переменные окружения:

```
SERVER_PORT=8080

DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=pr_reviewer
```

`config` загружает эти значения через `cleanenv` и использует для подключения к БД и настройки HTTP-сервера.[file:94]

### Запуск через docker-compose

В корне проекта:

```
docker-compose up --build
```

docker-compose поднимает:

- контейнер с PostgreSQL;
- контейнер с приложением, который:
  - применяет миграции из папки `migrations`;
  - стартует HTTP-сервис на порту 8080

После запуска сервис доступен по адресу:

```
http://localhost:8080
```

### Локальный запуск без Docker

1. Запустить PostgreSQL локально и создать БД `pr_reviewer`
2. Указать параметры подключения в `.env`.
3. Выполнить:

```
make build
./bin/server
```

или:

```
go run ./cmd/app
```

При старте будут автоматически применены миграции.

## Makefile

В корне проекта есть `Makefile` со стандартными таргетами (названия можно подстроить под фактический файл):

```
build:
	go build -o bin/server ./cmd/app

run:
	go run ./cmd/app

lint:
	golangci-lint run ./...

clean:
	go clean
	rm -rf bin/
```

`make build` собирает бинарник в `./bin/server`, `make run` запускает приложение, `make lint` запускает статический анализ

## Линтер

Для статического анализа используется `golangci-lint` v2 (конфигурация в `.golangci.yml`)

Пример минимальной конфигурации:

```
version: "2"

run:
  timeout: 5m

linters:
  enable:
    - govet
    - staticcheck
    - ineffassign
    - unused
    - misspell

formatters:
  enable:
    - gofmt
    - goimports
```

Запуск линтера:

```
make lint
# или
golangci-lint run ./...
```

## Эндпоинт статистики

`GET /stats` возвращает:

```
{
  "total_pr": 42,
  "open_pr": 10,
  "merged_pr": 32,
  "reviewers": [
    { "user_id": "u1", "username": "Alice", "assignments": 15 },
    { "user_id": "u2", "username": "Bob", "assignments": 7 }
  ]
}
```

- `total_pr` — общее количество PR;
- `open_pr` — количество PR в статусе `OPEN`;
- `merged_pr` — количество PR в статусе `MERGED`;
- `reviewers` — список ревьюверов с количеством назначений

Схемы `Stats` и `ReviewerStat` описаны в `openapi.yml`.

## Принятые решения и допущения

- При создании PR ревьюверы выбираются из активных участников **команды автора**, максимум 2, автор не может быть ревьювером своего PR
- При `/pullRequest/reassign` ревьювер заменяется на случайного активного участника его команды; если кандидатов нет — возвращается доменная ошибка
- После `merge` изменение списка ревьюверов запрещено — соответствующие запросы возвращают ошибку `PR_MERGED`
- Идентификаторы (`user_id`, `team_name`, `pull_request_id`) хранятся как строки, без surrogate key — этого достаточно для ограниченного объёма данных в рамках задания
- Миграции применяются при старте сервиса из папки `migrations` в корне проекта.
