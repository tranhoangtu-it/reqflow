package workflow

import (
	"fmt"
	"strings"

	"github.com/ye-kart/reqflow/internal/domain"
	"gopkg.in/yaml.v3"
)

// rawWorkflow is the YAML-friendly intermediate representation.
type rawWorkflow struct {
	Name  string    `yaml:"name"`
	Env   string    `yaml:"env"`
	Steps []rawStep `yaml:"steps"`
}

type rawStep struct {
	Name        string            `yaml:"name"`
	Method      string            `yaml:"method"`
	URL         string            `yaml:"url"`
	Headers     map[string]string `yaml:"headers"`
	Body        interface{}       `yaml:"body"`
	Extract     map[string]string `yaml:"extract"`
	Assert      []rawAssertion    `yaml:"assert"`
	Auth        *rawAuth          `yaml:"auth"`
	ContentType string            `yaml:"content_type"`
}

type rawAssertion struct {
	Field    string      `yaml:"field"`
	Operator string      `yaml:"operator"`
	Expected interface{} `yaml:"expected"`
}

type rawAuth struct {
	Type   string `yaml:"type"`
	Token  string `yaml:"token"`
	User   string `yaml:"user"`
	Pass   string `yaml:"pass"`
	Key    string `yaml:"key"`
	Value  string `yaml:"value"`
	In     string `yaml:"in"`
	Prefix string `yaml:"prefix"`
}

// Parse parses YAML data into a Workflow domain type.
// It validates required fields: name, at least one step, and each step
// must have name, method, and url.
func Parse(data []byte) (domain.Workflow, error) {
	var raw rawWorkflow
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return domain.Workflow{}, fmt.Errorf("parsing workflow YAML: %w", err)
	}

	if raw.Name == "" {
		return domain.Workflow{}, fmt.Errorf("workflow name is required")
	}
	if len(raw.Steps) == 0 {
		return domain.Workflow{}, fmt.Errorf("workflow must have at least one step")
	}

	steps := make([]domain.Step, 0, len(raw.Steps))
	for i, rs := range raw.Steps {
		step, err := convertStep(rs, i)
		if err != nil {
			return domain.Workflow{}, err
		}
		steps = append(steps, step)
	}

	return domain.Workflow{
		Name:  raw.Name,
		Env:   raw.Env,
		Steps: steps,
	}, nil
}

func convertStep(rs rawStep, index int) (domain.Step, error) {
	if rs.Name == "" {
		return domain.Step{}, fmt.Errorf("step %d: name is required", index)
	}
	if rs.Method == "" {
		return domain.Step{}, fmt.Errorf("step %q: method is required", rs.Name)
	}
	if rs.URL == "" {
		return domain.Step{}, fmt.Errorf("step %q: url is required", rs.Name)
	}

	method := domain.HTTPMethod(strings.ToUpper(rs.Method))
	if !method.IsValid() {
		return domain.Step{}, fmt.Errorf("step %q: invalid method %q", rs.Name, rs.Method)
	}

	assertions := make([]domain.Assertion, 0, len(rs.Assert))
	for _, ra := range rs.Assert {
		assertions = append(assertions, domain.Assertion{
			Field:    ra.Field,
			Operator: ra.Operator,
			Expected: ra.Expected,
		})
	}

	var authConfig *domain.AuthConfig
	if rs.Auth != nil {
		ac, err := convertAuth(rs.Auth)
		if err != nil {
			return domain.Step{}, fmt.Errorf("step %q: %w", rs.Name, err)
		}
		authConfig = ac
	}

	return domain.Step{
		Name:        rs.Name,
		Method:      method,
		URL:         rs.URL,
		Headers:     rs.Headers,
		Body:        rs.Body,
		Extract:     rs.Extract,
		Assert:      assertions,
		Auth:        authConfig,
		ContentType: rs.ContentType,
	}, nil
}

func convertAuth(ra *rawAuth) (*domain.AuthConfig, error) {
	switch ra.Type {
	case "basic":
		return &domain.AuthConfig{
			Type: domain.AuthBasic,
			Basic: &domain.BasicAuthConfig{
				Username: ra.User,
				Password: ra.Pass,
			},
		}, nil
	case "bearer":
		return &domain.AuthConfig{
			Type: domain.AuthBearer,
			Bearer: &domain.BearerAuthConfig{
				Token:  ra.Token,
				Prefix: ra.Prefix,
			},
		}, nil
	case "apikey":
		location := domain.APIKeyInHeader
		if ra.In == "query" {
			location = domain.APIKeyInQuery
		}
		return &domain.AuthConfig{
			Type: domain.AuthAPIKey,
			APIKey: &domain.APIKeyAuthConfig{
				Key:      ra.Key,
				Value:    ra.Value,
				Location: location,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported auth type %q", ra.Type)
	}
}
