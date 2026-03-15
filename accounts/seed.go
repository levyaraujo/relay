package accounts

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/levyaraujo/relay/shared"
)

type seedEntry struct {
	Code       string
	Name       string
	Type       AccountType
	ParentCode string
}

// defaultCOA defines a simplified Plano de Contas Referencial
// based on the Brazilian Lucro Presumido model.
var defaultCOA = []seedEntry{
	// 1 - Ativo (Assets)
	{"1", "Ativo", AccountAsset, ""},
	{"1.1", "Ativo Circulante", AccountAsset, "1"},
	{"1.1.1", "Caixa", AccountAsset, "1.1"},
	{"1.1.2", "Bancos", AccountAsset, "1.1"},
	{"1.1.3", "Contas a Receber", AccountAsset, "1.1"},
	{"1.1.4", "Estoques", AccountAsset, "1.1"},
	{"1.1.5", "Impostos a Recuperar", AccountAsset, "1.1"},
	{"1.2", "Ativo Não Circulante", AccountAsset, "1"},
	{"1.2.1", "Imobilizado", AccountAsset, "1.2"},
	{"1.2.2", "Intangível", AccountAsset, "1.2"},
	{"1.2.3", "Investimentos", AccountAsset, "1.2"},

	// 2 - Passivo (Liabilities)
	{"2", "Passivo", AccountLiability, ""},
	{"2.1", "Passivo Circulante", AccountLiability, "2"},
	{"2.1.1", "Contas a Pagar", AccountLiability, "2.1"},
	{"2.1.1.1", "Fornecedores", AccountLiability, "2.1.1"},
	{"2.1.2", "Impostos a Pagar", AccountLiability, "2.1"},
	{"2.1.3", "Salários a Pagar", AccountLiability, "2.1"},
	{"2.1.4", "Encargos Sociais a Pagar", AccountLiability, "2.1"},
	{"2.1.5", "Empréstimos CP", AccountLiability, "2.1"},
	{"2.2", "Passivo Não Circulante", AccountLiability, "2"},
	{"2.2.1", "Empréstimos LP", AccountLiability, "2.2"},
	{"2.2.2", "Provisões", AccountLiability, "2.2"},

	// 3 - Patrimônio Líquido (Equity)
	{"3", "Patrimônio Líquido", AccountEquity, ""},
	{"3.1", "Capital Social", AccountEquity, "3"},
	{"3.2", "Reservas de Capital", AccountEquity, "3"},
	{"3.3", "Reservas de Lucros", AccountEquity, "3"},
	{"3.4", "Lucros ou Prejuízos Acumulados", AccountEquity, "3"},

	// 4 - Receitas (Revenue)
	{"4", "Receitas", AccountRevenue, ""},
	{"4.1", "Receita Operacional", AccountRevenue, "4"},
	{"4.1.1", "Receita de Vendas", AccountRevenue, "4.1"},
	{"4.1.2", "Receita de Serviços", AccountRevenue, "4.1"},
	{"4.1.3", "Deduções da Receita", AccountRevenue, "4.1"},
	{"4.2", "Receitas Não Operacionais", AccountRevenue, "4"},
	{"4.2.1", "Receitas Financeiras", AccountRevenue, "4.2"},
	{"4.2.2", "Outras Receitas", AccountRevenue, "4.2"},

	// 5 - Despesas (Expenses)
	{"5", "Despesas", AccountExpense, ""},
	{"5.1", "Custos", AccountExpense, "5"},
	{"5.1.1", "Custo das Mercadorias Vendidas", AccountExpense, "5.1"},
	{"5.1.2", "Custo dos Serviços Prestados", AccountExpense, "5.1"},
	{"5.2", "Despesas Operacionais", AccountExpense, "5"},
	{"5.2.1", "Folha de Pagamento", AccountExpense, "5.2"},
	{"5.2.2", "Encargos Sociais", AccountExpense, "5.2"},
	{"5.2.3", "Aluguel", AccountExpense, "5.2"},
	{"5.2.4", "Energia e Água", AccountExpense, "5.2"},
	{"5.2.5", "Telecomunicações", AccountExpense, "5.2"},
	{"5.2.6", "Material de Escritório", AccountExpense, "5.2"},
	{"5.2.7", "Depreciação", AccountExpense, "5.2"},
	{"5.3", "Despesas Financeiras", AccountExpense, "5"},
	{"5.3.1", "Juros", AccountExpense, "5.3"},
	{"5.3.2", "Tarifas Bancárias", AccountExpense, "5.3"},
	{"5.3.3", "Multas e Mora", AccountExpense, "5.3"},
}

// SeedDefaultCOA creates the default chart of accounts for a company.
// Call this after creating a new company.
func SeedDefaultCOA(repo AccountRepository, companyId uuid.UUID) error {
	codeToId := make(map[string]uuid.UUID, len(defaultCOA))

	for _, entry := range defaultCOA {
		a := &Account{
			Model:     shared.NewModel(),
			CompanyId: companyId,
			Code:      entry.Code,
			Name:      entry.Name,
			Type:      entry.Type,
		}

		if entry.ParentCode != "" {
			parentId, ok := codeToId[entry.ParentCode]
			if !ok {
				return fmt.Errorf("seed coa: parent %q not found for %q", entry.ParentCode, entry.Code)
			}
			a.ParentId = &parentId
		}

		created, err := repo.Create(a)
		if err != nil {
			return fmt.Errorf("seed coa %q: %w", entry.Code, err)
		}
		codeToId[entry.Code] = created.Id
	}

	return nil
}
