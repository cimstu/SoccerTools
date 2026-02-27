package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/soccertools/soccertools/internal/crawler"
	"github.com/soccertools/soccertools/internal/store"
)

//go:embed static
var staticFS embed.FS

const (
	addr        = ":3000"
	crawlPeriod = 5 * time.Second
	defaultDays = 7
	minDays     = 1
	maxDays     = 30
)

func main() {
	s := store.New()

	// 启动时先爬一次
	go runCrawl(s)

	// 定时爬取
	ticker := time.NewTicker(crawlPeriod)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			runCrawl(s)
		}
	}()

	http.HandleFunc("/health", methodGET(health))
	http.HandleFunc("/replays", methodGET(replays(s)))
	http.HandleFunc("/replays/refresh", methodPOST(replaysRefresh(s)))
	// Web 页面：查询与刷新巴萨录像
	staticRoot, _ := fs.Sub(staticFS, "static")
	http.Handle("/", http.FileServer(http.FS(staticRoot)))

	log.Printf("listen %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func runCrawl(s *store.Store) {
	items, err := crawler.FetchAndParse()
	if err != nil {
		log.Printf("crawl: %v", err)
		return
	}
	s.Add(items)
	log.Printf("crawl: got %d replays", len(items))
}

func methodGET(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}

func methodPOST(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func replaysRefresh(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runCrawl(s)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "message": "已刷新"})
	}
}

func replays(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		days := defaultDays
		if d := r.URL.Query().Get("days"); d != "" {
			n, err := strconv.Atoi(d)
			if err != nil || n < minDays || n > maxDays {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"error": "days must be 1-30"})
				return
			}
			days = n
		}
		items := s.GetLastNDays(days)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"items": items})
	}
}
