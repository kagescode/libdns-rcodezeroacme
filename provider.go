package rcodezeroacme

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

type Provider struct {
	APIToken string
	BaseURL  string

	HTTPClient HTTPClient

	client *Client
}

func (p *Provider) init() error {
	if p.client != nil {
		return nil
	}
	c, err := NewClient(p.APIToken, p.BaseURL, p.HTTPClient)
	if err != nil {
		return err
	}
	p.client = c
	return nil
}

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	if err := p.init(); err != nil {
		return nil, err
	}
	zoneTrim := strings.TrimSuffix(strings.TrimSpace(zone), ".")
	if zoneTrim == "" {
		return nil, fmt.Errorf("empty zone")
	}

	var out []libdns.Record
	page := 1
	pageSize := 100

	for {
		resp, err := p.client.GetRRsets(ctx, zoneTrim, page, pageSize)
		if err != nil {
			return nil, err
		}

		for _, rrset := range resp.Data {
			if strings.ToUpper(rrset.Type) != "TXT" {
				continue
			}
                        if !isAcmeChallengeName(rrset.Name) {
                            continue
                        }

			for _, rec := range rrset.Records {
				if rec.Disabled {
					continue
				}
				// Map back to libdns TXT
				nameRel := libdns.RelativeName(rrset.Name, zoneTrim+".")
				out = append(out, libdns.TXT{
					Name: nameRel,
					Text: normalizeTXT(rec.Content),
					TTL:  timeSeconds(rrset.TTL),
				})
			}
		}

		if resp.LastPage <= page || resp.LastPage == 0 {
			break
		}
		page++
	}

	return out, nil
}

func (p *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	if err := p.init(); err != nil {
		return nil, err
	}
	zoneTrim := strings.TrimSuffix(strings.TrimSpace(zone), ".")
	if zoneTrim == "" {
		return nil, fmt.Errorf("empty zone")
	}

	for _, r := range recs {
		fqdn, txt, ttl, err := ensureAcmeTXT(zoneTrim, r)
		if err != nil {
			return nil, err
		}
		txt = normalizeTXT(txt)

		// Check existing rrset values
		existing, existingTTL, err := p.getExistingTXTValues(ctx, zoneTrim, fqdn)
		if err != nil {
			return nil, err
		}

		if len(existing) == 0 {
			// rrset doesn't exist -> ADD
			sets := []UpdateRRSet{{
				Name:       fqdn,
				Type:       "TXT",
				TTL:        ttl,
				ChangeType: changeTypeAdd,
				Records: []Record{{
					Content:  txt,
					Disabled: false,
				}},
			}}

			if _, err := p.client.PatchRRsets(ctx, zoneTrim, sets); err != nil {
				return nil, err
			}
			continue
		}

		// rrset exists -> UPDATE with merged set
		existing[txt] = true

		merged := make([]Record, 0, len(existing))
		for v := range existing {
			merged = append(merged, Record{Content: v, Disabled: false})
		}

		ttlToUse := ttl
		if existingTTL > 0 {
			ttlToUse = existingTTL
		}

		sets := []UpdateRRSet{{
			Name:       fqdn,
			Type:       "TXT",
			TTL:        ttlToUse,
			ChangeType: changeTypeUpdate,
			Records:    merged,
		}}

		if _, err := p.client.PatchRRsets(ctx, zoneTrim, sets); err != nil {
			return nil, err
		}
	}

	return recs, nil
}

func (p *Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	if err := p.init(); err != nil {
		return nil, err
	}
	zoneTrim := strings.TrimSuffix(strings.TrimSpace(zone), ".")
	if zoneTrim == "" {
		return nil, fmt.Errorf("empty zone")
	}

	for _, r := range recs {
		fqdn, txt, ttl, err := ensureAcmeTXT(zoneTrim, r)
		if err != nil {
			return nil, err
		}
		txt = normalizeTXT(txt)

		// IMPORTANT:
		// Remove exactly this TXT value using changetype=delete with records payload.
		// Do NOT rely on changetype=update to remove a missing value (it doesn't work here).
		sets := []UpdateRRSet{{
			Name:       fqdn,
			Type:       "TXT",
			TTL:        ttl,
			ChangeType: changeTypeDelete,
			Records: []Record{{
				Content:  txt,
				Disabled: false,
			}},
		}}

		if _, err := p.client.PatchRRsets(ctx, zoneTrim, sets); err != nil {
			return nil, err
		}
	}

	return recs, nil
}

// SetRecords is defined by libdns; for ACME usage you usually don't need it.
// Implement as "replace this rrset's values" by:
// - delete existing rrset label (not value-specific, depending on backend)
// - add all desired values
//
// The ACME endpoint description implies only add/delete operations on the label and errors on other labels. :contentReference[oaicite:11]{index=11}
func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	// safest behavior: just Append (ACME clients generally do Present/CleanUp)
	return p.AppendRecords(ctx, zone, recs)
}

func timeSeconds(ttl int) time.Duration {
	if ttl <= 0 {
		return 0
	}
	return time.Duration(ttl) * time.Second
}
