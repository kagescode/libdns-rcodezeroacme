package rcodezeroacme

import (
	"context"
	"strings"
)

// normalizeName normalizes rrset names for comparisons.
func normalizeName(n string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(n), "."))
}

// getExistingTXTValues returns (valueSet, ttl, error) for a TXT rrset name.
// rrsetFQDN can be "_acme-challenge.example.com." or "_acme-challenge".
func (p *Provider) getExistingTXTValues(ctx context.Context, zoneTrim string, rrsetFQDN string) (map[string]bool, int, error) {
	want := normalizeName(rrsetFQDN)

	page, pageSize := 1, 100
	values := map[string]bool{}
	ttl := 0

	for {
		resp, err := p.client.GetRRsets(ctx, zoneTrim, page, pageSize)
		if err != nil {
			return nil, 0, err
		}

		for _, rr := range resp.Data {
			if strings.ToUpper(rr.Type) != "TXT" {
				continue
			}

			// Accept both "_acme-challenge" and "_acme-challenge.<zone>."
			n := normalizeName(rr.Name)
			if n != want {
				// If caller passed "_acme-challenge.<zone>", but API returns "_acme-challenge" (or vice versa),
				// treat them as equivalent if both are ACME challenge names.
				if !(isAcmeChallengeName(rr.Name) && isAcmeChallengeName(rrsetFQDN)) {
					continue
				}
			}

			ttl = rr.TTL
			for _, r := range rr.Records {
				if r.Disabled {
					continue
				}
				values[unquoteTXT(r.Content)] = true
			}
			return values, ttl, nil
		}

		if resp.LastPage <= page || resp.LastPage == 0 {
			break
		}
		page++
	}

	return values, 0, nil // not found
}

