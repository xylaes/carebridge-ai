package main

import (
	"bytes"
	"encoding/json"
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

func TestUploadHandler(t *testing.T) {
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

	if status := rr.Code; status != http.StatusOK && status != http.StatusCreated {
		t.Errorf("uploadHandler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if resp["family_summary"] == nil {
		t.Errorf("Expected family_summary in response")
	}
	if resp["billing_note"] == nil {
		t.Errorf("Expected billing_note in response")
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
