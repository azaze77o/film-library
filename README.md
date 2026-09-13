# Film library

REST API на Go: поиск фильмов через OMDb и личная библиотека (избранное / просмотрено). Хранилище библиотеки — in-memory: после перезапуска сервера коллекция пустая.

Пользователя в API пока нет: все операции с библиотекой идут от фиксированного внутреннего id.

## Требования

- Go 1.22+ (в модуле указан 1.23.4)
- ключ [OMDb API](https://www.omdbapi.com/apikey.aspx) в переменной `OMDB_API_KEY`

Создайте в корне репозитория файл `.env` (он в `.gitignore`):

```
OMDB_API_KEY=ваш_ключ
```

Либо экспортируйте переменную в окружении. Без ключа процесс API не стартует.

## Запуск

Из корня модуля (чтобы подхватился `.env`):

```bash
go run ./cmd/api
```

Сервер слушает `:8080`. Логи запросов — JSON в stdout (`method`, `path`, `status`, `duration`).

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

# библиотека пустая
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
Тело `PATCH`: опционально `favourite` и/или `watched` (boolean). Повторный `POST` с тем же `imdb_id` → `409`. Нет записи → `404`.

## Postman

1. Создайте коллекцию, base URL `http://localhost:8080`.
2. Те же метод + путь, что в таблице.
3. Для `POST`/`PATCH`: Body → raw → JSON.
4. Для поиска: Params `q`, `page`.



## Структура

```
cmd/api                 — точка входа HTTP
internal/domain         — сущности и доменные ошибки
internal/service        — сценарии библиотеки и поиска
internal/repository/memory — in-memory библиотека
internal/repository/omdb   — исходящий клиент OMDb
transport/httpapi       — хендлеры, роутер, ошибки, middleware
```



## Замечания

- Ответ поиска сериализуется полями домена (`Movies`, `TotalResults`, …).
- `go build ./...` должен собираться: единственная точка входа — `cmd/api`.

