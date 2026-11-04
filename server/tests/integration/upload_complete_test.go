package integration_test

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"tano/internal/upload"
	"testing"
)

func generateData(t *testing.T, testFileSize int64) []byte {
	t.Helper()
	data := make([]byte, testFileSize)
	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}

	return data
}

func completeUpload(t *testing.T, server *httptest.Server, client *http.Client, data []byte, want int, fakeFileHash bool) string {
	hash := sha256.Sum256(data)
	fileHash := hex.EncodeToString(hash[:])
	uploadInitReq := upload.NewUploadRequest{
		Filehash: fileHash,
		Filename: "test.txt",
		Filesize: int64(len(data)),
		Mimetype: "text/plain",
		Fields: []upload.Field{
			{
				FieldKey:   "category",
				FieldValue: "documents",
			}},
	}

	if fakeFileHash {
		uploadInitReq.Filehash = initReq.Filehash
	}

	res := initUpload(t, client, server.URL, jsonBody(t, uploadInitReq), http.StatusCreated)
	var uploadInitRes upload.NewUploadRes
	err := json.Unmarshal(res, &uploadInitRes)
	if err != nil {
		t.Fatal("Error unmarshalling uploadInitRes")
	}

	for i := int64(0); i < int64(len(data)); i += 10 * 1024 * 1024 {
		end := min(i+10*1024*1024, int64(len(data)))
		chunkReq := upload.ChunkRequest{
			Start:    i,
			Length:   end - i,
			Data:     data[i:end],
			FileSize: int64(len(data)),
			UploadID: uploadInitRes.UploadID,
		}
		uploadChunk(t, client, server, jsonBody(t, chunkReq), http.StatusOK)
	}

	finishUpload(t, client, server.URL, uploadInitRes.UploadID, want)
	return uploadInitRes.UploadID
}

func finishUpload(t *testing.T, client *http.Client, url string, uploadID string, want int) {
	res, err := client.Post(
		url+"/upload/finish/"+uploadID,
		"",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := json.Marshal(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	assertStatus(t, res, want, resp)
}

func TestFinishUploadController(t *testing.T) {
	tests := []struct {
		name         string
		resendFile   bool
		wantStatus   int
		fileSize     int64
		resend       bool
		fakeFileHash bool
	}{
		{
			name:       "valid single upload",
			resendFile: false,
			wantStatus: http.StatusOK,
			fileSize:   49 * 1024 * 1024,
		},
		{
			name:       "valid double upload",
			resendFile: true,
			wantStatus: http.StatusOK,
			fileSize:   49 * 1024 * 1024,
		},
		{
			name:       " re-finish upload",
			resendFile: false,
			resend:     true,
			wantStatus: http.StatusBadRequest,
			fileSize:   49 * 1024 * 1024,
		},
		{
			name:         "invalid fileHash",
			wantStatus:   http.StatusBadRequest,
			fileSize:     10 * 1024 * 1024,
			fakeFileHash: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, client, conn := setupClient(t)
			data := generateData(t, tt.fileSize)
			want := tt.wantStatus
			if tt.resend {
				want = http.StatusOK
			}

			uploadID := completeUpload(t, server, client, data, want, tt.fakeFileHash)

			if tt.resendFile {
				completeUpload(t, server, client, data, tt.wantStatus, tt.fakeFileHash)
			}

			if tt.resend {
				finishUpload(t, client, server.URL, uploadID, tt.wantStatus)
			}

			if !tt.fakeFileHash {
				assertStorageContents(t, "data/storage", 1)
			} else {
				assertStorageContents(t, "data/storage", 0)
			}
			t.Cleanup(func() {
				server.Close()
				conn.Close()
				if err := os.RemoveAll("data"); err != nil {
					t.Errorf("failed to remove test data directory: %v", err)
				}

				if err := os.RemoveAll("tmp"); err != nil {
					t.Errorf("failed to remove temporary upload directory: %v", err)
				}
			})

		})
	}

}

func assertStorageContents(t *testing.T, dir string, want int) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil && want != 0 {
		t.Fatalf("failed to read storage directory: %v", err)
	}

	if len(entries) != want {
		t.Fatalf("expected %d objects, got %d", want, len(entries))
	}
}
