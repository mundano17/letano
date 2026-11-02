package integration_test

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"tano/http/routes"
	"tano/internal/auth"
	migrate "tano/internal/db"
	db "tano/internal/db/generated"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestApp(conn *sql.DB) http.Handler {

	dbconn := db.New(conn)
	authController := auth.NewAuthController(dbconn)
	authMux := routes.NewAuthMux(authController)

	mux := http.NewServeMux()
	mux.Handle("/user/", http.StripPrefix("/user", authMux))

	return mux
}

func registerUser(body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/user/register",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	return req, rec
}

func TestRegisterUserController(t *testing.T) {

	tests := []struct {
		name       string
		email      string
		password   string
		setup      bool
		wantStatus int
	}{
		{
			name:       "valid registration",
			email:      "test@example.com",
			password:   "VeryStrongPassword@1234",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "duplicate registration",
			email:      "test@example.com",
			password:   "VeryStrongPassword@1234",
			setup:      true,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "weak password",
			email:      "test@example.com",
			password:   "weakpassword",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		path := filepath.Join(t.TempDir(), t.Name())
		conn, err := migrate.SetupDB(path)
		if err != nil {
			t.Fatal(err)
		}
		app := setupTestApp(conn)
		body := fmt.Sprintf("{\"email\": \"%s\", \"password\": \"%s\"}", tt.email, tt.password)
		if tt.setup {
			req, rec := registerUser(body)
			app.ServeHTTP(rec, req)
			if rec.Result().StatusCode != http.StatusCreated {
				t.Fatal("setup registration failed")
			}
		}

		req, rec := registerUser(body)
		app.ServeHTTP(rec, req)
		if rec.Result().StatusCode != tt.wantStatus {
			t.Fatalf("Expected %d, got %d", tt.wantStatus, rec.Result().StatusCode)
		}

		t.Cleanup(func() {
			conn.Close()
		})
	}

}

func loginUser(body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/user/login",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	return req, rec
}

func TestLoginUserController(t *testing.T) {

	tests := []struct {
		name             string
		email            string
		registerPassword string
		loginPassword    string
		notRegister      bool
		wantStatus       int
	}{
		{
			name:             "valid login",
			email:            "test@example.com",
			registerPassword: "VeryStrongPassword@1234",
			loginPassword:    "VeryStrongPassword@1234",
			wantStatus:       http.StatusOK,
		},
		{
			name:             "invalid credentials",
			email:            "test@example.com",
			registerPassword: "VeryStrongPassword@1234",
			loginPassword:    "VeryStrongPassword@123",
			wantStatus:       http.StatusUnauthorized,
		},
		{
			name:             "not registered",
			email:            "test@example.com",
			registerPassword: "weakpassword",
			notRegister:      true,
			wantStatus:       http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		path := filepath.Join(t.TempDir(), t.Name())
		conn, err := migrate.SetupDB(path)
		if err != nil {
			t.Fatal(err)
		}

		app := setupTestApp(conn)

		if !tt.notRegister {
			body := fmt.Sprintf("{\"email\": \"%s\", \"password\": \"%s\"}", tt.email, tt.registerPassword)
			req, rec := registerUser(body)
			app.ServeHTTP(rec, req)
			if rec.Result().StatusCode != http.StatusCreated {
				t.Fatal("setup registration failed")
			}
		}

		body := fmt.Sprintf("{\"email\": \"%s\", \"password\": \"%s\"}", tt.email, tt.loginPassword)
		req, rec := loginUser(body)
		app.ServeHTTP(rec, req)
		if rec.Result().StatusCode != tt.wantStatus {
			t.Fatalf("Expected %d, got %d", tt.wantStatus, rec.Result().StatusCode)
		}

		t.Cleanup(func() {
			conn.Close()
		})
	}
}
