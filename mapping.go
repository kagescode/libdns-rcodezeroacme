package rcodezeroacme

import (
	"fmt"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

const acmeLabelPrefix = "_acme-challenge."

func isAcmeChallengeName(name string) bool {
	n := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(name), "."))
	if n == "_acme-challenge" {
		return true
	}
	return strings.HasPrefix(n, "_acme-challenge.")
}

func unquoteTXT(s string) string {
    s = strings.TrimSpace(s)
    if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
        return s[1 : len(s)-1]
    }
    return s
}

func ensureAcmeTXT(zone string, r libdns.Record) (fqdn string, txt string, ttlSec int, err error) {
	zoneFQDN := strings.TrimSuffix(strings.TrimSpace(zone), ".") + "."
	if zoneFQDN == "." {
		return "", "", 0, fmt.Errorf("empty zone")
	}

	// Hard fail: ACME-only provider supports only TXT records.
	v, ok := r.(libdns.TXT)
	if !ok {
		return "", "", 0, fmt.Errorf("acme-only provider supports only TXT records (got %T)", r)
	}

	abs := libdns.AbsoluteName(v.Name, zoneFQDN)

	absLower := strings.ToLower(abs)
	if !strings.HasPrefix(absLower, "_acme-challenge.") {
		return "", "", 0, fmt.Errorf("only _acme-challenge TXT records are allowed (got %q)", abs)
	}

	// Keep trailing dot in name (API examples use it)
	fqdn = abs
	txt = v.Text

	ttlSec = durationToSeconds(v.TTL)
	if ttlSec == 0 {
		ttlSec = 60
	}

	return fqdn, txt, ttlSec, nil
}


func durationToSeconds(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	return int((d + (time.Second - 1)) / time.Second)
}
