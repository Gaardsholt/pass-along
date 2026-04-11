package types

import "testing"

func TestGenerateAndParseTokenAsExpected(t *testing.T) {
	lookupID, accessKey, token, err := GenerateToken()
	if err != nil {
		t.Fatalf("expected token generation to succeed: %v", err)
	}

	if lookupID == "" || accessKey == "" || token == "" {
		t.Fatalf("expected generated token values to be non-empty")
	}

	gotLookupID, gotAccessKey, err := ParseToken(token)
	if err != nil {
		t.Fatalf("expected token parsing to succeed: %v", err)
	}

	if gotLookupID != lookupID {
		t.Fatalf("unexpected lookup id, got %s want %s", gotLookupID, lookupID)
	}

	if gotAccessKey != accessKey {
		t.Fatalf("unexpected access key, got %s want %s", gotAccessKey, accessKey)
	}
}

func TestParseTokenInvalid(t *testing.T) {
	if _, _, err := ParseToken("invalid-token"); err == nil {
		t.Fatalf("expected invalid token to fail parsing")
	}
}
