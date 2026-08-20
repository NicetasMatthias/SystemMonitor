package server

import (
	"context"
	"encoding/json"
	"html/template"
	"mime"
	"net/http"
	"time"

	"github.com/NicetasMatthias/SystemMonitor/internal/collector"
	"github.com/gorilla/mux"
)

func init() {
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".json", "application/json")
	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".jpg", "image/jpeg")
	mime.AddExtensionType(".jpeg", "image/jpeg")
	mime.AddExtensionType(".gif", "image/gif")
	mime.AddExtensionType(".svg", "image/svg+xml")
	mime.AddExtensionType(".woff", "font/woff")
	mime.AddExtensionType(".woff2", "font/woff2")
	mime.AddExtensionType(".ttf", "font/ttf")
}

type Server struct {
	router    *mux.Router
	collector *collector.Collector
	templates *template.Template
	srv       *http.Server
}

func New(collector *collector.Collector) *Server {
	s := &Server{
		router:    mux.NewRouter(),
		collector: collector,
	}

	s.templates = template.Must(template.ParseGlob("web/templates/*html"))

	s.routes()

	return s
}

func (s *Server) routes() {
	s.router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))),
	)

	s.router.HandleFunc("/", s.handleIndex())

	s.router.HandleFunc("/api/stats", s.handleApiStats())
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.templates.ExecuteTemplate(w, "index.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) handleApiStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := s.collector.GetStats()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) Start(port string) error {
	s.srv = &http.Server{
		Handler:      s.router,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
