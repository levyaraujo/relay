package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/levyaraujo/relay/shared"
	"github.com/levyaraujo/relay/users"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrWrongUsernameOrPassword = errors.New("wrong username or password")
)

type Controller struct {
	userRepo users.UserRepository
}

func NewController(userRepo users.UserRepository) *Controller {
	return &Controller{userRepo: userRepo}
}

type Token struct {
	AccessToken string
}

func (c *Controller) Login(email, password string) (*Token, error) {
	user, err := c.userRepo.GetByEmail(email)

	if err != nil {
		return nil, ErrWrongUsernameOrPassword
	}
	correctPassword := checkPasswordHash(password, user.Password)

	if correctPassword == false {
		return nil, ErrWrongUsernameOrPassword
	}

	token, err := createAccessToken(user.Id)

	if err != nil {
		return nil, err
	}

	return &Token{AccessToken: token}, nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func createAccessToken(id uuid.UUID) (string, error) {

	settings := shared.LoadConfig()
	secret := settings.JWT_SECRET
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": id.String(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	s, err := t.SignedString([]byte(secret))

	if err != nil {
		return "", err
	}

	return s, nil
}
