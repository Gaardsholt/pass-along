package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// SecretsRead counts the total number of secrets that have been read.
	// It is a Prometheus counter metric used for monitoring read operations on secrets.
	SecretsRead = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "secrets_read",
		Help: "Number of secrets that have been read.",
	})
	// ExpiredSecretsRead counts the number of attempts to read expired secrets.
	// This Prometheus counter helps monitor how often expired secrets are accessed.
	ExpiredSecretsRead = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "expired_secrets_read",
		Help: "Number of expired secrets that have been attempted to read.",
	})
	// NonExistentSecretsRead counts the number of attempts to read secrets that do not exist.
	// This Prometheus counter helps track how often clients try to access missing secrets.
	NonExistentSecretsRead = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "nonexistent_secrets_read",
		Help: "Number of non-existent secrets that have been attempted to read.",
	})
	// SecretsCreated counts the total number of secrets that have been created.
	// It is a Prometheus counter metric used for monitoring secret creation events.
	SecretsCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "secrets_created",
		Help: "Number of secrets that have been created.",
	})
	// SecretsCreatedWithError counts the number of failed attempts to create a secret.
	// This Prometheus counter is incremented each time a secret creation operation fails.
	SecretsCreatedWithError = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "secrets_created_with_errors",
		Help: "Number of attempts to create a secret but it failed.",
	})
	// SecretsDeleted counts the total number of secrets that have been deleted.
	// This Prometheus counter is used to track deletion events for monitoring purposes.
	SecretsDeleted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "secrets_deleted",
		Help: "Number of secrets that have been deleted.",
	})
)
