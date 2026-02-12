package rcodezeroacme

import (
	"testing"
	"time"

	"github.com/libdns/libdns"
)

func TestEnsureAcmeTXT_HardFailAndPass(t *testing.T) {
	zone := "example.com."

	// Non-TXT must fail
	_, _, _, err := ensureAcmeTXT(zone, libdns.CNAME{
		Name:   "_acme-challenge",
		Target: "x.example.com.",
		TTL:    time.Minute,
	})
	if err == nil {
		t.Fatalf("expected error for non-TXT record")
	}

	// Wrong name must fail
	_, _, _, err = ensureAcmeTXT(zone, libdns.TXT{
		Name: "www",
		Text: "nope",
		TTL:  time.Minute,
	})
	if err == nil {
		t.Fatalf("expected error for non-_acme-challenge TXT")
	}

	// Good must pass
	_, _, _, err = ensureAcmeTXT(zone, libdns.TXT{
		Name: "_acme-challenge",
		Text: "ok",
		TTL:  time.Minute,
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

