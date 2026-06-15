package api

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Gaardsholt/pass-along/config"
	"github.com/Gaardsholt/pass-along/types"
)

func FuzzValidateEntry(f *testing.F) {
	f.Add("hello", 60, "file.txt", []byte("contents"), false, 1)
	f.Add("", 60, "", []byte{}, true, 0)
	f.Add(strings.Repeat("a", 1024), 3600, "file.txt", []byte("contents"), false, 1)
	f.Add(strings.Repeat("a", 1025), 3600, "file.txt", []byte("contents"), false, 1)
	f.Add("hello", 10, "file.txt", []byte("contents"), false, 1)

	f.Fuzz(func(t *testing.T, content string, expiresIn int, fileName string, fileBytes []byte, unlimitedViews bool, fileCount int) {
		setupTestConfig()
		if len(content) > config.Config.MaxSecretBytes+1 || len(fileBytes) > int(config.Config.MaxFileSizeBytes)+1 {
			t.Skip()
		}

		entry := types.Entry{
			Content:        content,
			ExpiresIn:      expiresIn,
			UnlimitedViews: unlimitedViews,
			Files:          fuzzFiles(fileName, fileBytes, fileCount),
		}

		err := validateEntry(entry)
		if err != nil {
			return
		}

		assertValidEntry(t, entry)
	})
}

func FuzzGetFormData(f *testing.F) {
	f.Add(`{"content":"hello","expires_in":60}`, "file.txt", []byte("contents"), true)
	f.Add(`{"content":"","expires_in":60}`, "file.txt", []byte("contents"), true)
	f.Add(`{"content":"hello","expires_in":10}`, "file.txt", []byte("contents"), true)
	f.Add(`not-json`, "file.txt", []byte("contents"), true)
	f.Add(`{"content":"hello","expires_in":60}`, "file.txt", []byte{}, false)
	f.Add(``, "", []byte{}, true)

	f.Fuzz(func(t *testing.T, dataPayload string, fileName string, fileBytes []byte, includeData bool) {
		setupTestConfig()
		if len(dataPayload) > 2048 || len(fileName) > 256 || len(fileBytes) > int(config.Config.MaxFileSizeBytes)+1 {
			t.Skip()
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		if includeData {
			if err := writer.WriteField("data", dataPayload); err != nil {
				t.Fatalf("expected writing multipart field to succeed: %v", err)
			}
		}
		if fileName != "" || len(fileBytes) > 0 {
			part, err := writer.CreateFormFile("files", fileName)
			if err != nil {
				t.Fatalf("expected creating multipart file to succeed: %v", err)
			}
			if _, err := part.Write(fileBytes); err != nil {
				t.Fatalf("expected writing multipart file to succeed: %v", err)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("expected closing multipart writer to succeed: %v", err)
		}

		req := httptest.NewRequest("POST", "/api", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()
		entry := types.Entry{}

		err := getFormData(rec, req, &entry)
		if err != nil {
			return
		}

		if err := validateEntry(entry); err == nil {
			assertValidEntry(t, entry)
		}
	})
}

func FuzzHumanDuration(f *testing.F) {
	f.Add(-1)
	f.Add(0)
	f.Add(1)
	f.Add(2)
	f.Add(60)
	f.Add(120)
	f.Add(3600)
	f.Add(7200)
	f.Add(86400)
	f.Add(172800)

	f.Fuzz(func(t *testing.T, duration int) {
		got := humanDuration(duration)
		if got == "" {
			t.Fatalf("expected humanDuration to return a non-empty string")
		}

		expectedOutputs := map[int]string{
			1:      "1 second",
			2:      "2 seconds",
			60:     "1 minute",
			120:    "2 minutes",
			3600:   "1 hour",
			7200:   "2 hours",
			86400:  "1 day",
			172800: "2 days",
		}
		if want, ok := expectedOutputs[duration]; ok && got != want {
			t.Fatalf("unexpected duration text, got %q want %q", got, want)
		}
	})
}

func fuzzFiles(fileName string, fileBytes []byte, fileCount int) map[string][]byte {
	if fileCount < 0 {
		fileCount = -fileCount
	}
	if fileCount > config.Config.MaxFiles+1 {
		fileCount = config.Config.MaxFiles + 1
	}
	if fileCount == 0 {
		return nil
	}

	files := map[string][]byte{}
	for i := 0; i < fileCount; i++ {
		name := fileName
		if name == "" {
			name = "file"
		}
		files[fmt.Sprintf("%s-%d", name, i)] = fileBytes
	}
	return files
}

func assertValidEntry(t *testing.T, entry types.Entry) {
	t.Helper()

	if !config.Config.IsValidExpiration(entry.ExpiresIn) {
		t.Fatalf("expected accepted entry to have a valid expiration")
	}
	if len(entry.Content) == 0 && len(entry.Files) == 0 {
		t.Fatalf("expected accepted entry to have content or files")
	}
	if len(entry.Content) > config.Config.MaxSecretBytes {
		t.Fatalf("expected accepted entry content to fit max secret bytes")
	}
	if len(entry.Files) > config.Config.MaxFiles {
		t.Fatalf("expected accepted entry file count to fit max files")
	}
	for _, content := range entry.Files {
		if int64(len(content)) > config.Config.MaxFileSizeBytes {
			t.Fatalf("expected accepted entry file content to fit max file size")
		}
	}
}
