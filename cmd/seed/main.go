// Command seed creates a default company, user, vendors, and transactions
// for local development.
package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/levyaraujo/relay/accounts"
	"github.com/levyaraujo/relay/companies"
	"github.com/levyaraujo/relay/shared"
	"github.com/levyaraujo/relay/users"
)

const (
	vendorCount = 50
	txPerMonth  = 20_000
	monthsBack  = 12
	batchSize   = 4000
)

var vendorNames = []string{
	"Fornecedor Alpha", "Distribuidora Beta", "Serviços Gamma", "Logística Delta",
	"Papelaria Epsilon", "Consultoria Zeta", "TI Eta", "Marketing Theta",
	"Transportes Iota", "Materiais Kappa", "Segurança Lambda", "Limpeza Mu",
	"Contabilidade Nu", "Jurídico Xi", "Telecom Omicron", "Energia Pi",
	"Água Rho", "Internet Sigma", "Aluguel Tau", "Seguros Upsilon",
	"Manutenção Phi", "Construção Chi", "Design Psi", "Alimentação Omega",
	"Combustível Alfa2", "Ferramentas Beta2", "Embalagens Gamma2", "Químicos Delta2",
	"Têxtil Epsilon2", "Elétrica Zeta2", "Hidráulica Eta2", "Refrigeração Theta2",
	"Automação Iota2", "Pintura Kappa2", "Vidraçaria Lambda2", "Serralheria Mu2",
	"Gráfica Nu2", "Eventos Xi2", "Treinamento Omicron2", "Auditoria Pi2",
	"Frete Rho2", "Courier Sigma2", "Cloud Tau2", "SaaS Upsilon2",
	"Hospedagem Phi2", "Domínios Chi2", "Publicidade Psi2", "Mídia Omega2",
	"Coworking Alfa3", "Estacionamento Beta3",
}

var descriptions = []string{
	"Pagamento mensal", "Compra de materiais", "Serviço prestado", "Manutenção preventiva",
	"Reembolso", "Fatura recorrente", "Compra avulsa", "Assinatura mensal",
	"Consultoria", "Projeto especial", "Suporte técnico", "Licença de software",
	"Aluguel de equipamento", "Serviço de entrega", "Instalação",
}

func main() {
	db := shared.Connect()
	defer db.Close()

	companyRepo := companies.NewRepository(db)
	companyCtrl := companies.NewController(companyRepo)
	accRepo := accounts.NewRepository(db)
	userRepo := users.NewRepository(db)
	userCtrl := users.NewController(userRepo)

	// Seed company (skip if exists)
	var companyID uuid.UUID
	existingCompany := &companies.Company{}
	err := db.Get(existingCompany, `SELECT * FROM companies WHERE cnpj = $1 AND deleted = false`, "11222333000181")
	if err == nil {
		companyID = existingCompany.Id
		fmt.Printf("company exists: %s (%s)\n", existingCompany.Name, companyID)
	} else {
		co := &companies.Company{
			Name: "Relay ERP",
			CNPJ: "11222333000181",
		}
		created, err := companyCtrl.Create(co, accRepo)
		if err != nil {
			log.Fatalf("seed company: %v", err)
		}
		companyID = created.Id
		fmt.Printf("company created: %s (%s)\n", created.Name, companyID)
	}

	// Seed user (skip if exists)
	var userID uuid.UUID
	existingUser, err := userRepo.GetByEmail("admin@relay.com")
	if err == nil {
		userID = existingUser.Id
		fmt.Printf("user exists: %s <%s> (%s)\n", existingUser.Name, existingUser.Email, userID)
	} else {
		u := &users.User{
			Model:     shared.NewModel(),
			Name:      "Admin Relay",
			Email:     "admin@relay.com",
			Password:  "levyaraujo",
			CompanyId: companyID,
		}
		createdUser, err := userCtrl.Create(u)
		if err != nil {
			log.Fatalf("seed user: %v", err)
		}
		userID = createdUser.Id
		fmt.Printf("user created: %s <%s> (%s)\n", createdUser.Name, createdUser.Email, userID)
	}

	// Seed vendors
	fmt.Println("seeding vendors...")
	vendorIDs := seedVendors(db, companyID)
	fmt.Printf("  %d vendors created\n", vendorCount)

	// Fetch account IDs for transactions
	accountIDs := fetchAccountIDs(db, companyID)
	fmt.Printf("  %d accounts available\n", len(accountIDs))

	// Seed transactions
	fmt.Println("seeding transactions...")
	total := seedTransactions(db, companyID, userID, vendorIDs, accountIDs)
	fmt.Printf("done: %d vendors, %d transactions\n", vendorCount, total)
}

func seedVendors(db *sqlx.DB, companyID uuid.UUID) []uuid.UUID {
	for i := 0; i < vendorCount; i++ {
		now := time.Now()
		id := uuid.New()
		cnpj := fmt.Sprintf("%014d", rand.Int63n(99999999999999))
		_, err := db.Exec(`
			INSERT INTO vendors (id, company_id, name, cnpj, email, phone, payment_terms, created_at, updated_at, deleted)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false)
			ON CONFLICT DO NOTHING`,
			id, companyID, vendorNames[i], cnpj,
			fmt.Sprintf("contato@vendor%d.com", i),
			fmt.Sprintf("11%09d", rand.Int63n(999999999)),
			[]int{15, 30, 45, 60}[rand.Intn(4)],
			now, now,
		)
		if err != nil {
			log.Printf("vendor %d: %v", i, err)
		}
	}

	var ids []uuid.UUID
	err := db.Select(&ids, `SELECT id FROM vendors WHERE company_id = $1 AND deleted = false`, companyID)
	if err != nil {
		log.Fatalf("fetch vendor ids: %v", err)
	}
	return ids
}

func fetchAccountIDs(db *sqlx.DB, companyID uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	err := db.Select(&ids, `
		SELECT id FROM accounts
		WHERE company_id = $1 AND deleted = false
		  AND id NOT IN (SELECT DISTINCT parent_id FROM accounts WHERE parent_id IS NOT NULL AND company_id = $1)`,
		companyID,
	)
	if err != nil {
		log.Fatalf("fetch accounts: %v", err)
	}
	return ids
}

func seedTransactions(db *sqlx.DB, companyID, userID uuid.UUID, vendorIDs, accountIDs []uuid.UUID) int {
	now := time.Now()
	total := 0

	for m := monthsBack - 1; m >= 0; m-- {
		monthStart := time.Date(now.Year(), now.Month()-time.Month(m), 1, 0, 0, 0, 0, time.Local)
		monthEnd := monthStart.AddDate(0, 1, -1)
		daysInMonth := monthEnd.Day()

		fmt.Printf("  %s: %d transactions...\n", monthStart.Format("Jan 2006"), txPerMonth)

		batch := make([]string, 0, batchSize)
		args := make([]any, 0, batchSize*14)
		paramIdx := 1

		for i := 0; i < txPerMonth; i++ {
			id := uuid.New()
			day := rand.Intn(daysInMonth) + 1
			txDate := time.Date(monthStart.Year(), monthStart.Month(), day, rand.Intn(18)+6, rand.Intn(60), 0, 0, time.Local)

			txType := rand.Intn(2) + 1 // 1=debit, 2=credit
			amount := float64(rand.Intn(500000)+100) / 100.0
			desc := descriptions[rand.Intn(len(descriptions))]
			vendorID := vendorIDs[rand.Intn(len(vendorIDs))]
			accountID := accountIDs[rand.Intn(len(accountIDs))]

			var paidDate time.Time
			if rand.Float32() < 0.7 {
				paidDate = txDate.AddDate(0, 0, rand.Intn(30))
			}

			dueDate := txDate.AddDate(0, 0, []int{15, 30, 45, 60}[rand.Intn(4)])

			batch = append(batch, fmt.Sprintf(
				"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
				paramIdx, paramIdx+1, paramIdx+2, paramIdx+3, paramIdx+4,
				paramIdx+5, paramIdx+6, paramIdx+7, paramIdx+8, paramIdx+9,
				paramIdx+10, paramIdx+11, paramIdx+12, paramIdx+13,
			))
			args = append(args, id, companyID, accountID, vendorID, txType, amount, desc, dueDate, paidDate, "seed", userID, txDate, txDate, false)
			paramIdx += 14

			if len(batch) >= batchSize {
				flushBatch(db, batch, args)
				total += len(batch)
				batch = batch[:0]
				args = args[:0]
				paramIdx = 1
			}
		}

		if len(batch) > 0 {
			flushBatch(db, batch, args)
			total += len(batch)
		}
	}

	return total
}

func flushBatch(db *sqlx.DB, batch []string, args []any) {
	query := "INSERT INTO transactions (id, company_id, account_id, vendor_id, type, amount, description, due_date, paid_date, origin, creator_id, created_at, updated_at, deleted) VALUES "
	for i, b := range batch {
		if i > 0 {
			query += ","
		}
		query += b
	}

	_, err := db.Exec(query, args...)
	if err != nil {
		log.Fatalf("batch insert: %v", err)
	}
}
