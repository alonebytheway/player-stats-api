package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCalculatedKD(t *testing.T) {
	tests := []struct {
		name     string
		kills    int
		deaths   int
		expected float64
	}{
		{name: "alonebtw", kills: 10, deaths: 2, expected: 5.0},
		{name: "DenisiniPenisini", kills: 2, deaths: 10, expected: 0.2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateKD(tt.kills, tt.deaths)
			if result != tt.expected {
				t.Errorf("Для %s ожидалось %v, а получили %v", tt.name, tt.expected, result)
			}
		})
	}

}
func intPtr(i int) *int {
	return &i
}
func TestPlayerUpdateRequest_Validator(t *testing.T) {
	tests := []struct {
		name    string
		req     UpdatePlayer
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     UpdatePlayer{Kills: intPtr(5), Deaths: intPtr(2)},
			wantErr: false,
		},
		{
			name:    "negative kills",
			req:     UpdatePlayer{Kills: intPtr(-5), Deaths: intPtr(2)},
			wantErr: true,
		},
		{
			name:    "negative deaths",
			req:     UpdatePlayer{Kills: intPtr(5), Deaths: intPtr(-2)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Test Miss: wait err = %v, but take = %v", tt.wantErr, err)
			}
		})
	}
}

type MockRepository struct {
	Repository

	UpdateErr error
}

func (m *MockRepository) Update(ctx context.Context, name string, update UpdatePlayer) error {
	return m.UpdateErr
}

func TestUpdatePlayer_Success(t *testing.T) {
	mockRepo := &MockRepository{
		UpdateErr: nil,
	}

	svc := &PlayerService{
		repo: mockRepo,
	}

	handler := &PlayerHandler{
		service: svc,
	}

	body := `{"kills": 5, "deaths": 2}`

	req := httptest.NewRequest(http.MethodPut, "/players/alonebtw", bytes.NewBufferString(body))

	req.Header.Set("Content-Type", "application/json")

	r := chi.NewRouter()
	r.Put("/players/{name}", handler.UpdatePlayer)

	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Take status %d, answered body: %s", rr.Code, rr.Body.String())
	}

}

func TestUpdatePlayer_InvalideJSON(t *testing.T) {
	mockRepo := &MockRepository{
		UpdateErr: nil,
	}

	svc := &PlayerService{
		repo: mockRepo,
	}
	handler := &PlayerHandler{
		service: svc,
	}
	body := `{"kills": 5, "deaths": 2`

	req := httptest.NewRequest(http.MethodPut, "/players/alonebtw", bytes.NewBufferString(body))

	req.Header.Set("Content-Type", "application/json")

	r := chi.NewRouter()
	r.Put("/players/{name}", handler.UpdatePlayer)

	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Take status %d, answered body: %s", rr.Code, rr.Body.String())
	}

}

func TestUpdatePlayer_ServerError(t *testing.T) {
	mockRepo := &MockRepository{
		UpdateErr: errors.New("database is on fire"),
	}

	svc := &PlayerService{
		repo: mockRepo,
	}

	handler := &PlayerHandler{
		service: svc,
	}

	body := `{"kills": 5, "deaths": 2}`

	req := httptest.NewRequest(http.MethodPut, "/players/alonebtw", bytes.NewBufferString(body))

	req.Header.Set("Content-Type", "application/json")

	r := chi.NewRouter()
	r.Put("/players/{name}", handler.UpdatePlayer)

	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Take status %d, answered body: %s", rr.Code, rr.Body.String())
	}

}
