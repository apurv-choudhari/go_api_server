package main

import (
	"database/sql"
	"go-service/api_server"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var dbConn *sql.DB

func main() {
	router := mux.NewRouter()

	var err error
	dbConn, err = api_server.ConnectDB(sql.Open)
	if err != nil {
		log.Fatal("Database connection failed: ", err)
		return
	}
	defer dbConn.Close()

	router.HandleFunc("/scan", func(w http.ResponseWriter, r *http.Request) {
		api_server.ScanRepoHandler(dbConn, w, r)
	}).Methods("POST")
	router.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		api_server.QueryVulnHandler(dbConn, w, r)
	}).Methods("POST")

	log.Println("Serverrrr starting on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", router))
}
