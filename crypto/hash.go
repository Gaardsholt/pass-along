package crypto

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
)

func Hash(data interface{}) string {
	checksum := sha512.Sum512([]byte(fmt.Sprintf("%v", data)))
	hash := base64.RawURLEncoding.EncodeToString(checksum[:])
	return hash
}
