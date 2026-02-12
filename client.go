package rcodezeroacme

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://my.rcodezero.at"

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	apiToken   string
	baseURL    *url.URL
	httpClient HTTPClient
	timeout    time.Duration
}

func NewClient(apiToken, baseURL string, hc HTTPClient) (*Client, error) {
	if strings.TrimSpace(apiToken) == "" {
		return nil, fmt.Errorf("APIToken is required")
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base url: %w", err)
	}
	return &Client{
		apiToken:   apiToken,
		baseURL:    u,
		httpClient: hc,
		timeout:    10 * time.Second,
	}, nil
}

func (c *Client) do(req *http.Request, out any) error {
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Accept", "application/json")

	client := c.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http do: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		// Spec doesn’t document a structured ACME error payload; keep raw.
		return fmt.Errorf("rcodezero acme http %d: %s", resp.StatusCode, string(raw))
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	// If out is *APIResponse, treat non-ok as error.
	if r, ok := out.(*APIResponse); ok {
		if strings.ToLower(r.Status) != "ok" {
			return *r
		}
	}

	return nil
}

func (c *Client) GetRRsets(ctx context.Context, zone string, page, pageSize int) (*GetRRsetsResponse, error) {
	zone = strings.TrimSuffix(strings.TrimSpace(zone), ".")
	if zone == "" {
		return nil, fmt.Errorf("empty zone")
	}

	// /api/v1/acme/zones/{zone}/rrsets  (paginated) :contentReference[oaicite:6]{index=6}
	endpoint := c.baseURL.JoinPath("api", "v1", "acme", "zones", zone, "rrsets")
	q := endpoint.Query()
	if page > 0 {
		q.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		q.Set("page_size", fmt.Sprintf("%d", pageSize))
	}
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	var out GetRRsetsResponse
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) PatchRRsets(ctx context.Context, zone string, sets []UpdateRRSet) (*APIResponse, error) {
	zone = strings.TrimSuffix(strings.TrimSpace(zone), ".")
	if zone == "" {
		return nil, fmt.Errorf("empty zone")
	}

	// /api/v1/acme/zones/{zone}/rrsets  PATCH :contentReference[oaicite:7]{index=7}
	endpoint := c.baseURL.JoinPath("api", "v1", "acme", "zones", zone, "rrsets")

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(sets); err != nil {
		return nil, fmt.Errorf("encode json: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint.String(), buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var out APIResponse
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
