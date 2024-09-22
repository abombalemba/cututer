package main

import (
	"log"
	"net/http"
	"database/sql"
	"time"
	"encoding/json"
	"math/rand"
	"sync"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
	mu sync.Mutex
)

const (
	protocol = "http://"
	host = "localhost"
	port = "8080"
	path = "/c/"

	lengthShortUrl = 3
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
	http.HandleFunc("/api", apiUrlHandler)
	http.HandleFunc("/c/", cUrlHandler)

	if err := http.ListenAndServe(":" + port, nil); err != nil {
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

	return str
}

func generateRandomString() string {
	rand.Seed(time.Now().UnixNano())
	str := make([]byte, lengthShortUrl)

	for i := range str {
		str[i] = letters[rand.Intn(len(letters))]
	}

	return string(str)
}

func checkGeneratedShortUrl(shortUrl string) {

}

func initDB() {
	var err error

	db, err = sql.Open("sqlite3", "../database/urls.db")
	if err != nil {
		log.Fatal(err)
		return
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
		return
	}

	log.Println("initDB successfully executed")
}

func indexUrlHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../web/index.html")

	log.Println("indexUrlHandler successfully executed")
}

func apiUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Fatalf("apiUrlHandler failed because another method not allowed", r.Method)
		http.Error(w, "apiUrlHandler failed because another method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UrlRequest

	if err := r.ParseForm(); err != nil {
		log.Fatalf("apiUrlHandler failed because there was a parsing error", err)
		http.Error(w, "apiUrlHandler failed because there was a parsing error", http.StatusBadRequest)
		return
	}

	req.OriginalUrl = r.FormValue("original_url")

	if req.OriginalUrl == "" {
		log.Fatalf("apiUrlHandler failed because req.OriginalUrl = \"\"")
		http.Error(w, "apiUrlHandler failed because req.OriginalUrl = \"\"", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	is, err := originalUrlInDB(req.OriginalUrl)
	if err != nil {
		log.Fatalf("apiUrlHandler failed because SQL query got error 142", err)
		http.Error(w, "apiUrlHandler failed because SQL query got error 142", http.StatusBadRequest)
		return
	}

	var shortUrl string

	if is == true {
		row := db.QueryRow("SELECT short_url FROM urls WHERE original_url == ?", req.OriginalUrl)

		err := row.Scan(&shortUrl)

		if err != nil {
			log.Fatalf("apiUrlHandler failed because SQL query got error 155")
			http.Error(w, "apiUrlHandler failed because SQL query got error 155", http.StatusBadRequest)
			return
		}
	} else {
		shortUrl = generateShortUrl(req.OriginalUrl)

		_, err := db.Exec(
			"INSERT INTO urls (original_url, short_url) VALUES (?, ?)", req.OriginalUrl, shortUrl)

		if err != nil {
			log.Fatalf("apiUrlHandler failed because SQL query got error 164", err)
			http.Error(w, "apiUrlHandler failed because SQL query got error", http.StatusInternalServerError)
			return
		}
	}

	shortUrl = protocol + host + ":" + port + path + shortUrl

	response := UrlResponse{ShortUrl: shortUrl}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Println("apiUrlHandler successfully executed")
}

func cUrlHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/c/")

	row := db.QueryRow(
		"SELECT original_url FROM urls WHERE short_url == ? LIMIT 1", path)

	var originalUrl string

	err := row.Scan(&originalUrl)

	if err == nil {
		http.Redirect(w, r, originalUrl, http.StatusFound)

		log.Println("cUrlHandler successfully executed")
	} else if err == sql.ErrNoRows {
		http.ServeFile(w, r, "../web/notfound.html")

		log.Println("cUrlHandler not found url")
	} else {
		http.Error(w, "cUrlHandler failed because SQL query got error", http.StatusBadRequest)

		log.Fatalf("cUrlHandler failed because SQL query got error", err)
	}
}

func originalUrlInDB(originalUrl string) (bool, error) {
	row := db.QueryRow(
		"SELECT id FROM urls WHERE original_url == ?", originalUrl)

	var id int

	err := row.Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	if id == 0 {
		return false, nil
	}
	return true, nil
}
