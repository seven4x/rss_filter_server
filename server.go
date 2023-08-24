package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
)

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/ping", handlePing)
	router.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
}

func handleRequest(w http.ResponseWriter, r *http.Request) {

	filterTitle := r.URL.Query().Get("filter_title")
	keywords := strings.Split(filterTitle, "|")

	url := r.URL.Query().Get("url")
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	//todo get by http request and response all headers to client
	if err != nil {
		http.Error(w, "Failed to parse RSS feed", http.StatusInternalServerError)
		return
	}

	output := &feeds.Feed{
		Title:       feed.Title,
		Link:        &feeds.Link{Href: feed.Link},
		Description: feed.Description,
		Created:     time.Now(),
	}

	filteredItems := make([]*feeds.Item, 0)

	for _, item := range feed.Items {
		for _, keyword := range keywords {
			if strings.Contains(item.Title, keyword) {
				f := &feeds.Item{
					Title:       item.Title,
					Link:        &feeds.Link{Href: item.Link},
					Description: item.Description,
				}
				if item.UpdatedParsed != nil {
					f.Created = *item.UpdatedParsed
				}
				filteredItems = append(filteredItems, f)
				break
			}
		}
	}
	output.Items = filteredItems
	atom, err := output.ToAtom()
	if err != nil {
		http.Error(w, "Failed to generate filtered RSS content", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/atom+xml;charset=UTF-8")
	fmt.Fprint(w, atom)
}
