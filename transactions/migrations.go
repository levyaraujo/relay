package transactions

import "lucrerp/shared"

func Migrate() {
	shared.DB.MustExec(`
		CREATE TABLE IF NOT EXISTS transactions (
			id          UUID PRIMARY KEY,
			company_id  UUID NOT NULL REFERENCES companies(id),
			type        INTEGER NOT NULL,
			amount      NUMERIC(15,2) NOT NULL,
			description TEXT NOT NULL,
			due_date    TIMESTAMPTZ,
			paid_date   TIMESTAMPTZ,
			origin      TEXT NOT NULL,
			creator_id  UUID NOT NULL REFERENCES users(id),
			created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted     BOOLEAN NOT NULL DEFAULT false
		)
	`)
}
