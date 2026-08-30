package server

import (
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
		name   string
		status int
		data   map[string]string
	}{
		{
			name:   "Simple JSON",
			status: http.StatusOK,
			data:   map[string]string{"status": "ok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			recorder := httptest.NewRecorder()
			writeJSON(recorder, tt.status, tt.data)

			if recorder.Code != tt.status {
				t.Errorf("status = %d, want %d", recorder.Code, tt.status)
			}

			if ctHeader := recorder.Header().Get("Content-Type"); ctHeader != "application/json" {
				t.Errorf("Content-Type = %q, want %q", ctHeader, "application/json")
			}

			var response map[string]string
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if !maps.Equal(response, tt.data) {
				t.Errorf("the data differs")
			}

		})
	}
}

// /=== TODO: сделать норм
func TestServer_APIStatsGET(t *testing.T) {
	tests := []struct {
		name string
		//=== FIXME: WIP
	}{
		{
			name: "DUMMY",
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//=== FIXME: WIP
		})
	}
}

func TestAPIStatsMethodNotAllowed(t *testing.T) {
	tests := []struct {
		name string
		//=== FIXME: WIP
	}{
		{
			name: "DUMMY",
		},
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//=== FIXME: WIP
		})
	}
}
