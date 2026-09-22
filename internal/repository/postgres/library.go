package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/project/omdbapp/internal/domain"
)

// По умолчанию squirrel ставит ? (стиль MySQL).
// Postgres понимает только $1, $2. Поэтому инициализируем psql
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type LibraryStore struct {
	db *sql.DB
}

func NewLibraryStore(db *sql.DB) *LibraryStore {
	return &LibraryStore{db: db}
}

// domainErr переводит ошибку postgres в доменную
func domainErr(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}

	switch pgErr.Code {
	case "23505":
		return domain.ErrAlreadyExists
	case "23503":
		return domain.ErrNotFound
	default:
		return nil
	}
}

var libraryColumns = []string{
	"le.user_id",
	"le.imdb_id",
	"m.title",
	"m.year",
	"m.kind",
	"m.poster",
	"le.is_favourite",
	"le.is_watched",
	"le.created_at",
}

func (s *LibraryStore) Get(ctx context.Context, u domain.UserID, id domain.ImdbID) (domain.LibraryEntry, error) {
	// 1. Собираем запрос
	q := psql.Select(libraryColumns...).
		From("library_entries le").
		Join("movies m ON m.imdb_id = le.imdb_id").
		Where(sq.Eq{
			"le.user_id": string(u),
			"le.imdb_id": string(id),
		})

	// 2. Получить строку sql и аргументы
	query, args, err := q.ToSql()
	if err != nil {
		return domain.LibraryEntry{}, fmt.Errorf("build get query: %w", err)
	}

	// 3. Выполнить и прочитать одну строку
	var row libraryRow
	err = s.db.QueryRowContext(ctx, query, args...).Scan(
		&row.UserID,
		&row.ImdbID,
		&row.Title,
		&row.Year,
		&row.Kind,
		&row.Poster,
		&row.IsFavourite,
		&row.IsWatched,
		&row.CreatedAt,
	)
	// 4.Разбираем ошибки, приводим их к доменным
	if errors.Is(err, sql.ErrNoRows) {
		return domain.LibraryEntry{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.LibraryEntry{}, fmt.Errorf("get library entry: %w", err)
	}

	return row.toDomain(), nil
}

func (s *LibraryStore) Create(ctx context.Context, e domain.LibraryEntry) error {
	// 1. Берем одно соединение из пула для транзакции
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// Делаем откат изменений, возвращаем соединение в пул
	// Без отката, соединение может остаться «в открытой транзакции» и испортить следующий запрос
	defer tx.Rollback()

	// Запрос на запись фильма
	q := psql.Insert("movies").
		Columns("imdb_id", "title", "year", "kind", "poster").
		Values(
			string(e.MoviePreview.ID),
			e.MoviePreview.Title,
			e.MoviePreview.Year,
			e.MoviePreview.Type,
			e.MoviePreview.Poster,
		).
		Suffix(`ON CONFLICT (imdb_id) DO UPDATE SET
		title = EXCLUDED.title,
		year = EXCLUDED.year,
		kind = EXCLUDED.kind,
		poster = EXCLUDED.poster`)

	query, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("upsert movie: %w", err)
	}

	// Запрос на запись фильма в библиотеку
	q = psql.Insert("library_entries").
		Columns("user_id", "imdb_id", "is_favourite", "is_watched", "created_at").
		Values(
			string(e.UserID),
			string(e.MoviePreview.ID),
			e.IsFavorite,
			e.IsWatched,
			e.CreatedAt,
		)
	query, args, err = q.ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		if dom := domainErr(err); dom != nil {
			return dom
		}
		return fmt.Errorf("insert library entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

func (s *LibraryStore) List(ctx context.Context, u domain.UserID) ([]domain.LibraryEntry, error) {
	// 1. Собираем запрос
	q := psql.Select(libraryColumns...).
		From("library_entries le").
		Join("movies m ON m.imdb_id = le.imdb_id").
		Where(sq.Eq{
			"le.user_id": string(u)}).
		OrderBy("le.created_at DESC")

	// 2. Получить строку sql и аргументы
	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list query: %w", err)
	}

	// 3. Выполнить запрос и прочитать результат
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list library: %w", err)
	}
	defer rows.Close()

	result := make([]domain.LibraryEntry, 0)

	for rows.Next() {
		var row libraryRow
		if err := rows.Scan(
			&row.UserID,
			&row.ImdbID,
			&row.Title,
			&row.Year,
			&row.Kind,
			&row.Poster,
			&row.IsFavourite,
			&row.IsWatched,
			&row.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan library row: %w", err)
		}
		result = append(result, row.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate library rows: %w", err)
	}
	return result, nil
}

func (s *LibraryStore) Update(ctx context.Context, u domain.UserID, id domain.ImdbID, patch domain.LibraryPatch) (domain.LibraryEntry, error) {
	// 1.Создаем map для вставки значений.
	sets := map[string]any{}

	// 2.Обновляем признаки просмотрено и избранное
	if patch.Favourite != nil {
		sets["is_favourite"] = *patch.Favourite
	}
	if patch.Watched != nil {
		sets["is_watched"] = *patch.Watched
	}

	// 3.Формируем запрос
	q := psql.Update("library_entries").
		SetMap(sets).
		Where(sq.Eq{
			"user_id": string(u),
			"imdb_id": string(id),
		})

	// 4.Собираем строку запроса в SQL
	query, args, err := q.ToSql()
	if err != nil {
		return domain.LibraryEntry{}, fmt.Errorf("build update query: %w", err)
	}

	// 5.Выполняем запрос
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return domain.LibraryEntry{}, fmt.Errorf("update library entry: %w", err)
	}

	// 6.Обрабатываем результаты запроса
	n, err := res.RowsAffected()
	if err != nil {
		return domain.LibraryEntry{}, fmt.Errorf("update rows affected: %w", err)
	}
	if n == 0 {
		return domain.LibraryEntry{}, domain.ErrNotFound
	}

	return s.Get(ctx, u, id)
}

func (s *LibraryStore) Delete(ctx context.Context, u domain.UserID, id domain.ImdbID) error {
	// 1.Собираем запрос
	q := psql.Delete("library_entries").
		Where(sq.Eq{
			"user_id": string(u),
			"imdb_id": string(id),
		})

	// 2.Получаем строку sql и аргументы
	query, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build delete query: %w", err)
	}

	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete entry: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}

	return nil
}
