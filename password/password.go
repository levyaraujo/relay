package password

import (
	"log/slog"
	"os"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func HashPassword(pass string) string {
	p := &argon2id.Params{
		Memory:      16 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}

	hash, err := argon2id.CreateHash(pass, p)

	if err != nil {
		slog.Error(err.Error())
	}

	return hash
}

func Verify(storedHash, pass string) bool {
	match, err := argon2id.ComparePasswordAndHash(pass, storedHash)

	if err != nil {
		return false
	}

	return match
}

func JWT(id string) (string, error) {
	godotenv.Load("../.env")
	SECRET := os.Getenv("SECRET_KEY")

	now := time.Now()

	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": "relay-server",
			"sub": id,
			"iat": now.Unix(),
			"exp": now.Add(time.Hour * 24).Unix(),
		},
	)
	s, err := t.SignedString([]byte(SECRET))

	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	return s, nil
}
