package shared

import (
	"log"
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
			log.Fatal("Error loading .env file")
		}

		settings = &Settings{
			DatabaseURL: os.Getenv("DATABASE_URL"),
		}
	})

	return settings
}
