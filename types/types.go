package types

import (
	"sync"
)

type Page struct {
	Content string `json:"content"`
}
type Entry struct {
	Content        string            `json:"content"`
	ExpiresIn      int               `json:"expires_in"`
	UnlimitedViews bool              `json:"unlimited_views"`
	Files          map[string][]byte `json:"files"`
}

type SecretStore struct {
	Data map[string][]byte
	Lock *sync.RWMutex
}
