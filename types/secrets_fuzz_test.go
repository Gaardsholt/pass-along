package types

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func FuzzParseToken(f *testing.F) {
	lookupID, accessKey, token, err := GenerateToken()
	if err != nil {
		f.Fatalf("expected token generation to succeed: %v", err)
	}

	f.Add(token)
	f.Add(lookupID + "." + accessKey)
	f.Add("")
	f.Add("invalid-token")
	f.Add(".")
	f.Add("too.many.parts")
	f.Add(" " + lookupID + " . " + accessKey + " ")
	f.Add("not base64." + accessKey)
	f.Add(lookupID + ".not base64")

	f.Fuzz(func(t *testing.T, token string) {
		lookupID, accessKey, err := ParseToken(token)
		if err != nil {
			return
		}

		if lookupID == "" || accessKey == "" {
			t.Fatalf("expected parsed token parts to be non-empty")
		}
		if strings.TrimSpace(lookupID) != lookupID || strings.TrimSpace(accessKey) != accessKey {
			t.Fatalf("expected parsed token parts to be trimmed")
		}
		if _, err := base64.RawURLEncoding.DecodeString(lookupID); err != nil {
			t.Fatalf("expected lookup id to be valid base64url: %v", err)
		}
		if _, err := base64.RawURLEncoding.DecodeString(accessKey); err != nil {
			t.Fatalf("expected access key to be valid base64url: %v", err)
		}
	})
}

func FuzzSecretEncryptDecryptRoundTrip(f *testing.F) {
	f.Add("secret", "passphrase", int64(60), []byte("file contents"), false)
	f.Add("", "", int64(0), []byte{}, true)
	f.Add("unicode secret", "another-key", int64(86400), []byte{0, 1, 2, 3}, false)

	f.Fuzz(func(t *testing.T, content string, passphrase string, expiresIn int64, fileBytes []byte, unlimitedViews bool) {
		if len(content) > 1024 || len(passphrase) > 256 || len(fileBytes) > 512 {
			t.Skip()
		}

		baseTime := time.Unix(1700000000, 0).UTC()
		secret := Secret{
			Content:        content,
			Files:          map[string][]byte{},
			Expires:        baseTime.Add(time.Duration(expiresIn%86400) * time.Second),
			TimeAdded:      baseTime,
			UnlimitedViews: unlimitedViews,
		}
		if len(fileBytes) > 0 {
			secret.Files["fuzz.bin"] = fileBytes
		}

		encryptedData, err := secret.Encrypt(passphrase)
		if err != nil {
			t.Fatalf("expected encryption to succeed: %v", err)
		}

		decryptedSecret, err := Decrypt(encryptedData, passphrase)
		if err != nil {
			t.Fatalf("expected decryption to succeed: %v", err)
		}

		if decryptedSecret.Content != secret.Content {
			t.Fatalf("unexpected content after round trip")
		}
		if decryptedSecret.UnlimitedViews != secret.UnlimitedViews {
			t.Fatalf("unexpected unlimited views value after round trip")
		}
		if !decryptedSecret.Expires.Equal(secret.Expires) {
			t.Fatalf("unexpected expiration after round trip, got %s want %s", decryptedSecret.Expires, secret.Expires)
		}
		if len(decryptedSecret.Files) != len(secret.Files) {
			t.Fatalf("unexpected file count after round trip, got %d want %d", len(decryptedSecret.Files), len(secret.Files))
		}
		if !bytes.Equal(decryptedSecret.Files["fuzz.bin"], secret.Files["fuzz.bin"]) {
			t.Fatalf("unexpected file content after round trip")
		}
	})
}
