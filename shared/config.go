package shared

import (
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Settings struct {
	DatabaseURL string
}

var (
	settings     *Settings
	settingsOnce sync.Once
)

func LoadConfig() *Settings {
	settingsOnce.Do(func() {
		err := godotenv.Load()
		if err != nil {
			godotenv.Load("../.env")
		}

		settings = &Settings{
			DatabaseURL: os.Getenv("DATABASE_URL"),
		}
	})

	return settings
}
