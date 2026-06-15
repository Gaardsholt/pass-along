package crypto

import "testing"

func FuzzDecryptRejectsInvalidCiphertext(f *testing.F) {
	f.Add([]byte{}, "")
	f.Add([]byte("short"), "encryptionkey")
	f.Add([]byte("this is not valid ciphertext"), "another-key")

	f.Fuzz(func(t *testing.T, data []byte, encryptionKey string) {
		if len(data) > 4096 || len(encryptionKey) > 256 {
			t.Skip()
		}

		_, _ = Decrypt(data, encryptionKey)
	})
}

func FuzzEncryptDecryptRoundTrip(f *testing.F) {
	f.Add("mysupersecretvalue", "encryptionkey")
	f.Add("", "")
	f.Add("content with symbols !@#$%^&*()", "different-key")

	f.Fuzz(func(t *testing.T, plaintext string, encryptionKey string) {
		if len(plaintext) > 1024 || len(encryptionKey) > 256 {
			t.Skip()
		}

		expectedBytes, err := getBytes(plaintext)
		if err != nil {
			t.Fatalf("expected gob encoding to succeed: %v", err)
		}

		encryptedData, err := Encrypt(plaintext, encryptionKey)
		if err != nil {
			t.Fatalf("expected encryption to succeed: %v", err)
		}

		decryptedData, err := Decrypt(encryptedData, encryptionKey)
		if err != nil {
			t.Fatalf("expected decryption to succeed: %v", err)
		}

		if string(decryptedData) != string(expectedBytes) {
			t.Fatalf("unexpected decrypted data after round trip")
		}
	})
}
