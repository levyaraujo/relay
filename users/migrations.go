package users

import "lucrerp/shared"

func Migrate() {
	shared.DB.MustExec(`
		CREATE TABLE IF NOT EXISTS users (
			id          UUID PRIMARY KEY,
			name        TEXT NOT NULL,
			email       TEXT NOT NULL,
			password    TEXT NOT NULL,
			company_id  UUID NOT NULL REFERENCES companies(id),
			created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted     BOOLEAN NOT NULL DEFAULT false
		)
	`)
}
