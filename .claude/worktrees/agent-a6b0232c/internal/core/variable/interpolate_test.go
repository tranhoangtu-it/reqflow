package variable

import "testing"

func TestInterpolate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]string
		want     string
	}{
		{
			name:     "simple replacement",
			template: "Hello {{name}}",
			vars:     map[string]string{"name": "World"},
			want:     "Hello World",
		},
		{
			name:     "multiple variables",
			template: "{{a}}/{{b}}",
			vars:     map[string]string{"a": "x", "b": "y"},
			want:     "x/y",
		},
		{
			name:     "missing variable left as-is",
			template: "{{unknown}}",
			vars:     map[string]string{},
			want:     "{{unknown}}",
		},
		{
			name:     "no variables in template",
			template: "plain text",
			vars:     map[string]string{},
			want:     "plain text",
		},
		{
			name:     "empty template",
			template: "",
			vars:     map[string]string{},
			want:     "",
		},
		{
			name:     "adjacent variables",
			template: "{{a}}{{b}}",
			vars:     map[string]string{"a": "1", "b": "2"},
			want:     "12",
		},
		{
			name:     "empty var name left as-is",
			template: "{{}}",
			vars:     map[string]string{},
			want:     "{{}}",
		},
		{
			name:     "unclosed brace left as-is",
			template: "{{unclosed",
			vars:     map[string]string{},
			want:     "{{unclosed",
		},
		{
			name:     "URL pattern",
			template: "https://{{host}}/api/{{version}}",
			vars:     map[string]string{"host": "example.com", "version": "v2"},
			want:     "https://example.com/api/v2",
		},
		{
			name:     "special chars in value",
			template: "{{token}}",
			vars:     map[string]string{"token": "abc=123&x"},
			want:     "abc=123&x",
		},
		{
			name:     "nested braces",
			template: "{{{var}}}",
			vars:     map[string]string{"var": "value"},
			want:     "{value}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Interpolate(tt.template, tt.vars)
			if got != tt.want {
				t.Errorf("Interpolate(%q, %v) = %q, want %q", tt.template, tt.vars, got, tt.want)
			}
		})
	}
}
