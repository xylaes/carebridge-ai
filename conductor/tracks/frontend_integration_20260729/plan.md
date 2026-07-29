# Implementation Plan: Frontend Integration & Dual-Column Dashboard

## Phase 1: Go Backend Multipart Audio Endpoint Support
- [ ] Task: Update `backend/main.go` `uploadHandler` to handle multipart audio form uploads (`audio/webm`, `audio/wav`) in addition to JSON payloads
- [ ] Task: Add multipart upload unit test to `backend/main_test.go`
- [ ] Task: Verify Go test suite passes

## Phase 2: Dual-Column Layout & Audio Recorder Integration
- [ ] Task: Update `frontend/index.html` to establish side-by-side grid layout for Clinical Log and Family Summary / Billing Note
- [ ] Task: Update `frontend/style.css` with responsive dual-column styles (`@media (min-width: 768px)`)
- [ ] Task: Update `frontend/app.js` with `MediaRecorder` audio blob capture and FormData POST upload to `/api/upload`

## Phase 3: Verification & End-to-End Test Checkpoint
- [ ] Task: Run Go backend tests and verify API endpoints
- [ ] Task: Manual verification of web dashboard layout and Stripe checkout modal
