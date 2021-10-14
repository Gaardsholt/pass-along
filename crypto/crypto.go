package crypto

import (
	"crypto/sha512"

	"github.com/Gaardsholt/pass-along/config"
	"golang.org/x/crypto/pbkdf2"
)

func deriveKey(passphrase string) []byte {
	return pbkdf2.Key([]byte(passphrase), []byte(config.Config.ServerSalt), 1000, 32, sha512.New)
}
