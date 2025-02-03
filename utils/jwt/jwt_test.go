package jwt

import (
	"golang.org/x/crypto/bcrypt"
	"sso/internal/domain/models"
	"testing"
	"time"
)

const Hour = 30 * time.Minute
const Password = "coolpassword123"

func TestNewToken(t *testing.T) {
	PassHash, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		t.Error(err)
	}

	user := models.User{
		ID:       1,
		Email:    "test@example.com",
		PassHash: PassHash,
	}

	app := models.App{
		ID:   1,
		Name: "Auth",
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(Password)); err != nil {
		t.Error(err)
	}

	actual, err := NewToken(user, app, Hour)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(actual)
}
