package variable

import "github.com/ye-kart/reqflow/internal/domain"

// Resolve merges variable layers by precedence. Later layers override earlier ones.
// This implements scope precedence: pass global first, then collection,
// then environment, then local.
func Resolve(layers ...[]domain.Variable) map[string]string {
	result := make(map[string]string)
	for _, layer := range layers {
		for _, v := range layer {
			result[v.Key] = v.Value
		}
	}
	return result
}
