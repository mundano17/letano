package upload

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func getHash(content []byte) string {
	expected := sha256.Sum256(content)
	return hex.EncodeToString(expected[:])
}

func TestVerifyFileHash(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectedHash string
		want         bool
	}{
		{
			name:         "validHash",
			content:      "helloWorld",
			expectedHash: getHash([]byte("helloWorld")),
			want:         true,
		},
		{
			name:         "invalidHash",
			content:      "helloWorld",
			expectedHash: getHash([]byte("helloworld")),
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			content := []byte(tt.content)
			path := filepath.Join(dir, "test.txt")

			if err := os.WriteFile(path, content, 0644); err != nil {
				t.Fatal(err)
			}

			got := verifyFileHash(path, tt.expectedHash)
			if got != tt.want {
				t.Errorf("verifyFileHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
