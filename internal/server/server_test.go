package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/NicetasMatthias/SystemMonitor/internal/collector"
)

type testCollector struct {
	stats   collector.CollectorExport
	cpu     collector.CPUExport
	disk    collector.DiskExport
	memory  collector.MemoryExport
	network collector.NetworkExport
	system  collector.SystemExport
}

func (c *testCollector) Get() collector.CollectorExport {
	return c.stats
}

func (c *testCollector) GetCPU() collector.CPUExport {
	return c.cpu
}

func (c *testCollector) GetDisk() collector.DiskExport {
	return c.disk
}

func (c *testCollector) GetMemory() collector.MemoryExport {
	return c.memory
}

func (c *testCollector) GetNetwork() collector.NetworkExport {
	return c.network
}

func (c *testCollector) GetSystem() collector.SystemExport {
	return c.system
}

func Test_writeAPIError(t *testing.T) {

	tests := []struct {
		name    string
		status  int
		message string
	}{
		{
			name:    "bad request",
			status:  http.StatusBadRequest,
			message: "bad request",
		},
		{
			name:    "method not allowed",
			status:  http.StatusMethodNotAllowed,
			message: "method not allowed",
		},
		{
			name:    "internal server error",
			status:  http.StatusInternalServerError,
			message: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			recorder := httptest.NewRecorder()
			writeAPIError(recorder, tt.status, tt.message)

			if recorder.Code != tt.status {
				t.Errorf("status = %d, want %d", recorder.Code, tt.status)
			}

			if ctHeader := recorder.Header().Get("Content-Type"); ctHeader != "application/json" {
				t.Errorf("Content-Type = %q, want %q", ctHeader, "application/json")
			}

			if corsHeader := recorder.Header().Get("Access-Control-Allow-Origin"); corsHeader != "*" {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", corsHeader, "*")
			}

			var response apiError
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response.Error != tt.message {
				t.Errorf("error = %q, want %q", response.Error, tt.message)
			}

		})
	}

}

func Test_writeJSON(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		inputData map[string]any
	}{
		{
			name:      "simple JSON",
			status:    http.StatusOK,
			inputData: map[string]any{"status": "ok"},
		},
		{
			name:      "created response",
			status:    http.StatusCreated,
			inputData: map[string]any{"status": "created"},
		},
		{
			name:      "nested JSON",
			status:    http.StatusCreated,
			inputData: map[string]any{"key": map[string]any{"nested_key": "data"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			recorder := httptest.NewRecorder()
			writeJSON(recorder, tt.status, tt.inputData)

			if recorder.Code != tt.status {
				t.Errorf("status = %d, want %d", recorder.Code, tt.status)
			}

			if ctHeader := recorder.Header().Get("Content-Type"); ctHeader != "application/json" {
				t.Errorf("Content-Type = %q, want %q", ctHeader, "application/json")
			}

			if corsHeader := recorder.Header().Get("Access-Control-Allow-Origin"); corsHeader != "*" {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", corsHeader, "*")
			}

			var response map[string]any
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if !reflect.DeepEqual(response, tt.inputData) {
				t.Errorf("the data differs")
			}

		})
	}
}

func Test_writeJSONEncodingError(t *testing.T) {

	type Node struct {
		Next *Node `json:"next"`
	}

	var cyclicNode Node
	cyclicNode.Next = &cyclicNode

	tests := []struct {
		name   string
		status int
		value  any
	}{
		{
			name:   "cyclic struct",
			status: http.StatusOK,
			value:  cyclicNode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			writeJSON(recorder, tt.status, tt.value)

			if recorder.Code != http.StatusInternalServerError {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
			}

			if ctHeader := recorder.Header().Get("Content-Type"); ctHeader != "application/json" {
				t.Errorf("Content-Type = %q, want %q", ctHeader, "application/json")
			}

			var response apiError

			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response.Error != "failed to encode response" {
				t.Errorf("error = %q, want %q", response.Error, "failed to encode response")
			}

		})
	}

}

func TestServer_APIStatsGET(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "GET /api/stats",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := New(&testCollector{}, nil, DefaultConfig())
			if err != nil {
				t.Fatalf("failed to setup server: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)

			recorder := httptest.NewRecorder()

			s.router.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
			}

			if ctHeader := recorder.Header().Get("Content-Type"); ctHeader != "application/json" {
				t.Errorf("Content-Type = %q, want %q", ctHeader, "application/json")
			}

			if corsHeader := recorder.Header().Get("Access-Control-Allow-Origin"); corsHeader != "*" {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", corsHeader, "*")
			}

			var response collector.CollectorExport

			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
		})
	}
}

func TestAPIStatsMethodNotAllowed(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{
			name:   "POST",
			method: http.MethodPost,
		},
		{
			name:   "PUT",
			method: http.MethodPut,
		},
		{
			name:   "PATCH",
			method: http.MethodPatch,
		},
		{
			name:   "DELETE",
			method: http.MethodDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := New(&testCollector{}, nil, DefaultConfig())
			if err != nil {
				t.Fatalf("failed to setup server: %v", err)
			}

			req := httptest.NewRequest(tt.method, "/api/stats", nil)

			recorder := httptest.NewRecorder()

			s.router.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusMethodNotAllowed {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
			}

			if allowHeader := recorder.Header().Get("Allow"); allowHeader != http.MethodGet {
				t.Errorf("Allow = %q, want %q", allowHeader, http.MethodGet)
			}

			if ctHeader := recorder.Header().Get("Content-Type"); ctHeader != "application/json" {
				t.Errorf("Content-Type = %q, want %q", ctHeader, "application/json")
			}

			var response apiError

			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response.Error != "method not allowed" {
				t.Errorf("error = %q, want %q", response.Error, "method not allowed")
			}
		})
	}
}
