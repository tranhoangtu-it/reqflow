package output

import (
	"fmt"
	"io"

	"github.com/ye-kart/reqflow/internal/domain"
)

func dim(s string) string { return "\033[2m" + s + "\033[0m" }

// FormatTrace writes a timing breakdown to w.
func FormatTrace(w io.Writer, timing domain.TimingInfo, noColor bool) error {
	lines := []struct {
		label string
		value int64
	}{
		{"DNS Lookup:", timing.DNSLookup.Milliseconds()},
		{"TCP Connect:", timing.TCPConnect.Milliseconds()},
		{"TLS Handshake:", timing.TLSHandshake.Milliseconds()},
		{"First Byte:", timing.FirstByte.Milliseconds()},
		{"Total:", timing.Total.Milliseconds()},
	}

	for _, l := range lines {
		line := fmt.Sprintf("  %-17s %dms", l.label, l.value)
		if !noColor {
			line = dim(line)
		}
		fmt.Fprintln(w, line)
	}

	return nil
}
