package integration_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"strings"
	"tano/internal/upload"
	"testing"
)

const EMAIL string = "abc@xyz.com"
const PASSWORD string = "VeryStrongPassword@1234"

func newTestClient(t *testing.T) *http.Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	return &http.Client{
		Jar: jar,
	}
}

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return bytes.NewReader(data)
}

func registerAndLogin(t *testing.T, client *http.Client, server *httptest.Server, body string) *http.Response {
	t.Helper()

	resp, err := client.Post(
		server.URL+"/user/register",
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Register: Expected Status:%d ; Recieved Status:%d\n", http.StatusCreated, resp.StatusCode)
	}

	resp, err = client.Post(
		server.URL+"/user/login",
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login: Expected Status:%d ; Recieved Status:%d\n", http.StatusOK, resp.StatusCode)
	}

	defer resp.Body.Close()
	return resp
}

func assertStatus(t *testing.T, res *http.Response, want int, resBody []byte) {
	t.Helper()
	if res.StatusCode != want {
		t.Fatalf(
			"expected status %d, got %d: %s",
			want,
			res.StatusCode,
			resBody,
		)
	}
}

func initUpload(t *testing.T, client *http.Client, url string, body *bytes.Reader, want int) []byte {
	res, err := client.Post(
		url+"/upload/init",
		"application/json",
		body,
	)
	if err != nil {
		t.Fatal(err)
		return nil
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err.Error())
		return nil
	}

	assertStatus(t, res, want, resBody)

	return resBody
}

func TestUploadInitController(t *testing.T) {
	tests := []struct {
		name       string
		initReq    upload.NewUploadRequest
		wantStatus int
	}{
		{
			name: "valid request",
			initReq: upload.NewUploadRequest{
				Filehash: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
				Filename: "test.txt",
				Filesize: 100,
				Mimetype: "text/plain",
				Fields: []upload.Field{
					{
						FieldKey:   "category",
						FieldValue: "documents",
					}},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid request",
			initReq: upload.NewUploadRequest{
				Filename: "",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	server, client, conn := setupClient(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initUpload(t, client, server.URL, jsonBody(t, tt.initReq), tt.wantStatus)
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
