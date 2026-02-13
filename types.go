package rcodezeroacme

import "fmt"

type UpdateRRSet struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	ChangeType string   `json:"changetype"`
	Records    []Record `json:"records,omitempty"`
	TTL        int      `json:"ttl"`
}

type Record struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled,omitempty"`
}

type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (a APIResponse) Error() string {
	return fmt.Sprintf("%s: %s", a.Status, a.Message)
}

type GetRRsetsResponse struct {
	CurrentPage int     `json:"current_page"`
	Data        []RRSet `json:"data"`
	LastPage    int     `json:"last_page"`
	PerPage     int     `json:"per_page"`
	Total       int     `json:"total"`
	NextPageURL *string `json:"next_page_url"`
}

type RRSet struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	TTL     int      `json:"ttl"`
	Records []Record `json:"records"`
}

const (
	changeTypeAdd    = "add"
	changeTypeUpdate = "update"
	changeTypeDelete = "delete"
)
