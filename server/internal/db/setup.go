package setup

import (
	"database/sql"
	"embed"
	"net/http"
	"tano/http/routes"
	"tano/internal/auth"
	db "tano/internal/db/generated"
	"tano/internal/upload"

	_ "modernc.org/sqlite"

	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/golang-migrate/migrate/v4"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func SetupDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	cleanup := func(err error) (*sql.DB, error) {
		db.Close()
		return nil, err
	}

	sourceDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return cleanup(err)
	}

	dbDriver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return cleanup(err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", dbDriver)
	if err != nil {
		return cleanup(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		m.Close()
		return cleanup(err)
	}

	appdb, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	sourceError, databaseError := m.Close()

	if sourceError != nil {
		appdb.Close()
		return cleanup(sourceError)
	}
	if databaseError != nil {
		appdb.Close()
		return cleanup(databaseError)
	}

	return appdb, err
}

func SetupApp(conn *sql.DB) *http.ServeMux {
	dbconn := db.New(conn)
	authController := auth.NewAuthController(dbconn)
	uploadController := upload.NewUploadController(dbconn, conn)

	authMux := routes.NewAuthMux(authController)
	uploadMux := routes.NewUploadMux(uploadController)

	mux := http.NewServeMux()
	mux.Handle("/user/", http.StripPrefix("/user", authMux))
	mux.Handle("/upload/", http.StripPrefix("/upload", uploadMux))
	return mux
}
