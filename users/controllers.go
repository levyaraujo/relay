package users

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPassword     = errors.New("error occurred during password hashing")
	ErrUserCreation = errors.New("error occurred during user creation")
	ErrUserNotFound = errors.New("user not found")
)

type Controller struct {
	repo UserRepository
}

func NewController(repo UserRepository) *Controller {
	return &Controller{repo: repo}
}

func (c *Controller) Create(u *User) (*User, error) {
	password, err := hashPassword(u.Password)

	if err != nil {
		return nil, ErrPassword
	}

	u.Password = password

	user, errUser := c.repo.Create(u)

	if errUser != nil {
		return nil, ErrUserCreation
	}

	return user, nil
}

func (c *Controller) GetUser(id uuid.UUID) (*User, error) {
	u, err := c.repo.GetByID(id)

	if err != nil {
		return nil, ErrUserNotFound
	}

	return u, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
