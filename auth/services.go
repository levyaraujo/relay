package auth

import (
	"errors"
	"log/slog"
	"relay/password"
	"relay/user"
)

type Service struct {
	uRepo user.Repo
}

func NewService(r user.Repo) *Service {
	return &Service{uRepo: r}
}

var IncorrectEmailOrPasswordErr = errors.New("The email or password is incorrect")

func (s Service) Login(l LoginPayload) (string, error) {
	u, err := s.uRepo.FindByEmail(l.Email)

	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	passIsCorrect := password.Verify(u.Password, l.Password)

	if passIsCorrect != true {
		return "", IncorrectEmailOrPasswordErr
	}

	t, err := password.JWT(u.ID.String())

	return t, err
}
