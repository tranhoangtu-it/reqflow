package domain

import (
	"testing"
	"time"
)

func TestTimingInfo_ZeroValue(t *testing.T) {
	var ti TimingInfo
	if ti.DNSLookup != 0 {
		t.Errorf("expected zero DNSLookup, got %v", ti.DNSLookup)
	}
	if ti.TCPConnect != 0 {
		t.Errorf("expected zero TCPConnect, got %v", ti.TCPConnect)
	}
	if ti.TLSHandshake != 0 {
		t.Errorf("expected zero TLSHandshake, got %v", ti.TLSHandshake)
	}
	if ti.FirstByte != 0 {
		t.Errorf("expected zero FirstByte, got %v", ti.FirstByte)
	}
	if ti.Total != 0 {
		t.Errorf("expected zero Total, got %v", ti.Total)
	}
}

func TestHTTPResponse_HasTimingField(t *testing.T) {
	resp := HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Timing: TimingInfo{
			DNSLookup:    10 * time.Millisecond,
			TCPConnect:   20 * time.Millisecond,
			TLSHandshake: 30 * time.Millisecond,
			FirstByte:    50 * time.Millisecond,
			Total:        60 * time.Millisecond,
		},
	}

	if resp.Timing.Total != 60*time.Millisecond {
		t.Errorf("expected Total 60ms, got %v", resp.Timing.Total)
	}
	if resp.Timing.DNSLookup != 10*time.Millisecond {
		t.Errorf("expected DNSLookup 10ms, got %v", resp.Timing.DNSLookup)
	}
}
