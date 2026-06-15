package datastore

import "context"

type SecretStore interface {
	Add(id string, secret []byte, expiresIn int) error
	Get(id string) (secret []byte, gotData bool)
	Delete(id string)
	DeleteExpiredSecrets(ctx context.Context)
	Close() error
}
