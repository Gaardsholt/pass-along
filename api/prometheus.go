package api

import (
	"github.com/Gaardsholt/pass-along/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var pr *prometheus.Registry

func registerPrometheusMetrics() {
	pr = prometheus.NewRegistry()
	pr.MustRegister(metrics.SecretsRead)
	pr.MustRegister(metrics.ExpiredSecretsRead)
	pr.MustRegister(metrics.NonExistentSecretsRead)
	pr.MustRegister(metrics.SecretsCreated)
	pr.MustRegister(metrics.SecretsCreatedWithError)
	pr.MustRegister(metrics.SecretsDeleted)
}
