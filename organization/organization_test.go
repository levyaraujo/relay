package organization

import "testing"

func TestOrganization(t *testing.T) {
	t.Run("cnpj validation", func(t *testing.T) {
		tests := []struct {
			cnpj string
			want bool
		}{
			{"VVJ6HCZA000112", true},
			{"57082641000150", true},
			{"WWY8WVB3000171", false},
		}

		for _, tt := range tests {
			got := ValidateCNPJ(tt.cnpj)

			if got != tt.want {
				t.Errorf("got %t, want %t", got, tt.want)
			}
		}
	})
}
