package auth

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestParseWWWAuthenticate(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   DigestChallenge
	}{
		{
			name:   "standard header with all fields",
			header: `Digest realm="testrealm@host.com", qop="auth", nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093", opaque="5ccc069c403ebaf9f0171e9517f40e41"`,
			want: DigestChallenge{
				Realm:  "testrealm@host.com",
				Nonce:  "dcd98b7102dd2f0e8b11d0f600bfb0c093",
				QOP:    "auth",
				Opaque: "5ccc069c403ebaf9f0171e9517f40e41",
			},
		},
		{
			name:   "header without opaque",
			header: `Digest realm="myrealm", nonce="abc123", qop="auth"`,
			want: DigestChallenge{
				Realm: "myrealm",
				Nonce: "abc123",
				QOP:   "auth",
			},
		},
		{
			name:   "header with algorithm",
			header: `Digest realm="test", nonce="xyz", qop="auth", algorithm=MD5`,
			want: DigestChallenge{
				Realm:     "test",
				Nonce:     "xyz",
				QOP:       "auth",
				Algorithm: "MD5",
			},
		},
		{
			name:   "header with extra whitespace",
			header: `Digest  realm="test" ,  nonce="abc" ,  qop="auth"`,
			want: DigestChallenge{
				Realm: "test",
				Nonce: "abc",
				QOP:   "auth",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseWWWAuthenticate(tt.header)

			if got.Realm != tt.want.Realm {
				t.Errorf("Realm = %q, want %q", got.Realm, tt.want.Realm)
			}
			if got.Nonce != tt.want.Nonce {
				t.Errorf("Nonce = %q, want %q", got.Nonce, tt.want.Nonce)
			}
			if got.QOP != tt.want.QOP {
				t.Errorf("QOP = %q, want %q", got.QOP, tt.want.QOP)
			}
			if got.Opaque != tt.want.Opaque {
				t.Errorf("Opaque = %q, want %q", got.Opaque, tt.want.Opaque)
			}
			if got.Algorithm != tt.want.Algorithm {
				t.Errorf("Algorithm = %q, want %q", got.Algorithm, tt.want.Algorithm)
			}
		})
	}
}

func TestComputeDigestResponse(t *testing.T) {
	// RFC 7616 / RFC 2617 example-based test
	challenge := DigestChallenge{
		Realm:  "testrealm@host.com",
		Nonce:  "dcd98b7102dd2f0e8b11d0f600bfb0c093",
		QOP:    "auth",
		Opaque: "5ccc069c403ebaf9f0171e9517f40e41",
	}

	result := ComputeDigestResponse(challenge, "GET", "/dir/index.html", "Mufasa", "Circle Of Life", "0a4f113b", 1)

	// Verify the result contains the expected components
	if result == "" {
		t.Fatal("expected non-empty digest response")
	}

	// The response should be a properly formatted Digest Authorization header value
	// containing realm, nonce, uri, qop, nc, cnonce, response, and opaque
	for _, want := range []string{"realm=", "nonce=", "uri=", "qop=auth", "nc=00000001", "cnonce=", "response=", "opaque="} {
		if !containsStr(result, want) {
			t.Errorf("result missing %q, got: %s", want, result)
		}
	}
}

func TestComputeDigestResponse_KnownVector(t *testing.T) {
	// From RFC 2617 section 3.5
	// HA1 = MD5("Mufasa:testrealm@host.com:Circle Of Life")
	//      = 939e7578ed9e3c518a452acee763bce9
	// HA2 = MD5("GET:/dir/index.html")
	//      = 39aff3a2bab6126f332b942af5e6afc3
	// response = MD5("939e7578ed9e3c518a452acee763bce9:dcd98b7102dd2f0e8b11d0f600bfb0c093:00000001:0a4f113b:auth:39aff3a2bab6126f332b942af5e6afc3")
	//          = 6629fae49393a05397450978507c4ef1
	challenge := DigestChallenge{
		Realm:  "testrealm@host.com",
		Nonce:  "dcd98b7102dd2f0e8b11d0f600bfb0c093",
		QOP:    "auth",
		Opaque: "5ccc069c403ebaf9f0171e9517f40e41",
	}

	result := ComputeDigestResponse(challenge, "GET", "/dir/index.html", "Mufasa", "Circle Of Life", "0a4f113b", 1)

	if !containsStr(result, "response=\"6629fae49393a05397450978507c4ef1\"") {
		t.Errorf("digest response hash mismatch, got: %s", result)
	}
}

func TestApplyDigest(t *testing.T) {
	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://example.com/dir/index.html",
		Headers: []domain.Header{
			{Key: "Accept", Value: "application/json"},
		},
	}

	challenge := DigestChallenge{
		Realm:  "testrealm@host.com",
		Nonce:  "dcd98b7102dd2f0e8b11d0f600bfb0c093",
		QOP:    "auth",
		Opaque: "5ccc069c403ebaf9f0171e9517f40e41",
	}

	got := ApplyDigest(req, challenge, "Mufasa", "Circle Of Life")

	// Should have Authorization header
	var authValue string
	for _, h := range got.Headers {
		if h.Key == "Authorization" {
			authValue = h.Value
			break
		}
	}
	if authValue == "" {
		t.Fatal("expected Authorization header, got none")
	}
	if !containsStr(authValue, "Digest ") {
		t.Errorf("expected Authorization to start with 'Digest ', got: %s", authValue)
	}

	// Verify existing headers preserved
	found := false
	for _, h := range got.Headers {
		if h.Key == "Accept" && h.Value == "application/json" {
			found = true
			break
		}
	}
	if !found {
		t.Error("existing Accept header was not preserved")
	}

	// Verify original request was not modified
	for _, h := range req.Headers {
		if h.Key == "Authorization" {
			t.Error("original request was modified")
		}
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
