package types

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Page struct {
	Content string    `json:"content"`
	Startup time.Time `json:"startup"`
}
type Entry struct {
	Content        string `json:"content"`
	ExpiresIn      int    `json:"expires_in"`
	UnlimitedViews bool   `json:"unlimited_views"`
}

type SecretStore struct {
	Data map[string][]byte
	Lock *sync.RWMutex
}

// Prometheus stuff
type SecretsInCache struct {
	counterDesc *prometheus.Desc
	ss          *SecretStore
}

func (c *SecretsInCache) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.counterDesc
}

func (c *SecretsInCache) Collect(ch chan<- prometheus.Metric) {
	value := float64(len(c.ss.Data)) // Your code to fetch the counter value goes here.
	ch <- prometheus.MustNewConstMetric(
		c.counterDesc,
		prometheus.CounterValue,
		value,
	)
}

func NewSecretsInCache(ss *SecretStore) *SecretsInCache {
	return &SecretsInCache{
		counterDesc: prometheus.NewDesc("secrets_in_cache", "Current number of secrets in the cache.", nil, nil),
		ss:          ss,
	}
}
