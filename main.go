package main

import (
	"log"
	"net/http"
	"database/sql"
	"encoding/json"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
	mu sync.Mutex
)

type UrlRequest struct {
	OriginalUrl string `json:"original_url"`
}

type UrlResponse struct {
	ShortUrl string `json:"short_url"`
}

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/", indexUrlHandler)
	http.HandleFunc("/c/", cUrlHandler)
	http.HandleFunc("/shorten", shortenUrlHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func generateShortUrl(originalUrl string) string {
	return "http://85.193.81.143/" + originalUrl[len(originalUrl)-5:]
}

func initDB() {
	var err error

	db, err = sql.Open("sqlite3", "./urls.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		original_url TEXT NOT NULL,
		short_url TEXT NOT NULL UNIQUE
	);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func indexUrlHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func cUrlHandler(w http.ResponseWriter, r *http.Request) {
	var req UrlRequest

	log.Println("45 ok")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "71", http.StatusBadRequest)
		return
	}

	req.OriginalUrl = r.FormValue("original_url")

	if req.OriginalUrl == "" {
		http.Error(w, "95", http.StatusBadRequest)
		return
	}
}

func shortenUrlHandler(w http.ResponseWriter, r *http.Request) {
	var req UrlRequest

	log.Println("12 ok")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "74", http.StatusBadRequest)
		return
	}

	req.OriginalUrl = r.FormValue("original_url")

	log.Println(req.OriginalUrl)

	if req.OriginalUrl == "" {
		http.Error(w, "81", http.StatusBadRequest)
		return
	}

	log.Println("123 ok")

	mu.Lock()
	defer mu.Unlock()

	shortUrl := generateShortUrl(req.OriginalUrl)

	log.Println("456 ok")

	_, err := db.Exec(
		"INSERT INTO urls (original_url, short_url) VALUES (?, ?)",
		req.OriginalUrl, shortUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("789 ok")

	response := UrlResponse{ShortUrl: shortUrl}
	w.Header().Set("Content-Type", "x-www-form-urlencoded")
	json.NewEncoder(w).Encode(response)

	log.Println("1234 ok")
}
