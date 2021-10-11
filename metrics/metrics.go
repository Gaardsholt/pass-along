package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	SecretsRead = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "secrets_read",
		Help: "Number of secrets that has been read.",
	})
	ExpiredSecretsRead = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "expired_secrets_read",
		Help: "Number of expired secrets that has been attempted to read.",
	})
	NonExistentSecretsRead = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "nonexistent_secrets_read",
		Help: "Number of non-existent secrets that has been attempted to read.",
	})
	SecretsCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "secrets_created",
		Help: "Number of secrets that has been created.",
	})
	SecretsCreatedWithError = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "secrets_created_with_errors",
		Help: "Number of attempts to create a secret but it failed.",
	})
	SecretsDeleted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "secrets_deleted",
		Help: "Number of secrets that has been deleted.",
	})
)
