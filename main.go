package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

type Feed struct {
	Channel struct {
		Title		string	`xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Language    string    `xml:"language"`
		Item	[]FeedItem		`xml:"item"`
	} `xml:"channel"`
}

type FeedItem struct {
	Title		string	`xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchData(url string) (Feed, error) {
	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return Feed{}, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Feed{}, err
	}

	feed := Feed{}

	err = xml.Unmarshal(data, &feed)
	if err != nil {
		return Feed{}, err
	}
	return feed, nil
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal JSON response: %v", payload)
		w.WriteHeader(500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func (feed *Feed) handleFeeds(w http.ResponseWriter, r *http.Request) {
	type Message struct {
		Greeting string `json:"greeting"`
		Response Feed `json:"response"`
	}

	resp := Message {
		Greeting: "Hello there",
		Response: *feed,
	}

	respondWithJson(w, 200, resp)
}

func main() {
	godotenv.Load(".env")

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT is not found in the environment")
	}

	url := os.Args[1]

	fmt.Printf("Fetching data from: %v", url)

	feed, err := fetchData(url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(feed)

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		AllowCredentials: false,
		ExposedHeaders: []string{"Link"},
		MaxAge: 300,
	}))

	router.Get("/feeds", feed.handleFeeds)

	srv := &http.Server{
		Handler: router,
		Addr:	":" + port,
	}

	log.Printf("Server starting on port %v", port)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Port:", port)
}