package domain

// VariableScope represents the precedence level of a variable.
// Higher numeric values take precedence over lower ones.
type VariableScope int

const (
	ScopeGlobal     VariableScope = iota // lowest precedence
	ScopeCollection                      // collection-level
	ScopeEnvironment                     // environment-level
	ScopeLocal                           // highest precedence
)

// Variable represents a key-value pair with an associated scope.
type Variable struct {
	Key   string
	Value string
	Scope VariableScope
}

// Environment represents a named set of variables (e.g., "dev", "staging", "prod").
type Environment struct {
	Name      string
	Variables []Variable
}
