package output

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestNew_OutputPretty_ReturnsPrettyFormatter(t *testing.T) {
	f := New(domain.OutputPretty, false)
	if _, ok := f.(*PrettyFormatter); !ok {
		t.Errorf("expected *PrettyFormatter, got %T", f)
	}
}

func TestNew_OutputJSON_ReturnsJSONFormatter(t *testing.T) {
	f := New(domain.OutputJSON, false)
	if _, ok := f.(*JSONFormatter); !ok {
		t.Errorf("expected *JSONFormatter, got %T", f)
	}
}

func TestNew_OutputRaw_ReturnsRawFormatter(t *testing.T) {
	f := New(domain.OutputRaw, false)
	if _, ok := f.(*RawFormatter); !ok {
		t.Errorf("expected *RawFormatter, got %T", f)
	}
}

func TestNew_OutputMinimal_ReturnsMinimalFormatter(t *testing.T) {
	f := New(domain.OutputMinimal, false)
	if _, ok := f.(*MinimalFormatter); !ok {
		t.Errorf("expected *MinimalFormatter, got %T", f)
	}
}

func TestNew_UnknownFormat_DefaultsToPrettyFormatter(t *testing.T) {
	f := New(domain.OutputFormat("unknown"), false)
	if _, ok := f.(*PrettyFormatter); !ok {
		t.Errorf("expected *PrettyFormatter for unknown format, got %T", f)
	}
}
