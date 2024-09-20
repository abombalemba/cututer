package main

import (
	"log"
	"net/http"
	"database/sql"
	"time"
	"encoding/json"
	"math/rand"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
	mu sync.Mutex
)

const (
	lengthShortUrl = 10
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
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
	//http.HandleFunc("/c/", cUrlHandler)
	http.HandleFunc("/shorten", shortenUrlHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func generateShortUrl(originalUrl string) string {
	str := generateRandomString()

	checkUrlSQL := `
	SELECT COUNT(*) FROM urls WHERE short_url == ?
	`

	res, err := db.Exec(checkUrlSQL, str)
	if err != nil {
		log.Fatalf("error or this short is existing", res, err)
	}

	return "http://85.193.81.143/" + str
}

func generateRandomString() string {
	rand.Seed(time.Now().UnixNano())
	str := make([]byte, lengthShortUrl)

	for i := range str {
		str[i] = letters[rand.Intn(len(letters))]
	}

	return string(str)
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

	if err := r.ParseForm(); err != nil {
		log.Fatalf("shortenUrlHandler failed because there was a parsing error", err)
		http.Error(w, "shortenUrlHandler failed because there was a parsing error", http.StatusBadRequest)
		return
	}

	req.OriginalUrl = r.FormValue("original_url")

	if req.OriginalUrl == "" {
		log.Fatalf("shortenUrlHandler failed because req.OriginalUrl = \"\"")
		http.Error(w, "shortenUrlHandler failed because req.OriginalUrl = \"\"", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	shortUrl := generateShortUrl(req.OriginalUrl)

	_, err := db.Exec(
		"INSERT INTO urls (original_url, short_url) VALUES (?, ?)", req.OriginalUrl, shortUrl)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := UrlResponse{ShortUrl: shortUrl}
	w.Header().Set("Content-Type", "x-www-form-urlencoded")
	json.NewEncoder(w).Encode(response)

	log.Println("shortenUrlHandler successfully executed")
}
