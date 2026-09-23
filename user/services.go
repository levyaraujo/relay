package user

import (
	"errors"
	"relay/password"
)

type Service struct {
	repo Repo
}

var UserRegistrationErr = errors.New("An error occurred trying to save the user.")

func (s Service) Register(u *UserPayload) (string, error) {
	var email string

	u.Password = password.HashPassword(u.Password)
	row := s.repo.Create(*u)
	if err := row.Scan(&email); err != nil {
		return "", UserRegistrationErr
	}

	return email, nil
}

func NewService(r Repo) *Service {
	return &Service{repo: r}
}
