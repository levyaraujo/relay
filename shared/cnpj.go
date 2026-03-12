package shared

import "strconv"

const CNPJSize = 14

// ValidateCNPJ checks whether a CNPJ string is structurally valid
// using the official Brazilian two-digit check algorithm.
func ValidateCNPJ(cnpj string) bool {
	digits := stripNonDigits(cnpj)
	if len(digits) != CNPJSize {
		return false
	}
	if allSameDigit(digits) {
		return false
	}

	w1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	if checkDigit(digits[:12], w1) != int(digits[12]-'0') {
		return false
	}

	w2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	if checkDigit(digits[:13], w2) != int(digits[13]-'0') {
		return false
	}

	return true
}

func stripNonDigits(s string) string {
	buf := make([]byte, 0, CNPJSize)
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			buf = append(buf, s[i])
		}
	}
	return string(buf)
}

func allSameDigit(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

func checkDigit(digits string, weights []int) int {
	sum := 0
	for i, w := range weights {
		d, _ := strconv.Atoi(string(digits[i]))
		sum += d * w
	}
	rem := sum % 11
	if rem < 2 {
		return 0
	}
	return 11 - rem
}