package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"

	"url-shortener/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(StoragePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", StoragePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия базы данных %s, %w", op, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE, 
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания таблицы базы данных %w", err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("ошибка входа в базу данных %s, %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления URL %s:%w", op, err)
	}

	rez, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("alias не уникален:%s, %w", op, storage.ErrURLExist)
		}
		return 0, fmt.Errorf("%s, %w", op, err)
	}

	id, err := rez.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: ошибка добавления последнего ID %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("принят запрос %s %w", op, err)
	}

	var rezURL string
	err = stmt.QueryRow(alias).Scan(&rezURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("выполняется запрос: %s %w", op, err)
	}

	return rezURL, nil

}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s ошибка принятия операции удаления: %w", op, err)
	}

	result, err := stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: ошибка выполнения запроса: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: ошибка при проверке количества удалённых строк: %w", op, err)
	}
	if rowsAffected == 0 {
		return storage.ErrURLNotFound // Возвращаем ошибку, если строка не была удалена
	}

	return nil

}
