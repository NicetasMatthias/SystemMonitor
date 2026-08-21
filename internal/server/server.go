package server

import (
	"context"
	"encoding/json"
	"html/template"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"time"

	"github.com/NicetasMatthias/SystemMonitor/internal/collector"
	"github.com/NicetasMatthias/SystemMonitor/internal/web"
	"github.com/gorilla/mux"
)

func init() {
	addMimeExtType(".css", "text/css")
	addMimeExtType(".js", "application/javascript")
	addMimeExtType(".json", "application/json")
	addMimeExtType(".png", "image/png")
	addMimeExtType(".jpg", "image/jpeg")
	addMimeExtType(".jpeg", "image/jpeg")
	addMimeExtType(".gif", "image/gif")
	addMimeExtType(".svg", "image/svg+xml")
	addMimeExtType(".woff", "font/woff")
	addMimeExtType(".woff2", "font/woff2")
	addMimeExtType(".ttf", "font/ttf")
}

func addMimeExtType(ext, typeStr string) {
	if err := mime.AddExtensionType(ext, typeStr); err != nil {
		slog.Error("Failed to add mime extension type",
			slog.Any("extension", ext),
			slog.Any("mime type", typeStr),
			slog.Any("error", err))
	}
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

	s.templates = template.Must(template.ParseFS(web.FS, "templates/*.html"))

	s.routes()

	return s
}

func (s *Server) routes() {

	staticFS, err := fs.Sub(web.FS, "static")
	if err != nil {
		panic(err)
	}

	s.router.PathPrefix("/static/").Handler(
		http.StripPrefix(
			"/static/",
			http.FileServer(http.FS(staticFS)),
		),
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
		stats := s.collector.Get()

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
