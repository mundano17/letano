package integration_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	setup "tano/internal/db"
	"tano/internal/upload"
	"testing"
)

const data string = "test"

var initReq upload.NewUploadRequest = upload.NewUploadRequest{
	Filehash: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
	Filename: "test.txt",
	Filesize: int64(len([]byte(data))),
	Mimetype: "text/plain",
	Fields: []upload.Field{
		{
			FieldKey:   "category",
			FieldValue: "documents",
		}},
}

func setupClient(t *testing.T) (*httptest.Server, *http.Client, *sql.DB) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := setup.SetupDB(path)
	if err != nil {
		t.Fatal(err)
	}

	app := setupTestApp(conn)
	server := httptest.NewServer(app)
	client := newTestClient(t)

	body := fmt.Sprintf("{\"email\": \"%s\", \"password\": \"%s\"}", EMAIL, PASSWORD)
	resp := registerAndLogin(t, client, server, body)

	defer resp.Body.Close()

	return server, client, conn
}

func uploadChunk(t *testing.T, client *http.Client, server *httptest.Server, JSONBody *bytes.Reader, want int) {
	t.Helper()
	res, err := client.Post(
		server.URL+"/upload/add",
		"application/json",
		JSONBody,
	)
	if err != nil {
		t.Fatal(err)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err.Error())

	}

	assertStatus(t, res, want, resBody)
}

func TestUploadChunkController(t *testing.T) {
	tests := []struct {
		name   string
		modify func(u *upload.ChunkRequest)
		want   int
	}{
		{
			name:   "valid chunk",
			modify: func(c *upload.ChunkRequest) {},
			want:   http.StatusOK,
		},
		{
			name: "unaligned offset",
			modify: func(c *upload.ChunkRequest) {
				c.Start = 1
			},
			want: http.StatusBadRequest,
		},
		{
			name: "zero length",
			modify: func(c *upload.ChunkRequest) {
				c.Length = 0
			},
			want: http.StatusBadRequest,
		},
		{
			name: "incorrect length",
			modify: func(c *upload.ChunkRequest) {
				c.Length = 5
			},
			want: http.StatusBadRequest,
		},
		{
			name: "data length mismatch",
			modify: func(c *upload.ChunkRequest) {
				c.Data = []byte("short")
			},
			want: http.StatusBadRequest,
		}}

	server, client, conn := setupClient(t)

	respBody := initUpload(t, client, server.URL, jsonBody(t, initReq), http.StatusCreated)

	var uploadInitRes upload.NewUploadRes
	err := json.Unmarshal(respBody, &uploadInitRes)
	if err != nil {
		t.Fatal("Error unmarshalling uploadInitRes")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := upload.ChunkRequest{
				UploadID: uploadInitRes.UploadID,
				Start:    0,
				Length:   int64(len([]byte(data))),
				Data:     []byte(data),
				FileSize: int64(len([]byte(data))),
			}
			tt.modify(&chunk)
			uploadChunk(t, client, server, jsonBody(t, chunk), tt.want)
		})
	}

	t.Cleanup(func() {
		server.Close()
		conn.Close()
		if err := os.RemoveAll("tmp"); err != nil {
			t.Errorf("failed to remove temporary upload directory: %v", err)
		}
	})
}
