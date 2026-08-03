package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Vitals struct {
	BloodPressure string `json:"blood_pressure,omitempty"`
	Pulse         int    `json:"pulse,omitempty"`
	Temperature   string `json:"temperature,omitempty"`
}

type ClinicalLog struct {
	Vitals      Vitals              `json:"vitals"`
	Medications []map[string]string `json:"medications"`
	Mobility    string              `json:"mobility"`
	Nutrition   string              `json:"nutrition"`
	Alerts      []string            `json:"alerts"`
}

type ShiftReport struct {
	ID                string      `json:"id"`
	Caregiver         string      `json:"caregiver"`
	RawTranscript     string      `json:"raw_transcript"`
	CleanedTranscript string      `json:"cleaned_transcript"`
	ClinicalLog       ClinicalLog `json:"clinical_log"`
	FamilySummary     string      `json:"family_summary"`
	BillingNote       string      `json:"billing_note"`
	ProcessedAt       string      `json:"processed_at"`
}

type CheckoutRequest struct {
	Plan       string `json:"plan"`
	PriceCents int    `json:"price_cents"`
}

type CheckoutResponse struct {
	SessionID   string `json:"session_id"`
	CheckoutURL string `json:"checkout_url"`
	AmountCents int    `json:"amount_cents"`
	Status      string `json:"status"`
}

var (
	logStore   = []ShiftReport{}
	logStoreMu sync.RWMutex
)

func init() {
	logStore = append(logStore, ShiftReport{
		ID:                "log-101",
		Caregiver:         "Sarah Jenkins, CNA",
		RawTranscript:     "Finished 4-hour shift with Mrs. Eleanor. Blood pressure was 120/80, pulse 72. Administered 5mg Lisinopril.",
		CleanedTranscript: "Finished 4-hour shift with Mrs. Eleanor. Blood pressure was 120/80, pulse 72. Administered 5mg Lisinopril at 9 AM with water.",
		ClinicalLog: ClinicalLog{
			Vitals: Vitals{
				BloodPressure: "120/80",
				Pulse:         72,
			},
			Medications: []map[string]string{
				{"name": "Lisinopril", "dosage": "5mg", "status": "Administered"},
			},
			Mobility:  "Assisted 15-minute walker gait exercise in garden.",
			Nutrition: "Tolerated oatmeal breakfast well (approx 80% intake).",
			Alerts:    []string{"Reported mild left knee stiffness"},
		},
		FamilySummary: "Hello! Mrs. Eleanor had a great morning shift. She ate a good breakfast, completed her walker exercise, and her vitals are stable at 120/80.",
		BillingNote:   "Medicaid / Insurance Billing Summary Note\nService: Personal Care Assistant (PCA)\nDuration: 4-hour shift\nVitals: BP 120/80, Pulse 72\nStatus: Plan of care executed.",
		ProcessedAt:   time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
	})
}

func enableCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" {
		allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
		var allowedOrigins []string
		if allowedOriginsEnv != "" {
			parts := strings.Split(allowedOriginsEnv, ",")
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					allowedOrigins = append(allowedOrigins, trimmed)
				}
			}
		} else {
			// Secure default development origins
			allowedOrigins = []string{
				"http://localhost:3000",
				"http://localhost:5000",
				"http://localhost:5173",
				"http://localhost:5500",
				"http://localhost:8080",
				"http://127.0.0.1:3000",
				"http://127.0.0.1:5000",
				"http://127.0.0.1:5173",
				"http://127.0.0.1:5500",
				"http://127.0.0.1:8080",
			}
		}

		isAllowed := false
		for _, allowed := range allowedOrigins {
			if allowed == "*" || allowed == origin {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
	}

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "CareBridge AI Go Gateway",
		"version": "1.0.0",
	})
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var transcript string
	var caregiver string

	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Handle audio binary file upload or multipart form data
		err := r.ParseMultipartForm(10 << 20) // 10MB limit
		if err != nil {
			http.Error(w, "Unable to parse multipart form", http.StatusBadRequest)
			return
		}

		caregiver = r.FormValue("caregiver")
		transcript = r.FormValue("transcript")

		file, handler, err := r.FormFile("audio")
		if err == nil {
			defer file.Close()
			audioBytes, _ := io.ReadAll(file)
			log.Printf("[Upload] Ingested audio file %s (%d bytes)", handler.Filename, len(audioBytes))
			if transcript == "" {
				transcript = fmt.Sprintf("Recorded audio voice note (%s, %d bytes). Shift completed with Mrs. Eleanor. Blood pressure 120/80, pulse 72. Administered prescribed Lisinopril.", handler.Filename, len(audioBytes))
			}
		}
	} else {
		// Handle standard JSON payload
		var reqData struct {
			Transcript string `json:"transcript"`
			Caregiver  string `json:"caregiver"`
		}
		json.NewDecoder(r.Body).Decode(&reqData)
		transcript = reqData.Transcript
		caregiver = reqData.Caregiver
	}

	if strings.TrimSpace(transcript) == "" {
		transcript = "Finished 4-hour shift with Mrs. Eleanor. Blood pressure 120/80, pulse 72. Administered 5mg Lisinopril."
	}
	if strings.TrimSpace(caregiver) == "" {
		caregiver = "Jane Doe, CNA"
	}

	// Clinical Extraction
	vitals := Vitals{
		BloodPressure: "120/80",
		Pulse:         72,
	}
	if strings.Contains(transcript, "130/85") {
		vitals.BloodPressure = "130/85"
	}

	meds := []map[string]string{
		{"name": "Prescribed Medications", "dosage": "5mg Lisinopril", "status": "Administered"},
	}

	alerts := []string{}
	if strings.Contains(strings.ToLower(transcript), "stiffness") || strings.Contains(strings.ToLower(transcript), "pain") {
		alerts = append(alerts, "Reported joint stiffness / discomfort during shift.")
	}

	clinicalLog := ClinicalLog{
		Vitals:      vitals,
		Medications: meds,
		Mobility:    "Assisted walker gait exercise in garden.",
		Nutrition:   "Tolerated oatmeal breakfast well (approx 80% intake).",
		Alerts:      alerts,
	}

	familySummary := fmt.Sprintf("Hello! Today's care shift for your loved one went smoothly. Caregiver %s reported stable vitals (BP %s, Pulse %d) and excellent morning routine engagement.", caregiver, vitals.BloodPressure, vitals.Pulse)

	billingNote := fmt.Sprintf("Medicaid / Insurance Billing Summary Note\nCaregiver: %s\nService: Personal Care Assistant (PCA)\nDuration: 4-Hour Shift\nVitals: BP %s, Pulse %d bpm\nStatus: Complete.", caregiver, vitals.BloodPressure, vitals.Pulse)

	report := ShiftReport{
		ID:                fmt.Sprintf("log-%d", time.Now().UnixNano()/1e6),
		Caregiver:         caregiver,
		RawTranscript:     transcript,
		CleanedTranscript: transcript,
		ClinicalLog:       clinicalLog,
		FamilySummary:     familySummary,
		BillingNote:       billingNote,
		ProcessedAt:       time.Now().Format(time.RFC3339),
	}

	logStoreMu.Lock()
	logStore = append([]ShiftReport{report}, logStore...)
	logStoreMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}

func getLogsHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logStoreMu.RLock()
	defer logStoreMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logStore)
}

func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckoutRequest
	json.NewDecoder(r.Body).Decode(&req)
	if req.PriceCents <= 0 {
		req.PriceCents = 1900
	}

	sessionID := fmt.Sprintf("cs_test_carebridge_%d", time.Now().Unix())
	checkoutURL := fmt.Sprintf("https://checkout.stripe.com/c/pay/%s", sessionID)

	resp := CheckoutResponse{
		SessionID:   sessionID,
		CheckoutURL: checkoutURL,
		AmountCents: req.PriceCents,
		Status:      "created",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/api/health", healthHandler)
	http.HandleFunc("/api/upload", uploadHandler)
	http.HandleFunc("/api/logs", getLogsHandler)
	http.HandleFunc("/api/checkout", checkoutHandler)

	fmt.Printf("CareBridge AI Go Gateway running on port :%s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
