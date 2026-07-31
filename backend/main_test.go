package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestUploadHandlerJSON(t *testing.T) {
	payload := map[string]string{
		"transcript": "Finished 4-hour shift with Mrs. Eleanor. Blood pressure was 120/80, pulse 72. Administered 5mg Lisinopril.",
		"caregiver":  "Jane Doe",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/api/upload", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("uploadHandler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if resp["family_summary"] == nil {
		t.Errorf("Expected family_summary in response")
	}
}

func TestUploadHandlerMultipart(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("caregiver", "Sarah CNA")
	writer.WriteField("transcript", "Shift note via audio recording.")

	part, err := writer.CreateFormFile("audio", "shift_voice_note.webm")
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte("MOCK_AUDIO_HEADER_BYTES"))
	writer.Close()

	req, err := http.NewRequest("POST", "/api/upload", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("uploadHandler multipart returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestGetLogsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/logs", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getLogsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("getLogsHandler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestCheckoutHandler(t *testing.T) {
	payload := map[string]interface{}{
		"plan":        "pro_monthly",
		"price_cents": 1900,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(checkoutHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("checkoutHandler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if resp["checkout_url"] == nil {
		t.Errorf("Expected checkout_url in response")
	}
}

func TestCORSLogic(t *testing.T) {
	// 1. No Origin header
	req, _ := http.NewRequest("GET", "/api/health", nil)
	rr := httptest.NewRecorder()
	http.HandlerFunc(healthHandler).ServeHTTP(rr, req)
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "" {
		t.Errorf("Expected no Access-Control-Allow-Origin header when Origin request header is absent, got %s", origin)
	}

	// 2. Default allowed origin (localhost:3000) when ALLOWED_ORIGINS is empty
	req, _ = http.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr = httptest.NewRecorder()
	http.HandlerFunc(healthHandler).ServeHTTP(rr, req)
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:3000" {
		t.Errorf("Expected Access-Control-Allow-Origin to be http://localhost:3000, got %s", origin)
	}
	if vary := rr.Header().Get("Vary"); vary != "Origin" {
		t.Errorf("Expected Vary header to be Origin, got %s", vary)
	}

	// 3. Untrusted origin when ALLOWED_ORIGINS is empty
	req, _ = http.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "http://malicious.com")
	rr = httptest.NewRecorder()
	http.HandlerFunc(healthHandler).ServeHTTP(rr, req)
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "" {
		t.Errorf("Expected no Access-Control-Allow-Origin header for untrusted origin, got %s", origin)
	}

	// 4. Custom ALLOWED_ORIGINS config
	t.Setenv("ALLOWED_ORIGINS", "https://carebridge.ai, https://app.carebridge.ai")

	// Custom trusted origin
	req, _ = http.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "https://carebridge.ai")
	rr = httptest.NewRecorder()
	http.HandlerFunc(healthHandler).ServeHTTP(rr, req)
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "https://carebridge.ai" {
		t.Errorf("Expected Access-Control-Allow-Origin to be https://carebridge.ai, got %s", origin)
	}

	// Localhost now untrusted
	req, _ = http.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr = httptest.NewRecorder()
	http.HandlerFunc(healthHandler).ServeHTTP(rr, req)
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "" {
		t.Errorf("Expected localhost to be untrusted under custom ALLOWED_ORIGINS config, got %s", origin)
	}

	// 5. Wildcard ALLOWED_ORIGINS config
	t.Setenv("ALLOWED_ORIGINS", "*")
	req, _ = http.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "http://any-random-domain.com")
	rr = httptest.NewRecorder()
	http.HandlerFunc(healthHandler).ServeHTTP(rr, req)
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "http://any-random-domain.com" {
		t.Errorf("Expected wildcard configuration to allow any-random-domain.com, got %s", origin)
	}
}
