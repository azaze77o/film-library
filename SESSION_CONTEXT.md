# Контекст для новой сессии агента

Читать **в начале** новой сессии. Пользователь — Eugene, учит Go, пишет код сам.

Старый файл `review/SESSION_CONTEXT.md` про **закрытую** итерацию REST; актуальная итерация — SQL/PostgreSQL, этот файл.

---

## 0. Как работать (обязательно)

Учебный проект. **Не писать готовые файлы и не править код приложения молча.**

1. Объяснить что / почему / какие библиотеки, с деталями.
2. Небольшой каркас с TODO, не решение целиком.
3. Пользователь пишет → разбор построчно (принцип, не «поправьте строку X»).
4. По команде дописывать конспекты в `review/` **общими** формулировками, без имён этого проекта.

Пользователь явно просил **больше подробностей плана реализации** (куда класть файлы, где маппинг типов), не абстрактные «сделай Get».

Язык общения — **русский**. Не оставлять куски объяснений на английском.

Задавать вопросы по ходу — ему так эффективнее учиться.

**Секреты:** `.env` в `.gitignore` и `.cursorignore`. Не открывать `.env`. Не просить вставить DSN/пароль в чат. `#` в пароле в URL → `%23`. `godotenv.Load()` не перетирает уже заданные env.

`.cursorignore` коммитят (это не секрет). `.env` — нет.

---

## 1. Проект

Путь: `/home/azazello/film_library`  
Модуль: `github.com/project/omdbapp`  
REST: поиск OMDb + личная библиотека. Пользователя в API нет: `defaultUserID = "default"` в хендлере.

Стек БД (зафиксирован): `database/sql` + `pgx/v5/stdlib` + **squirrel везде** для DML. Миграции — чистый SQL, golang-migrate CLI (`~/go/bin/migrate`).

Postgres на удалённой ВМ, отдельная database `film_library`, пользователь приложения — владелец. Миграции **накатаны, версия 3**.

---

## 2. Слои (не ломать)

```
cmd/api → transport/httpapi → service → repository → domain
```

Интерфейс `LibraryStore` объявлен в `internal/service/library.go`.  
`ctx context.Context` — первый аргумент сервиса и стора. Источник: `r.Context()`. В `main` для Ping: `Background` + timeout 5s.

---

## 3. Что сделано в итерации SQL

| Шаг | Статус |
|---|---|
| Схема: `users`, `movies`, `library_entries` | готово, 3 пары миграций |
| `ctx` через слои | готово |
| Пул `postgres.New`, Ping, лимиты | готово |
| `Get` squirrel JOIN Scan NullString | готово, проверено 200/404 |
| `Create` транзакция + upsert movies + 23505 | готово в коде |
| `List` QueryContext Next/Close/Err ORDER BY | готово в коде |
| `Update` | **заглушка** — следующий шаг |
| `Delete` | **заглушка** |
| Хелпер ошибок БД целиком / индексы / README | не делали |

Конспекты: `review/ARCHITECTURE_NOTES_REST_API.md` (REST), `review/ARCHITECTURE_NOTES_SQL_DB.md` главы **1–18** (миграции, ctx, DSN, пул, squirrel, Scan/NULL, граница репозитория). Транзакции/`Create`/`List` в конспект **ещё не дописывали** — по команде пользователя.

---

## 4. Схема БД (кратко)

- `users.id TEXT PK`, seed `'default'`.
- `movies.imdb_id TEXT PK`, `title NOT NULL`, `year`/`kind`/`poster` nullable, `created_at TIMESTAMPTZ DEFAULT now()`.
- `library_entries` PK `(user_id, imdb_id)`, FK users **CASCADE**, FK movies **без CASCADE**, флаги `is_favourite`/`is_watched`, `created_at`.
- **`updated_at` нет** ни в домене, ни в таблицах. Добавлять только новой миграцией 000004, не править 001–003 (уже накатаны).
- Колонка `kind` = домен `Type`. JSON PATCH: `favourite` → домен `IsFavorite`.

Не править успешно применённые миграции. `dirty` ≠ «уже были данные».

---

## 5. Код postgres (сейчас)

```
internal/repository/postgres/db.go       — Open pgx, пул, Ping, Close при ошибке Ping
internal/repository/postgres/row.go      — libraryRow + toDomain (NullString → string)
internal/repository/postgres/library.go  — Get, Create, List; Update/Delete заглушки
```

- Билдер: `var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)` — цепочки **только от `psql`**, не от `sq.Select`.
- `libraryColumns` — порядок = Scan = поля `libraryRow`.
- `Create`: `s.db.BeginTx` + `defer tx.Rollback()` + `tx.ExecContext` (не `db.Exec`) + `Commit`. Movies: `ON CONFLICT (imdb_id) DO UPDATE SET ... EXCLUDED`. Library: INSERT, `isUniqueViolation` → `domain.ErrAlreadyExists`. Сервис **сам** возвращает `entry` клиенту, стор Create только `error`.
- `isUniqueViolation`: `errors.As` + `pgconn.PgError` код `23505`. Обёртки только `%w`.
- `List`: нет `ErrNoRows`; пусто → `[]`. После цикла `rows.Err()`. `make([]T, 0)` чтобы JSON был `[]` не `null`.
- `main`: `postgres.NewLibraryStore(db)`, лог `db connected` **без DSN**. Memory в main не используется, пакет `internal/repository/memory` оставлен (фейк/тесты).

Интерфейс `Update` всё ещё **мутатор** `func(*domain.LibraryEntry)` — артефакт in-memory. В SQL так нельзя. На шаге Update: патч-структура в **`domain`** (не в service — репозиторий не импортирует service). Поправить и memory под тот же интерфейс.

`service.UpdatePath` — опечатка имени (Patch), поля `Favorite`/`Watched *bool`.

---

## 6. Следующий шаг новой сессии: Update

1. Вынести патч в `domain` (или общее имя вроде `LibraryPatch` с `*bool`).
2. Сменить `LibraryStore.Update` — без мутатора; динамический `SET` squirrel `SetMap` / условные `Set`, `Where` user+imdb, `RETURNING` или повторный Get.
3. Сервис: собрать патч, не замыкание.
4. Memory: та же сигнатура.
5. Пустой патч → `domain.ErrInvalidInput` (маппер уже даёт 400).
6. Опционально миграция `updated_at`.

Потом: `Delete` (`Exec` + `RowsAffected==0` → NotFound). В хендлере Delete `return` после ошибки **уже есть**. Затем хелпер `23503` и прочие коды; индексы (`user_id, created_at DESC`); README (данные переживают рестарт); graceful shutdown по желанию.

Проверка API: `go run ./cmd/api` из корня модуля. Порт 8080. curl из README. Поиск OMDb без изменений (`q`, не `s`).

Пустой `psql "$DATABASE_URL"` без `source .env` в **этом** терминале → локальный сокет, не ВМ.

---

## 7. Решения, которые уже приняли

- Squirrel на все DML, не смешивать с сырым SQL в Go.
- Три таблицы, не одна плоская (ради транзакции Create).
- `users.id` TEXT + seed `default`, не BIGSERIAL.
- Кавычки в `.env` вокруг URL; спецсимволы пароля в URI кодировать.
- План реализации в чате, не генерировать MD-план заранее. Конспект — только по команде.

---

## 8. Известные шероховатости (не чинить без просьбы)

- Опечатки в коде поправлены (`already`, `starting`, тексты ошибок сборки запросов).
- `go.mod` go 1.25.x (раньше был 1.23.4).
- Валидация пустого title в сервисе отложена с REST-итерации.
- Ошибки «OMDb ничего не нашёл» по-прежнему не доменные → 500.
- `review/` в `.gitignore` — конспекты могут не быть в git.

---

## 9. Файлы заметок

- `SESSION_CONTEXT.md` (этот, **корень**) — передать агенту.
- `review/ARCHITECTURE_NOTES_SQL_DB.md` — шпаргалка SQL/БД.
- `review/ARCHITECTURE_NOTES_REST_API.md` — REST.
- `README.md` — ещё описывает in-memory; обновить в конце итерации.
