package user

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"

	"golang.org/x/crypto/argon2"
)

type Service struct {
	repo Repo
}

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var UserRegistrationErr = errors.New("an error occurred trying to save the user")

func (s Service) Register(u *UserPayload) (string, error) {
	var email string

	u.Password = hex.EncodeToString(hashPassword(u.Password))
	row := s.repo.Create(*u)
	if err := row.Scan(&email); err != nil {
		return "", UserRegistrationErr
	}

	return email, nil
}

func NewService(r Repo) *Service {
	return &Service{repo: r}
}

func hashPassword(pass string) []byte {
	p := &params{
		memory:      16 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}

	salt, err := generateRandomBytes(p.saltLength)

	if err != nil {
		log.Fatal(err)
	}

	hash := argon2.IDKey([]byte(pass), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	return hash
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
