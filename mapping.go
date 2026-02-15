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

func normalizeTXT(s string) string {
    return unquoteTXT(strings.TrimSpace(s))
}

func ensureAcmeTXT(zone string, r libdns.Record) (fqdn string, txt string, ttlSec int, err error) {
	zoneFQDN := strings.TrimSuffix(strings.TrimSpace(zone), ".") + "."
	if zoneFQDN == "." {
		return "", "", 0, fmt.Errorf("empty zone")
	}

	var (
		relName string
		value   string
		ttl     time.Duration
	)

	switch v := r.(type) {
	case libdns.TXT:
		relName = v.Name
		value = v.Text
		ttl = v.TTL

	case libdns.RR:
		// Caddy/CertMagic may pass generic RR records.
		if !strings.EqualFold(strings.TrimSpace(v.Type), "TXT") {
			return "", "", 0, fmt.Errorf("acme-only provider supports only TXT records (got libdns.RR type=%s)", v.Type)
		}
		relName = v.Name
		value = v.Data
		ttl = v.TTL

	default:
		return "", "", 0, fmt.Errorf("acme-only provider supports only TXT records (got %T)", r)
	}

	abs := libdns.AbsoluteName(relName, zoneFQDN)

	if !isAcmeChallengeName(abs) {
		return "", "", 0, fmt.Errorf("only _acme-challenge TXT records are allowed (got %q)", abs)
	}

	// Keep trailing dot in name (API examples use it)
	fqdn = abs

	ttlSec = durationToSeconds(ttl)
	if ttlSec == 0 {
		ttlSec = 60
	}

	// Keep value as-is (your provider already normalizes on GetRecords via unquoteTXT()).
	// If you want to force quoting for the API, do it here consistently.
	txt = value

	return fqdn, txt, ttlSec, nil
}

func durationToSeconds(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	return int((d + (time.Second - 1)) / time.Second)
}
