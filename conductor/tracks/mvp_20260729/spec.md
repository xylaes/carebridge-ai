# Track Specification: CareBridge AI MVP

## Overview
CareBridge AI is an AI administrative assistant for in-home caregivers, private aides, and small home care agencies. It ingests 45-second caregiver voice notes post-shift and generates three structured artifacts: a Clinical Care Shift Log, a Medicaid/insurance billing note, and a warm family update summary.

## Functional Requirements
1. **Core Processing Engine (Python - `carebridge/`)**:
   - Audio / Text Processing Module (`carebridge/analyzer.py`):
     - Voice Ingestion / Transcription & Cleaning (`VoiceIngestionAgent`)
     - Clinical Metric Extraction (`ClinicalLogAgent`): Extracts medications, mobility, nutrition, and vitals.
     - Family Summary Generation (`FamilySummaryAgent`): Generates warm, empathetic layperson updates.
     - Medicaid/Insurance Billing Note Generation: Structured audit note.
2. **Backend API Gateway (Go - `backend/`)**:
   - `POST /api/upload`: Endpoints for audio/text note upload & analysis triggering.
   - `GET /api/logs`: Fetch historical shift logs.
   - `POST /api/checkout`: Stripe subscription checkout session endpoint ($19/mo).
3. **Frontend Dashboard (Mobile-First Web - `frontend/`)**:
   - Audio recorder UI with one-tap start/stop.
   - Real-time processing indicator.
   - Tabbed/card view for Clinical Log, Billing Note, and Family Update.

## Non-Functional Requirements
- Unit test coverage >80% for core Python analyzer modules and Go endpoints.
- Lightweight, fast execution (<5 seconds response time for text analysis with mock/fallback support).
- Mobile-first responsive layout.

## Acceptance Criteria
- `carebridge/analyzer.py` implements all 3 core generation functions with full fallback capabilities and structured data outputs.
- Comprehensive unit tests in `carebridge/tests/test_analyzer.py` pass with 100% success rate.
- Go backend router in `backend/main.go` handles HTTP requests for upload, logs, and checkout.
- Frontend interface in `frontend/index.html` offers interactive voice note recording & report preview.
