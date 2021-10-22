package datastore

type SecretStore interface {
	Add(id string, secret []byte, expiresIn int) error
	Get(id string) (secret []byte, gotData bool)
	Delete(id string)
	DeleteExpiredSecrets()
}
