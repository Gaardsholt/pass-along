package secret

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"time"
)

type Secret struct {
	Content   string    `json:"content"`
	Expires   time.Time `json:"expires"`
	TimeAdded time.Time `json:"time_added"`
}

type SecretStore map[string]Secret

func new(content string, expires time.Time) Secret {
	return Secret{
		Content:   content,
		Expires:   expires,
		TimeAdded: time.Now(),
	}
}

func (s Secret) hash() string {
	checksum := sha512.Sum512([]byte(fmt.Sprintf("%v", s)))
	hash := base64.RawURLEncoding.EncodeToString(checksum[:])
	return hash
}

func (ss SecretStore) Add(content string, expiresIn int) (id string) {
	expires := time.Now().Add(
		time.Hour*time.Duration(0) +
			time.Minute*time.Duration(0) +
			time.Second*time.Duration(expiresIn),
	)

	mySecret := new(content, expires)
	id = mySecret.hash()
	ss[id] = mySecret

	return
}

func (ss SecretStore) Get(id string) (content string, gotData bool) {
	value, gotData := ss[id]
	if gotData {
		if value.Expires.After(time.Now().UTC()) {
			content = value.Content
		} else {
			gotData = false
		}

		delete(ss, id)
	}

	return content, gotData
}
