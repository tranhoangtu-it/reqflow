package variable

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name   string
		layers [][]domain.Variable
		want   map[string]string
	}{
		{
			name:   "no layers returns empty map",
			layers: nil,
			want:   map[string]string{},
		},
		{
			name: "single layer converts to map",
			layers: [][]domain.Variable{
				{
					{Key: "a", Value: "1"},
				},
			},
			want: map[string]string{"a": "1"},
		},
		{
			name: "later layer overrides earlier",
			layers: [][]domain.Variable{
				{
					{Key: "a", Value: "1"},
				},
				{
					{Key: "a", Value: "2"},
				},
			},
			want: map[string]string{"a": "2"},
		},
		{
			name: "merge layers with different keys",
			layers: [][]domain.Variable{
				{
					{Key: "a", Value: "1"},
				},
				{
					{Key: "b", Value: "2"},
				},
			},
			want: map[string]string{"a": "1", "b": "2"},
		},
		{
			name: "three layers with override chain",
			layers: [][]domain.Variable{
				{
					{Key: "a", Value: "global"},
					{Key: "b", Value: "global"},
				},
				{
					{Key: "a", Value: "collection"},
					{Key: "c", Value: "collection"},
				},
				{
					{Key: "a", Value: "local"},
				},
			},
			want: map[string]string{
				"a": "local",
				"b": "global",
				"c": "collection",
			},
		},
		{
			name: "empty value is valid",
			layers: [][]domain.Variable{
				{
					{Key: "empty", Value: ""},
				},
			},
			want: map[string]string{"empty": ""},
		},
		{
			name: "duplicate in same layer last wins",
			layers: [][]domain.Variable{
				{
					{Key: "dup", Value: "first"},
					{Key: "dup", Value: "second"},
				},
			},
			want: map[string]string{"dup": "second"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(tt.layers...)
			if len(got) != len(tt.want) {
				t.Errorf("Resolve() returned %d entries, want %d\ngot: %v\nwant: %v", len(got), len(tt.want), got, tt.want)
				return
			}
			for k, wantV := range tt.want {
				gotV, ok := got[k]
				if !ok {
					t.Errorf("Resolve() missing key %q", k)
					continue
				}
				if gotV != wantV {
					t.Errorf("Resolve()[%q] = %q, want %q", k, gotV, wantV)
				}
			}
		})
	}
}
