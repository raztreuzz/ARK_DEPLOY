package tailscale

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "descripción válida simple",
			input:    "My auth key",
			expected: "My auth key",
		},
		{
			name:     "descripción con caracteres permitidos",
			input:    "Test-key_123.prod:v1",
			expected: "Test-key_123.prod:v1",
		},
		{
			name:     "descripción con caracteres inválidos",
			input:    "Test@key#with$invalid%chars!",
			expected: "Test-key-with-invalid-chars",
		},
		{
			name:     "descripción con espacios múltiples",
			input:    "Test   key   here",
			expected: "Test   key   here",
		},
		{
			name:     "descripción con guiones múltiples",
			input:    "Test---key---here",
			expected: "Test-key-here",
		},
		{
			name:     "descripción con timestamp (formato fecha)",
			input:    "Test auth key - 2026-02-18 15:04:05",
			expected: "Test auth key - 2026-02-18 15-04-05",
		},
		{
			name:     "descripción vacía",
			input:    "",
			expected: "ark-auth-key",
		},
		{
			name:     "solo espacios",
			input:    "    ",
			expected: "ark-auth-key",
		},
		{
			name:     "descripción con guiones al inicio y final",
			input:    "--test-key--",
			expected: "test-key",
		},
		{
			name:     "descripción muy larga",
			input:    "Esta es una descripción extremadamente larga que debería ser truncada porque excede el límite de 100 caracteres establecido para las descripciones de auth keys",
			expected: "Esta es una descripción extremadamente larga que debería ser truncada porque excede el límite de 100",
		},
		{
			name:     "descripción con emojis",
			input:    "Test 🔑 auth key 🚀",
			expected: "Test - auth key",
		},
		{
			name:     "descripción con saltos de línea",
			input:    "Test\nkey\rwith\r\nnewlines",
			expected: "Test-key-with-newlines",
		},
		{
			name:     "solo caracteres inválidos",
			input:    "@#$%^&*()",
			expected: "ark-auth-key",
		},
		{
			name:     "descripción con paréntesis y corchetes",
			input:    "Test [prod] (v2.0)",
			expected: "Test -prod- -v2.0",
		},
		{
			name:     "descripción con URL",
			input:    "Deploy from https://github.com/user/repo",
			expected: "Deploy from https---github.com-user-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeDescription(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeDescriptionLength(t *testing.T) {
	// Crear una descripción de exactamente 150 caracteres
	longDesc := strings.Repeat("a", 150)
	result := SanitizeDescription(longDesc)
	
	assert.LessOrEqual(t, len(result), 100, "La descripción debe estar limitada a 100 caracteres")
}

func TestSanitizeDescriptionPreservesValid(t *testing.T) {
	validInputs := []string{
		"Production deployment key",
		"dev-server-001",
		"backend_service.v1",
		"Test key: staging",
		"Key-123_ABC.prod:v2",
	}

	for _, input := range validInputs {
		result := SanitizeDescription(input)
		// Debería preservar caracteres válidos
		assert.Contains(t, result, "deployment", "Debe preservar palabras válidas")
	}
}
