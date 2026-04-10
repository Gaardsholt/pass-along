package api

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/Gaardsholt/pass-along/config"
	"github.com/Gaardsholt/pass-along/types"
)

func setupTestConfig() {
	validFor := []int{60, 3600}
	config.Config.ValidForOptions = validFor
	config.Config.MaxSecretBytes = 1024
	config.Config.MaxMultipartBytes = 10 * 1024
	config.Config.MaxFiles = 2
	config.Config.MaxFileSizeBytes = 1024
	config.Config.MaxFilenameLength = 64
}

func TestValidateEntryRejectsInvalidExpiration(t *testing.T) {
	setupTestConfig()
	entry := types.Entry{
		Content:   "hello",
		ExpiresIn: 10,
	}

	if err := validateEntry(entry); err == nil {
		t.Fatalf("expected validateEntry to reject invalid expires_in")
	}
}

func TestGetFormDataMissingDataReturnsError(t *testing.T) {
	setupTestConfig()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/api", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	entry := types.Entry{}
	if err := getFormData(rec, req, &entry); err == nil {
		t.Fatalf("expected getFormData to fail when data payload is missing")
	}
}
