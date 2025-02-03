package auth

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/storage/sqlite"
	"sso/utils/jwt"
	"sso/utils/storage"
	"time"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    *sqlite.Storage
	usrProvider *sqlite.Storage
	appProvider *sqlite.Storage
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int64) (models.App, error)
}

func New(log *slog.Logger, userSaver *sqlite.Storage, userProvider *sqlite.Storage, appProvider *sqlite.Storage, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:         log,
		usrSaver:    userSaver,
		usrProvider: userProvider,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int64,
) (string, error) {
	const fn = "internal.services.auth.Login"

	log := a.log.With(slog.String("fn", fn))

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("User not found")

			return "", fmt.Errorf("%s: %w", fn, err)
		}

		a.log.Error("User login failed", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", fn, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("Invalid password", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", fn, err)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		a.log.Error("App login failed", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", fn, err)
	}

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("JWT login failed", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", fn, err)
	}

	log.Info("User ", email, " successfully logged")

	return token, nil

}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error) {
	const fn = "internal.services.auth.RegisterNewUser"

	log := a.log.With(slog.String("fn", fn))

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to hash password", slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExist) {
			a.log.Warn("User already exists", slog.String("error", err.Error()))

			return 0, fmt.Errorf("%s: %w", fn, "User already exists")
		}

		log.Error("Failed to save user", slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	log.Info("Registering new user:", email)

	return id, nil
}

func (a *Auth) IsAdmin(
	ctx context.Context,
	userID int64,
) (bool, error) {
	const fn = "internal.services.auth.IsAdmin"

	log := a.log.With(slog.String("fn", fn))

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			a.log.Warn("App not found")

			return false, fmt.Errorf("%s: %w", fn, "Invalid app id")
		}

		return false, fmt.Errorf("%s: %w", fn, err)
	}

	log.Info("Checked if user ", userID, " is admin", slog.Bool("isAdmin", isAdmin))

	return isAdmin, nil
}
