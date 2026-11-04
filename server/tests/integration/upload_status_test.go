package integration_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"tano/internal/upload"
	"testing"
)

func assertUploadStatus(t *testing.T, client *http.Client, server *httptest.Server, uploadID string, status string) {
	res, err := client.Get(server.URL + "/upload/status/" + uploadID)
	if err != nil {
		t.Fatal(err.Error())
	}

	resBody, err := io.ReadAll(res.Body)
	assertStatus(t, res, http.StatusOK, resBody)

	var statusRes upload.UploadStatusRes
	err = json.Unmarshal(resBody, &statusRes)

	if statusRes.UploadStatus != status {
		t.Fatal("The status is not pause")
	}

}

func statusCheck(t *testing.T, client *http.Client, server *httptest.Server, route string, status string) {

	respBody := initUpload(t, client, server.URL, jsonBody(t, initReq), http.StatusCreated)
	var uploadInitRes upload.NewUploadRes
	err := json.Unmarshal(respBody, &uploadInitRes)
	if err != nil {
		t.Fatal("Error unmarshalling uploadInitRes")
	}

	res, err := client.Post(server.URL+"/upload/"+route+"/"+uploadInitRes.UploadID, "", nil)
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err.Error())
	}

	assertStatus(t, res, http.StatusOK, resBody)

	assertUploadStatus(t, client, server, uploadInitRes.UploadID, status)

}

func TestPauseUploadController(t *testing.T) {
	server, client, conn := setupClient(t)

	statusCheck(t, client, server, "pause", string(upload.Paused))

	t.Cleanup(func() {
		server.Close()
		conn.Close()
	})
}

func TestResumeUploadController(t *testing.T) {
	server, client, conn := setupClient(t)

	statusCheck(t, client, server, "resume", string(upload.Resumed))

	t.Cleanup(func() {
		server.Close()
		conn.Close()
	})
}

func TestCancelUploadController(t *testing.T) {
	server, client, conn := setupClient(t)

	statusCheck(t, client, server, "cancel", string(upload.Cancelled))

	t.Cleanup(func() {
		server.Close()
		conn.Close()
	})
}
