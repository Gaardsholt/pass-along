package datastore

import "github.com/Gaardsholt/pass-along/types"

type SecretStore interface {
	Add(entry types.Entry) (id string, err error)
	Get(id string) (content string, gotData bool)
	Delete(id string)
	DeleteExpiredSecrets()
}
