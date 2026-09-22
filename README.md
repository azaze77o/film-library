# Film library

REST API на Go: поиск фильмов через OMDb и личная библиотека (избранное / просмотрено). Библиотека хранится в PostgreSQL: данные переживают перезапуск сервера.

Пользователя в API пока нет: все операции с библиотекой идут от фиксированного внутреннего id.

## Требования

- Go 1.25+ (версия модуля — `go 1.25.0`)
- PostgreSQL: отдельная database, пользователь приложения — её владелец
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI для миграций
- ключ [OMDb API](https://www.omdbapi.com/apikey.aspx) в переменной `OMDB_API_KEY`

Создайте в корне репозитория файл `.env` (он в `.gitignore`):

```
OMDB_API_KEY=ваш_ключ
DATABASE_URL="postgres://пользователь:пароль@хост:5432/имя_базы?sslmode=disable"
```

Либо экспортируйте переменные в окружении. Без `OMDB_API_KEY` или `DATABASE_URL` процесс API не стартует. 

## Миграции

Схема — чистый SQL в `migrations/`, пары `*.up.sql` / `*.down.sql`. Применять до первого запуска API, из корня модуля:

```bash
source .env
migrate -path ./migrations -database "$DATABASE_URL" up
migrate -path ./migrations -database "$DATABASE_URL" version
```

Текущая версия схемы — `4`: `users` (с seed-строкой пользователя по умолчанию), `movies`, `library_entries`, индекс `(user_id, created_at DESC)` под список библиотеки.

Откат на один шаг — `migrate ... down 1`. Уже применённые файлы не правят: изменение схемы — новая пара миграций.

## Запуск

Из корня модуля (чтобы подхватился `.env`):

```bash
go run ./cmd/api
```

Сервер слушает `:8080`. Логи запросов — JSON в stdout (`method`, `path`, `status`, `duration`). На старте проверяется подключение к базе (`db connected`); если база недоступна, процесс не поднимается.

Остановка: `Ctrl+C` в том же терминале (не `Ctrl+Z` — процесс останется и займёт порт).

Повторный запуск, если порт занят:

```bash
ss -lptn | grep 8080
kill <pid>
go run ./cmd/api
```



## API

Базовый префикс: `/api/v1`.


| Метод    | Путь                          | Назначение                                         | Успех                        |
| -------- | ----------------------------- | -------------------------------------------------- | ---------------------------- |
| `GET`    | `/api/v1/search?q=...&page=1` | поиск (OMDb), `page` необязателен (по умолчанию 1) | `200`                        |
| `GET`    | `/api/v1/library`             | список записей библиотеки                          | `200` (пустой список — `[]`) |
| `POST`   | `/api/v1/library`             | создать запись                                     | `201`                        |
| `GET`    | `/api/v1/library/{id}`        | одна запись (`id` — imdb id)                       | `200`                        |
| `PATCH`  | `/api/v1/library/{id}`        | частичное обновление флагов                        | `200`                        |
| `DELETE` | `/api/v1/library/{id}`        | удалить запись                                     | `204`                        |


Ошибки в едином JSON:

```json
{ "error": { "code": "NOT_FOUND", "message": "..." } }
```

Частые коды: `400` `INVALID_INPUT`, `404` `NOT_FOUND`, `409` `ALREADY_EXISTS`, `500` `INTERNAL`.

`q` — параметр **этого** API (строка поиска), не имя поля OMDb. Пустой `q` → `400`, внешний сервис не вызывается.

Поиск и библиотека независимы: найти фильм ≠ добавить в коллекцию. Добавление — отдельный `POST`.

## Примеры curl

Кавычки вокруг URL с `&` обязательны в shell.

```bash
# поиск
curl -i "http://localhost:8080/api/v1/search?q=matrix&page=1"

# пустой поиск
curl -i "http://localhost:8080/api/v1/search"

# список библиотеки (пустая — `[]`)
curl -i http://localhost:8080/api/v1/library

# создать
curl -i -X POST http://localhost:8080/api/v1/library \
  -H "Content-Type: application/json" \
  -d '{"imdb_id":"tt0133093","title":"The Matrix","year":"1999","type":"movie","poster":""}'

# одна запись
curl -i http://localhost:8080/api/v1/library/tt0133093

# избранное (только присланные поля)
curl -i -X PATCH http://localhost:8080/api/v1/library/tt0133093 \
  -H "Content-Type: application/json" \
  -d '{"favourite":true}'

# удалить
curl -i -X DELETE http://localhost:8080/api/v1/library/tt0133093
```

Тело `POST`: `imdb_id`, `title`, `year`, `type`, `poster`.  
Тело `PATCH`: опционально `favourite` и/или `watched` (boolean); присланные поля обновляются, остальные остаются как были. Пустое тело `{}` → `400`. Повторный `POST` с тем же `imdb_id` → `409`. Нет записи → `404`.

`DELETE` убирает запись из библиотеки, но не из каталога `movies`: фильм можно добавить снова. Повторный `DELETE` того же id → `404`.

## Postman

1. Создайте коллекцию, base URL `http://localhost:8080`.
2. Те же метод + путь, что в таблице.
3. Для `POST`/`PATCH`: Body → raw → JSON.
4. Для поиска: Params `q`, `page`.



## Структура

```
cmd/api                      — точка входа HTTP, пул БД, сборка зависимостей
migrations                   — SQL-миграции схемы (golang-migrate)
internal/domain              — сущности, патч обновления, доменные ошибки
internal/service             — сценарии библиотеки и поиска, интерфейс хранилища
internal/repository/postgres — библиотека в PostgreSQL (squirrel + database/sql)
internal/repository/memory   — in-memory библиотека (фейк для тестов)
internal/repository/omdb     — исходящий клиент OMDb
transport/httpapi            — хендлеры, роутер, ошибки, middleware
```

Схема БД: `users` (id, seed `default`), `movies` (каталог по `imdb_id`), `library_entries` (связь `user_id` + `imdb_id` с флагами). Создание записи идёт одной транзакцией: upsert фильма в каталог, затем вставка в библиотеку.

## Замечания

- Ответ поиска сериализуется полями домена (`Movies`, `TotalResults`, …).
- `go build ./...` должен собираться: единственная точка входа — `cmd/api`.
- In-memory хранилище в `main` не подключено, пакет оставлен как фейк для тестов.
- Graceful shutdown не реализован: `Ctrl+C` обрывает процесс, пул закрывается при выходе из `main`.

