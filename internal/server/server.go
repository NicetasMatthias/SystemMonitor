package server

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"log/slog"
	"mime"
	"net"
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
	port      string
}

type apiError struct {
	Error string `json:"error"`
}

func New(c statsCollector, port string) (*Server, error) {
	s := &Server{
		router:    mux.NewRouter(),
		collector: c,
		port:      port,
	}

	s.templates = template.Must(template.ParseFS(web.FS, "templates/*.html"))

	if err := s.routes(); err != nil {
		return nil, err
	}

	s.srv = &http.Server{
		Handler:      s.router,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	return s, nil
}

func (s *Server) Start() error {
	slog.Info("Server staring",
		slog.String("port", s.port))

	listener, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return err
	}

	go func() {

		if err := s.srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			//==== TODO: log
		}
	}()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) routes() error {

	staticFS, err := fs.Sub(web.FS, "static")
	if err != nil {
		//=== TODO: log
		return err
	}

	//=== TODO: improve processing MethodNotAllowed

	s.router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", http.MethodGet)
		writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
	})

	s.router.PathPrefix("/static/").Handler(
		http.StripPrefix(
			"/static/",
			http.FileServer(http.FS(staticFS)),
		),
	)

	s.router.HandleFunc("/", s.handleIndex()).Methods(http.MethodGet)

	s.router.HandleFunc("/api/stats", s.handleApiStats()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/stats/", s.handleApiStats()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/stats/cpu", s.handleApiStatsCPU()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/stats/cpu/", s.handleApiStatsCPU()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/stats/disk", s.handleApiStatsDisk()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/stats/disk/", s.handleApiStatsDisk()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/stats/memory", s.handleApiStatsMemory()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/stats/memory/", s.handleApiStatsMemory()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/stats/network", s.handleApiStatsNetwork()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/stats/network/", s.handleApiStatsNetwork()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/stats/system", s.handleApiStatsSystem()).Methods(http.MethodGet)
	s.router.HandleFunc("/api/stats/system/", s.handleApiStatsSystem()).Methods(http.MethodGet)

	return nil //=== TODO: сделать проверки где можно
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
