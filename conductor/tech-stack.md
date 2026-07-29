# Technology Stack: CareBridge AI

## Core AI & Data Processing
- **Language**: Python 3.11+
- **SDK**: `google-genai` SDK (Gemini 3.5 Flash)
- **Core Module**: `carebridge/analyzer.py` for transcription, clinical metric extraction, and family summary generation.

## Backend Gateway
- **Language**: Go (1.21+)
- **HTTP Server**: Standard library `net/http` API Gateway
- **REST Endpoints**:
  - `POST /api/upload`: Audio upload & processing trigger
  - `GET /api/logs`: Shift log history
  - `POST /api/checkout`: Stripe subscription checkout ($19/mo)

## Agentic Orchestration
- **Framework**: Google Antigravity Sub-agents
- **Agents**:
  1. `Voice Ingestion Agent`: Speech-to-text transcription & noise cleaning
  2. `Clinical Log Agent`: Extraction of medications, mobility, nutrition, vitals
  3. `Family Summary Agent`: Empathetic layperson summary generation

## Frontend Dashboard
- **Tech**: HTML5, CSS3, JavaScript (Mobile-first responsive dashboard)
- **Key Views**: Voice recording trigger, Clinical Shift Log preview, Billing Note preview, Family Update preview.

## Testing & Tooling
- **Testing**: Python `unittest` / `pytest`, Go `testing` package
- **Environment**: Virtual environment (`venv`), `.env` configuration
