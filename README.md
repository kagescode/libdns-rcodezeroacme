# libdns-rcodezeroacme

ACME-only DNS provider for **RcodeZero** using the **libdns** interfaces.

This provider is intentionally minimal and only supports the DNS operations required for **ACME DNS-01 challenges**.

---

## Features

- Implements the libdns provider interfaces for ACME usage
- Supports creating and removing DNS-01 TXT challenge records
- Uses the dedicated RcodeZero ACME endpoint

Endpoint used:

- PATCH /api/v1/acme/zones/{zone}/rrsets
- GET   /api/v1/acme/zones/{zone}/rrsets

See the official RcodeZero OpenAPI specification:

https://my.rcodezero.at/openapi/

---

## Authentication

Authentication is done using an HTTP Bearer token:

Authorization: Bearer <TOKEN>

The token must have permission to manage ACME challenge records for the zone.

---

## Supported Records

This provider is ACME-only and supports:

- Record type: TXT
- Name must start with: _acme-challenge

All other record types or names will fail immediately.

Accepted names:

- _acme-challenge
- _acme-challenge.sub

Rejected names:

- www
- mail
- any A, AAAA, CNAME, MX, etc.

---

## Limitations

### Multi-value TXT records are NOT supported

The RcodeZero ACME endpoint does not allow multiple TXT values for the same rrset.

If a TXT record already exists, adding another value will fail with an error like:

RRset '_acme-challenge.<zone>. TXT' already exists, could not add additional

This matches the ACME endpoint behavior and is intentionally not worked around.

### No non-ACME record management

This provider does NOT support:

- Managing A/AAAA/CNAME/MX records
- Updating arbitrary zone records
- Full DNS provider semantics

It is strictly designed for ACME DNS-01 challenge automation.

---

## Usage (Go Example)

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/libdns/libdns"
	rcodezero "github.com/kagescode/libdns-rcodezeroacme"
)

func main() {
	provider := &rcodezero.Provider{
		APIToken: "your-token-here",
		BaseURL:  "https://my.rcodezero.at", // optional
	}

	ctx := context.Background()
	zone := "example.com."

	records := []libdns.Record{
		libdns.TXT{
			Name: "_acme-challenge",
			Text: "challenge-token-value",
			TTL:  60 * time.Second,
		},
	}

	// Present ACME TXT challenge
	_, err := provider.AppendRecords(ctx, zone, records)
	if err != nil {
		panic(err)
	}

	fmt.Println("TXT challenge record created")

	// Cleanup afterwards
	_, err = provider.DeleteRecords(ctx, zone, records)
	if err != nil {
		panic(err)
	}

	fmt.Println("TXT challenge record removed")
}
````

---

## Running Tests

### Compile + Unit Tests

Run:

```bash
go test ./...
go vet ./...
```

---

## Integration Tests

Integration tests require a real zone and an API token.

### Environment Variables

Set:

```bash
export LIBDNSTEST_ZONE="example.com."
export LIBDNSTEST_API_TOKEN="YOUR_TOKEN"
export LIBDNSTEST_BASE_URL="https://my.rcodezero.at"
```

### Run integration suite

```bash
go test ./libdnstest -v
```

Integration tests verify:

* TXT record creation under _acme-challenge
* Record cleanup after challenge completion
* Correct failure on unsupported multi-value rrsets

---

## CI

A GitHub Actions workflow is recommended to ensure:

* provider always compiles
* unit tests stay green

Minimal CI steps:

```yaml
go test ./...
go vet ./...
```

---

## License

MIT

---

## Disclaimer

This provider is designed only for ACME DNS-01 automation.

For full DNS record management, use the RcodeZero v2 API instead.