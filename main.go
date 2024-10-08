package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

// all the business logic except make server up will move to specific packages

// Structure to store real and bot URLs for each ID
type Link struct {
	Real string `json:"real"`
	Bot  string `json:"bot"`
}

// In-memory map to store ID -> Link mappings Basic For Now. We can customize with our requirements later. we will be using db for storing this link map
var linkMap = make(map[string]Link)
var mapMutex = &sync.RWMutex{}

func isBot(userAgent string) bool {
	botIdentifiers := []string{
		"bot", "crawl", "spider", "slurp", "facebook", "google", "bing", "yahoo",
		"duckduckgo", "baidu", "yandex", "sogou", "exabot", "ia_archiver", "twitterbot",
		"telegrambot", "whatsapp", "mediapartners", "applebot", "embedly", "quora",
		"pinterest", "redditbot", "slackbot", "vkshare", "w3c_validator", "wget", "curl",
		"java", "libwww-perl", "python-requests", "httpclient", "aiohttp", "okhttp", "scrapy",
		"php", "go-http-client", "ruby", "node-fetch",
	}
	for _, identifier := range botIdentifiers {
		if strings.Contains(strings.ToLower(userAgent), identifier) {
			return true
		}
	}
	return false
}

// Cloaker handler to handle requests based on ID
func cloakedHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	mapMutex.RLock()
	link, exists := linkMap[id]
	mapMutex.RUnlock()

	if !exists {
		http.Error(w, "ID not found", http.StatusNotFound)
		return
	}

	userAgent := r.Header.Get("User-Agent")

	// Redirect based on bot or not
	if isBot(userAgent) {
		http.Redirect(w, r, link.Bot, http.StatusFound)
	} else {
		// Check for JavaScript support
		if _, ok := r.URL.Query()["js"]; !ok {
			// Serve HTML with JavaScript to check if JavaScript is enabled
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
				<html>
					<head>
						<title>Checking...</title>
						<script>
							window.location.href = "` + r.URL.Path + `?js=enabled";
						</script>
					</head>
					<body>
						Redirecting...
					</body>
				</html>
			`))
			return
		}
		http.Redirect(w, r, link.Real, http.StatusFound)
	}
}

// Handle the POST request to update the link mappings which will save in DB later
func updateLinkHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var link Link
	err := json.NewDecoder(r.Body).Decode(&link)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	mapMutex.Lock()
	linkMap[id] = link
	mapMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Links updated successfully"))
}

func main() {
	r := chi.NewRouter()
	r.Use(loggingMiddleware)

	r.Get("/{id}", cloakedHandler)

	r.Post("/update/{id}", updateLinkHandler)

	log.Println("Server started on port 8080")

	http.ListenAndServe(":8080", r)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Custom ResponseWriter to capture status code
		rr := &responseRecorder{w, http.StatusOK}

		// Call the next handler
		next.ServeHTTP(rr, r)

		// Log request details and response status code
		log.Printf("%s %s %s %d %s", r.Method, r.RequestURI, r.Proto, rr.statusCode, time.Since(start))
	})
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}
