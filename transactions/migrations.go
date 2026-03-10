package transactions

import (
	"lucrerp/shared"
)

func Migrate() {
	shared.DB.AutoMigrate(&Transaction{})
}
