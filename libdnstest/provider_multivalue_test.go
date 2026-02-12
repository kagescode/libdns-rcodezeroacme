package libdnstest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/libdns/libdns"

	rcodezero "github.com/kagescode/libdns-rcodezeroacme"
)

func randHexMV(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func TestACME_MultiValueNotSupported(t *testing.T) {
	cfg, ok := FromEnv()
	if !ok {
		t.Skip("set LIBDNSTEST_ZONE and LIBDNSTEST_API_TOKEN to run integration tests")
	}

	p := &rcodezero.Provider{
		APIToken: cfg.APIToken,
		BaseURL:  cfg.BaseURL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tokenA := "libdnstest-A-" + randHexMV(10)
	tokenB := "libdnstest-B-" + randHexMV(10)

	txtA := libdns.TXT{
		Name: "_acme-challenge",
		Text: tokenA,
		TTL:  60 * time.Second,
	}
	txtB := libdns.TXT{
		Name: "_acme-challenge",
		Text: tokenB,
		TTL:  60 * time.Second,
	}

	// Append A should work
	if _, err := p.AppendRecords(ctx, cfg.Zone, []libdns.Record{txtA}); err != nil {
		t.Fatalf("AppendRecords(A) failed: %v", err)
	}
	// Ensure cleanup even if the test fails later
	defer func() {
		_, _ = p.DeleteRecords(context.Background(), cfg.Zone, []libdns.Record{txtA})
	}()

	// Append B should fail (RcodeZero ACME endpoint doesn't allow additional values)
	_, err := p.AppendRecords(ctx, cfg.Zone, []libdns.Record{txtB})
	if err == nil {
		// If it ever starts supporting it, this test will tell you to revisit behavior.
		t.Fatalf("expected AppendRecords(B) to fail, but it succeeded")
	}

	// Assert it's the expected failure mode (keep this a bit loose)
	msg := err.Error()
	if !strings.Contains(strings.ToLower(msg), "already exists") && !strings.Contains(strings.ToLower(msg), "could not add additional") {
		t.Fatalf("AppendRecords(B) failed with unexpected error: %v", err)
	}

	// Cleanup A should still work
	if _, err := p.DeleteRecords(ctx, cfg.Zone, []libdns.Record{txtA}); err != nil {
		t.Fatalf("DeleteRecords(A) failed: %v", err)
	}
}

