package types

import "time"

type Page struct {
	Content string    `json:"content"`
	Startup time.Time `json:"startup"`
}
type Entry struct {
	Content        string `json:"content"`
	ExpiresIn      int    `json:"expires_in"`
	UnlimitedViews bool   `json:"unlimited_views"`
}
