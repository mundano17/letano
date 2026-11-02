package main

import (
	"database/sql"
	"log"
	"net/http"
	"tano/http/middleware"
	"tano/http/routes"
	"tano/internal/auth"
	db "tano/internal/db/generated"

	_ "modernc.org/sqlite"
)

func main() {
	conn, err := sql.Open("sqlite", "./tano_server.db")
	if err != nil {
		panic(err)
	}

	dbconn := db.New(conn)
	authController := auth.NewAuthController(dbconn)
	authMux := routes.NewAuthMux(authController)

	mux := http.NewServeMux()
	mux.Handle("/user/", http.StripPrefix("/user", authMux))

	if err := http.ListenAndServe(":8080", middleware.LoggingMiddleware((middleware.CORSMiddleware(mux)))); err != nil {
		log.Fatal(err)
	}
}
