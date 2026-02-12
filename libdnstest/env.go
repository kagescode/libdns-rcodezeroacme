package libdnstest

import (
	"os"
	"strings"
)

type Config struct {
	Zone     string
	APIToken string
	BaseURL  string
}

func FromEnv() (Config, bool) {
	zone := strings.TrimSpace(os.Getenv("LIBDNSTEST_ZONE"))
	token := strings.TrimSpace(os.Getenv("LIBDNSTEST_API_TOKEN"))
	base := strings.TrimSpace(os.Getenv("LIBDNSTEST_BASE_URL"))

	if zone == "" || token == "" {
		return Config{}, false
	}
	if !strings.HasSuffix(zone, ".") {
		zone += "."
	}
	return Config{
		Zone:     zone,
		APIToken: token,
		BaseURL:  base,
	}, true
}

