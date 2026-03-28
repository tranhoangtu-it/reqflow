package variable

import "strings"

// Interpolate replaces {{varName}} placeholders in template with values from vars.
// Variables not found in the map are left as-is. Empty var names and unclosed braces
// are left unchanged.
func Interpolate(template string, vars map[string]string) string {
	if template == "" {
		return ""
	}

	var result strings.Builder
	i := 0
	for i < len(template) {
		// Look for opening braces
		if i+1 < len(template) && template[i] == '{' && template[i+1] == '{' {
			// Find closing braces, scanning for a valid var name (no braces inside)
			found := false
			for j := i + 2; j < len(template); j++ {
				if template[j] == '{' {
					// Brace inside var name means this isn't a valid placeholder start.
					// Output one '{' and let the next iteration try matching from i+1.
					break
				}
				if j+1 < len(template) && template[j] == '}' && template[j+1] == '}' {
					varName := template[i+2 : j]
					if varName == "" {
						// Empty var name, leave as-is
						result.WriteString("{{}}")
					} else if val, ok := vars[varName]; ok {
						result.WriteString(val)
					} else {
						result.WriteString("{{")
						result.WriteString(varName)
						result.WriteString("}}")
					}
					i = j + 2
					found = true
					break
				}
			}
			if !found {
				// No valid closing found, output one character and move on
				result.WriteByte(template[i])
				i++
			}
		} else {
			result.WriteByte(template[i])
			i++
		}
	}
	return result.String()
}
