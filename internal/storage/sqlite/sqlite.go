package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"sso/internal/domain/models"
	"sso/utils/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const fn = "internal.storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const fn = "internal.storage.sqlite.SaveUser"

	statement, err := s.db.Prepare("INSERT INTO users(email, pass_hash) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	result, err := statement.Exec(email, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrUserExist)
		}

		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const fn = "internal.storage.sqlite.User"

	statement, err := s.db.Prepare("SELECT id, email, pass_hash FROM users WHERE email=?")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", fn, err)
	}

	var UserToReturn models.User
	err = statement.QueryRow(email).Scan(&UserToReturn.ID, &UserToReturn.Email, &UserToReturn.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", fn, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", fn, err)
	}

	return UserToReturn, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const fn = "internal.storage.sqlite.IsAdmin"

	statement, err := s.db.Prepare("SELECT is_admin FROM users WHERE id=?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", fn, err)
	}

	var IsAdmin bool
	err = statement.QueryRow(userID).Scan(&IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", fn, storage.ErrUserNotFound)
		}

		return false, fmt.Errorf("%s: %w", fn, err)
	}

	return IsAdmin, nil
}

func (s *Storage) App(ctx context.Context, appID int64) (models.App, error) {
	const fn = "internal.storage.sqlite.App"

	statement, err := s.db.Prepare("SELECT id, name FROM apps WHERE id=?")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", fn, err)
	}

	var AppToReturn models.App
	err = statement.QueryRow(appID).Scan(&AppToReturn.ID, &AppToReturn.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", fn, storage.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf("%s: %w", fn, err)
	}

	return AppToReturn, nil
}
