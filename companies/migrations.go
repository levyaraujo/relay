package companies

import "lucrerp/shared"

func Migrate() {
	shared.DB.MustExec(`
		CREATE TABLE IF NOT EXISTS companies (
			id          UUID PRIMARY KEY,
			name        TEXT NOT NULL,
			cnpj        TEXT NOT NULL,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted     BOOLEAN NOT NULL DEFAULT false
		)
	`)
}
