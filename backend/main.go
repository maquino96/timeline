package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/maquino96/timeline/internal/adapters"
	"github.com/maquino96/timeline/internal/api"
	"github.com/maquino96/timeline/internal/models"
	"github.com/maquino96/timeline/internal/scheduler"
	"github.com/maquino96/timeline/internal/store"
	"github.com/maquino96/timeline/internal/topic"
)

func main() {
	teal := "\033[38;5;43m"
	reset := "\033[0m"

	fmt.Println("")
	fmt.Printf("   %s████████╗ ██╗ ███╗   ███╗ ███████╗ ██╗     ██╗ ███╗   ██╗ ███████╗%s\n", teal, reset)
	fmt.Printf("   %s╚══██╔══╝ ██║ ████╗ ████║ ██╔════╝ ██║     ██║ ████╗  ██║ ██╔════╝%s\n", teal, reset)
	fmt.Printf("   %s   ██║    ██║ ██╔████╔██║ █████╗   ██║     ██║ ██╔██╗ ██║ █████╗  %s\n", teal, reset)
	fmt.Printf("   %s   ██║    ██║ ██║╚██╔╝██║ ██╔══╝   ██║     ██║ ██║╚██╗██║ ██╔══╝  %s\n", teal, reset)
	fmt.Printf("   %s   ██║    ██║ ██║ ╚═╝ ██║ ███████╗ ███████╗██║ ██║ ╚████║ ███████╗%s\n", teal, reset)
	fmt.Printf("   %s   ╚═╝    ╚═╝ ╚═╝     ╚═╝ ╚══════╝ ╚══════╝╚═╝ ╚═╝  ╚═══╝ ╚══════╝%s\n", teal, reset)
	fmt.Printf("              %spersonal chronological feed aggregator%s\n", teal, reset)
	fmt.Println("")

	dbPath := os.Getenv("TIMELINE_DB_PATH")
	if dbPath == "" {
		dbPath = "timeline.db"
	}

	store, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer store.Close()

	seedSources(store)

	registry := adapters.NewRegistry()
	engine := topic.New()
	sch := scheduler.New(store, registry, engine)
	sch.Start()
	defer sch.Stop()

	sources, _ := store.GetEnabledSources()
	for _, src := range sources {
		if src.Type == models.SourceSECEDGAR {
			log.Printf("scheduler: SEC EDGAR source %q (CIK=%s) polling every %ds", src.Name, src.URL, src.Interval)
		}
	}
	log.Printf("scheduler: %d total enabled sources", len(sources))

	apiHandler := api.New(store, sch)

	mux := http.NewServeMux()
	apiHandler.RegisterRoutes(mux)

	fs := http.FileServer(http.Dir("../frontend/out"))
	mux.Handle("GET /", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("timeline server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, withCORS(mux)); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Expose-Headers", "X-Total-Count")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func seedSources(s *store.Store) {
	sources, err := s.GetSources()
	if err != nil || len(sources) > 0 {
		return
	}

	defaults := []models.Source{
		{Type: models.SourceHackerNews, Name: "Hacker News", URL: "topstories", Interval: 300, Enabled: true},
		{Type: models.SourceRSS, Name: "r/dev", URL: "https://www.reddit.com/r/programming+ExperiencedDevs+SoftwareEngineering/new/.rss", Interval: 300, Enabled: true},
		{Type: models.SourceRSS, Name: "r/startups", URL: "https://www.reddit.com/r/startups+ycombinator/new/.rss", Interval: 300, Enabled: true},
		{Type: models.SourceRSS, Name: "r/tech", URL: "https://www.reddit.com/r/selfhosted+homelab+deepseek/new/.rss", Interval: 300, Enabled: true},
		{Type: models.SourceRSS, Name: "r/AI", URL: "https://www.reddit.com/r/Anthropic+OpenAI/new/.rss", Interval: 300, Enabled: true},
		{Type: models.SourceRSS, Name: "r/lifestyle", URL: "https://www.reddit.com/r/digitalnomad+buildapcsales/new/.rss", Interval: 300, Enabled: true},
		{Type: models.SourceSECEDGAR, Name: "FDGRX - Fidelity Growth Company Fund (316200104)", URL: "0000707823", Interval: 21600, Enabled: true},
	}

	for i := range defaults {
		s.AddSource(&defaults[i])
	}
}
