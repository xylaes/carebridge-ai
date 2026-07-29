# Track Specification: Frontend API Integration & Dual-Column Dashboard

## Overview
Connect the frontend responsive web dashboard (`frontend/`) to the Go backend REST API (`/api/upload`, `/api/logs`, `/api/checkout`), enabling audio recording and binary audio/text upload, rendering Clinical Care Shift Logs and Warm Family Summaries side-by-side on desktop (stacked on mobile), and providing an integrated $19/mo Stripe checkout flow.

## Functional Requirements
1. **Audio Recording & Upload (`frontend/app.js` & `backend/main.go`)**:
   - Utilize Browser `MediaRecorder` API to capture real voice audio blobs (.webm/.wav).
   - Send audio blobs or JSON payloads to `POST /api/upload`.
   - Update Go backend `uploadHandler` to accept multipart audio files as well as JSON transcripts, passing audio/text to the Python analyzer engine.
2. **Side-by-Side Dual Column View (`frontend/index.html` & `frontend/style.css`)**:
   - Render Clinical Care Shift Log (left column) and Warm Family Update Summary + Medicaid Billing Note (right column) side-by-side on desktop devices (>768px).
   - Collapse into responsive stacked views on mobile screens.
3. **Stripe Checkout Integration (`frontend/app.js`)**:
   - Prominent "$19/mo Pro Upgrade" button in navigation and report cards.
   - Interactive checkout modal triggering `POST /api/checkout` and displaying redirect session links.

## Acceptance Criteria
- MediaRecorder captures audio blobs and posts to `/api/upload`.
- Dashboard renders Clinical Log and Family Update side-by-side in grid/flex container.
- Go backend `backend/main.go` parses multipart audio uploads and JSON payloads seamlessly.
- Go backend unit tests in `backend/main_test.go` pass 100%.
