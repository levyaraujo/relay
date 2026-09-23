package organization

import (
	"errors"
	"strings"
	"unicode"
)

type Service struct {
	repo Repo
}

var InvalidCNPJErr = errors.New("The CNPJ provided is invalid.")

func (s Service) Create(o Organization) (string, error) {
	if ValidateCNPJ(o.TaxID) == false {
		return "", InvalidCNPJErr
	}

	return o.Email, nil
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
