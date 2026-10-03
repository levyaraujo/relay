package organization

import (
	"errors"
	"log/slog"
	"regexp"
	"relay/password"
	"strings"
	"unicode"
)

type repository interface {
	CreateWithOwner(CreateOrganizationPayload) error
}

type Service struct {
	repo repository
}

var InvalidTaxIDErr = errors.New("The CNPJ or CPF provided is invalid.")
var InvalidOrgTypeErr = errors.New("The organization type is invalid. It must be one of the allowed values: SERVICES or PRODUCTS.")
var OrgRegistrationErr = errors.New("An error occurred trying to save the organization.")

func (s Service) Create(payload CreateOrganizationPayload) error {
	if !ValidateCNPJ(payload.Organization.TaxID) && !ValidateCPF(payload.Organization.TaxID) {
		return InvalidTaxIDErr
	}

	err := ValidateOrgType(payload.Organization.Type)
	if err != nil {
		return err
	}

	payload.User.Password = password.HashPassword(payload.User.Password)
	payload.User.Document = RemoveSpecialChars(payload.User.Document)
	payload.Organization.TaxID = RemoveSpecialChars(payload.Organization.TaxID)
	if err := s.repo.CreateWithOwner(payload); err != nil {
		slog.Error("organization.Create", "err", err.Error())
		return OrgRegistrationErr
	}

	return nil
}

func ValidateOrgType(t OrgType) error {
	if t != "PRODUCTS" && t != "SERVICES" {
		return InvalidOrgTypeErr
	}
	return nil
}

func ValidateCNPJ(cnpj string) bool {
	// 1. Limpa e normaliza a string (remove pontos, barras, traços e joga para maiúsculo)
	var sb strings.Builder
	for _, r := range cnpj {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(unicode.ToUpper(r))
		}
	}
	cleanCNPJ := sb.String()

	// O CNPJ precisa ter exatamente 14 caracteres
	if len(cleanCNPJ) != 14 {
		return false
	}

	// Os 12 primeiros caracteres podem ser letras ou números, mas os 2 últimos DVs DEVEM ser números
	for i := 12; i < 14; i++ {
		if !unicode.IsDigit(rune(cleanCNPJ[i])) {
			return false
		}
	}

	// Pesos oficiais do Módulo 11 para o CNPJ
	pesosDV1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	pesosDV2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	// 2. Cálculo do Primeiro Dígito Verificador (DV1)
	soma1 := 0
	for i := 0; i < 12; i++ {
		// Regra oficial: Valor ASCII do caractere menos 48
		valor := int(cleanCNPJ[i]) - 48
		soma1 += valor * pesosDV1[i]
	}

	resto1 := soma1 % 11
	dv1 := 0
	if resto1 >= 2 {
		dv1 = 11 - resto1
	}

	// 3. Cálculo do Segundo Dígito Verificador (DV2)
	soma2 := 0
	for i := 0; i < 12; i++ {
		valor := int(cleanCNPJ[i]) - 48
		soma2 += valor * pesosDV2[i]
	}
	// Inclui o primeiro DV calculado na soma do segundo DV
	soma2 += dv1 * pesosDV2[12]

	resto2 := soma2 % 11
	dv2 := 0
	if resto2 >= 2 {
		dv2 = 11 - resto2
	}

	// 4. Verificação final dos dígitos informados no CNPJ
	return int(cleanCNPJ[12]-'0') == dv1 && int(cleanCNPJ[13]-'0') == dv2
}

func ValidateCPF(cpf string) bool {
	var sb strings.Builder
	for _, r := range cpf {
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
			continue
		}

		if unicode.IsLetter(r) {
			return false
		}
	}

	cleanCPF := sb.String()
	if len(cleanCPF) != 11 {
		return false
	}

	allEqual := true
	for i := 1; i < len(cleanCPF); i++ {
		if cleanCPF[i] != cleanCPF[0] {
			allEqual = false
			break
		}
	}
	if allEqual {
		return false
	}

	soma1 := 0
	for i := 0; i < 9; i++ {
		soma1 += int(cleanCPF[i]-'0') * (10 - i)
	}

	resto1 := soma1 % 11
	dv1 := 0
	if resto1 >= 2 {
		dv1 = 11 - resto1
	}

	soma2 := 0
	for i := 0; i < 10; i++ {
		soma2 += int(cleanCPF[i]-'0') * (11 - i)
	}

	resto2 := soma2 % 11
	dv2 := 0
	if resto2 >= 2 {
		dv2 = 11 - resto2
	}

	return int(cleanCPF[9]-'0') == dv1 && int(cleanCPF[10]-'0') == dv2
}

func RemoveSpecialChars(s string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9 ]+")
	cleanStr := reg.ReplaceAllString(s, " ")
	cleanStr = strings.ReplaceAll(cleanStr, " ", "")
	return cleanStr
}

func NewService(r Repo) *Service {
	return &Service{repo: r}
}
