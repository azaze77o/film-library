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

Текущая версия схемы — `5`: `users` (с seed-строкой пользователя по умолчанию), `movies`, `library_entries`, индекс `(user_id, created_at DESC)` под список библиотеки, `library_audit_log` — журнал событий библиотеки (только добавление строк).

Откат на один шаг — `migrate ... down 1`. Уже применённые файлы не правят: изменение схемы — новая пара миграций.

## Запуск

Из корня модуля (чтобы подхватился `.env`):

```bash
go run ./cmd/api
```

Сервер слушает `:8080`. Логи запросов — JSON в stdout (`method`, `path`, `status`, `duration`). На старте проверяется подключение к базе (`db connected`) и запускается фоновый воркер аудита (`audit worker started`); если база недоступна, процесс не поднимается.

Остановка: `Ctrl+C` в том же терминале или `SIGTERM`. Порядок такой:

1. HTTP-сервер перестаёт принимать новые соединения и ждёт уже идущие запросы.
2. Закрывается очередь событий, воркер дочитывает то, что в ней осталось.
3. Закрывается пул базы.

Все три шага укладываются в общий бюджет 20 секунд. Если воркер не успел, его отменяют, и события, оставшиеся в очереди, теряются (в логе `worker did not finish in time`).

В логе при обычной остановке: `shutdown signal received`, `http server stopped`, `closing event queue`, `audit worker stopped`, `event queue drained`, `app stopped`. Повторный `Ctrl+C` во время ожидания завершает процесс сразу. `Ctrl+Z` процесс не останавливает: он останется и займёт порт.

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

## Журнал событий (аудит)

После успешного `POST /api/v1/library` сервис публикует событие `library_entry.created` в очередь в памяти (`internal/queue`, буфер на 100 событий). Фоновый воркер (`internal/worker`) читает очередь и пишет строки в таблицу `library_audit_log`:

| Колонка       | Смысл                                                  |
| ------------- | ------------------------------------------------------ |
| `id`          | суррогатный ключ                                       |
| `event_type`  | тип события, например `library_entry.created`          |
| `user_id`     | владелец записи                                        |
| `imdb_id`     | фильм                                                  |
| `occurred_at` | когда событие произошло (время создания записи)        |
| `created_at`  | когда строка попала в журнал (задержка очереди видна из разницы) |

Проверить журнал:

```bash
source .env
psql "$DATABASE_URL" -c "SELECT id, event_type, imdb_id, occurred_at, created_at FROM library_audit_log ORDER BY id DESC LIMIT 5;"
```

Ограничения, которые заданы сознательно:

- Очередь живёт в памяти. При аварийном завершении процесса (`kill -9`, паника) события из буфера теряются. При штатной остановке буфер дочитывается.
- Если очередь переполнена или уже закрыта, клиент всё равно получает `201`, а в лог пишется `Warn` (`publish library event`). Запись в библиотеке уже создана, откатывать её из-за журнала нельзя.
- Ошибка записи события в журнал не останавливает воркер: она пишется в лог (`insert library event`), воркер берёт следующее событие.
- Повторный `POST` с тем же `imdb_id` (`409`) события не создаёт.

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
internal/service             — сценарии библиотеки и поиска, интерфейсы хранилища и издателя событий
internal/queue               — очередь событий на буферизованном канале (Publish, Close, Events)
internal/worker              — фоновый воркер аудита: читает очередь, пишет в журнал
internal/repository/postgres — библиотека и журнал аудита в PostgreSQL (squirrel + database/sql)
internal/repository/omdb     — исходящий клиент OMDb
internal/transport/httpapi   — хендлеры, роутер, ошибки, middleware
```

Схема БД: `users` (id, seed `default`), `movies` (каталог по `imdb_id`), `library_entries` (связь `user_id` + `imdb_id` с флагами). Создание записи идёт одной транзакцией: upsert фильма в каталог, затем вставка в библиотеку.

## Замечания

- Ответ поиска сериализуется полями домена (`Movies`, `TotalResults`, …).
- `go build ./...` должен собираться: единственная точка входа — `cmd/api`.
- Остановка по `Ctrl+C` / `SIGTERM` укладывается в 20 секунд: сначала дожидаются текущие запросы, затем воркер аудита дочитывает очередь, затем закрывается пул базы.
- Тесты конкурентного кода запускайте с детектором гонок: `go test -race ./...`.

