package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"yabbi_test/internal/auction"
)

type fakeService struct {
	resp   auction.Response
	panics bool
}

func (f *fakeService) RunAuction(_ context.Context, _ auction.Request) auction.Response {
	if f.panics {
		panic("boom from fake service")
	}
	return f.resp
}

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHandler_ValidRequest_Returns200(t *testing.T) {
	fs := &fakeService{resp: auction.Response{
		RequestID:  "r1",
		MatchedDSP: []string{"dsp-alpha"},
		Sent:       1,
		Succeeded:  1,
		DurationMs: 5,
	}}
	h := NewHandler(fs, silentLogger())

	body := `{"request_id":"r1","country":"RU","device_type":"mobile","bid_floor":1.0,"categories":["news"]}`
	req := httptest.NewRequest(http.MethodPost, "/auction", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp auction.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "r1", resp.RequestID)
	require.Equal(t, 1, resp.Sent)
	require.Equal(t, 1, resp.Succeeded)
}

func TestHandler_InvalidJSON_Returns400(t *testing.T) {
	h := NewHandler(&fakeService{}, silentLogger())

	req := httptest.NewRequest(http.MethodPost, "/auction", bytes.NewBufferString(`{not json`))
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_MissingRequiredField_Returns400(t *testing.T) {
	h := NewHandler(&fakeService{}, silentLogger())

	// нет request_id
	body := `{"country":"RU","device_type":"mobile","bid_floor":1.0}`
	req := httptest.NewRequest(http.MethodPost, "/auction", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var apiErr apiError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &apiErr))
	require.Contains(t, apiErr.Error, "request_id")
}

func TestHandler_UnknownField_Returns400(t *testing.T) {
	h := NewHandler(&fakeService{}, silentLogger())

	body := `{"request_id":"r1","country":"RU","device_type":"mobile","bid_floor":1.0,"typo_field":"x"}`
	req := httptest.NewRequest(http.MethodPost, "/auction", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_InvalidDeviceType_Returns400(t *testing.T) {
	h := NewHandler(&fakeService{}, silentLogger())

	body := `{"request_id":"r1","country":"RU","device_type":"smarttv","bid_floor":1.0}`
	req := httptest.NewRequest(http.MethodPost, "/auction", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_WrongMethod_Returns405(t *testing.T) {
	h := NewHandler(&fakeService{}, silentLogger())

	req := httptest.NewRequest(http.MethodGet, "/auction", nil)
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandler_UnknownPath_Returns404(t *testing.T) {
	h := NewHandler(&fakeService{}, silentLogger())

	req := httptest.NewRequest(http.MethodPost, "/does-not-exist", nil)
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_ServicePanics_Returns500NotCrash(t *testing.T) {
	fs := &fakeService{panics: true}
	h := NewHandler(fs, silentLogger())

	body := `{"request_id":"r1","country":"RU","device_type":"mobile","bid_floor":1.0}`
	req := httptest.NewRequest(http.MethodPost, "/auction", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	require.NotPanics(t, func() {
		h.Routes().ServeHTTP(w, req)
	})
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
