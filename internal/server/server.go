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

type statsCollector interface {
	Get() collector.CollectorExport
	GetCPU() collector.CPUExport
	GetDisk() collector.DiskExport
	GetMemory() collector.MemoryExport
	GetNetwork() collector.NetworkExport
	GetSystem() collector.SystemExport
}

type Server struct {
	router    *mux.Router
	collector statsCollector
	templates *template.Template
	srv       *http.Server
}

type apiError struct {
	Error string `json:"error"`
}

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

func New(c statsCollector) *Server {
	s := &Server{
		router:    mux.NewRouter(),
		collector: c,
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

	s.router.HandleFunc("/", s.handleIndex()).Methods(http.MethodGet)

	apiStats := s.router.PathPrefix("/api/stats").Subrouter()

	apiStats.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", http.MethodGet)
		writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
	})

	apiStats.HandleFunc("", s.handleApiStats()).Methods(http.MethodGet)
	apiStats.HandleFunc("/", s.handleApiStats()).Methods(http.MethodGet)
	apiStats.HandleFunc("/cpu", s.handleApiStatsCPU()).Methods(http.MethodGet)
	apiStats.HandleFunc("/cpu/", s.handleApiStatsCPU()).Methods(http.MethodGet)
	apiStats.HandleFunc("/disk", s.handleApiStatsDisk()).Methods(http.MethodGet)
	apiStats.HandleFunc("/disk/", s.handleApiStatsDisk()).Methods(http.MethodGet)
	apiStats.HandleFunc("/memory", s.handleApiStatsMemory()).Methods(http.MethodGet)
	apiStats.HandleFunc("/memory/", s.handleApiStatsMemory()).Methods(http.MethodGet)
	apiStats.HandleFunc("/network", s.handleApiStatsNetwork()).Methods(http.MethodGet)
	apiStats.HandleFunc("/network/", s.handleApiStatsNetwork()).Methods(http.MethodGet)
	apiStats.HandleFunc("/system", s.handleApiStatsSystem()).Methods(http.MethodGet)
	apiStats.HandleFunc("/system/", s.handleApiStatsSystem()).Methods(http.MethodGet)
}

func writeAPIError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(apiError{
		Error: message,
	})

	if err != nil {
		slog.Error("failed to encode error to response", slog.Any("encode error", err), slog.Any("response error", message))
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		slog.Error("failed to encode response", slog.Any("error", err))
		writeAPIError(w, http.StatusInternalServerError, "failed to encode response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	_, err = w.Write(data)
	if err != nil {
		slog.Error("failed to write data to response", slog.Any("error", err))

	}
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
		writeJSON(w, http.StatusOK, s.collector.Get())
	}
}

func (s *Server) handleApiStatsCPU() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.collector.GetCPU())
	}
}

func (s *Server) handleApiStatsDisk() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.collector.GetDisk())
	}
}

func (s *Server) handleApiStatsMemory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.collector.GetMemory())
	}
}

func (s *Server) handleApiStatsNetwork() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.collector.GetNetwork())
	}
}

func (s *Server) handleApiStatsSystem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.collector.GetSystem())
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
