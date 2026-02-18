package tailscale

import (
	"regexp"
	"strings"
)

// SanitizeDescription sanitiza la descripción de auth keys según las reglas de Tailscale
// Permite solo: a-z A-Z 0-9 espacio - _ . :
// Todo lo demás se convierte a -
func SanitizeDescription(desc string) string {
	// 1. Trim espacios
	desc = strings.TrimSpace(desc)
	
	// 2. Si está vacío, usar default
	if desc == "" {
		return "ark-auth-key"
	}
	
	// 3. Reemplazar caracteres no permitidos por -
	// Permite: letras, números, espacio, - _ . :
	reg := regexp.MustCompile(`[^a-zA-Z0-9 \-_.:]+`)
	desc = reg.ReplaceAllString(desc, "-")
	
	// 4. Reemplazar múltiples - consecutivos por uno solo
	reg2 := regexp.MustCompile(`-+`)
	desc = reg2.ReplaceAllString(desc, "-")
	
	// 5. Remover - al inicio o final
	desc = strings.Trim(desc, "-")
	
	// 6. Limitar longitud a 100 caracteres
	if len(desc) > 100 {
		desc = desc[:100]
		desc = strings.TrimRight(desc, "-")
	}
	
	// 7. Si después de todo quedó vacío, usar default
	if desc == "" {
		return "ark-auth-key"
	}
	
	return desc
}
