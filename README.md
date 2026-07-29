# CareBridge AI (Voice-to-Report Caregiver Documentation Assistant)

CareBridge AI is an AI administrative assistant for in-home caregivers, private aides, and small home care agencies. Powered by **Gemini 3.5 Flash** and **Google Antigravity**, CareBridge AI transforms raw caregiver voice notes into:
1. **Clinical Care Shift Log** (Vitals, Medications, Mobility, Nutrition)
2. **Medicaid / Insurance Billing Note**
3. **Warm Family Update Summary**

## System Architecture
- `carebridge/`: Python Core engine using `google-genai` SDK for audio & text transcription and clinical extraction.
- `backend/`: Go `net/http` API gateway serving REST endpoints (`/api/upload`, `/api/logs`, `/api/checkout`).
- `frontend/`: Mobile-first responsive web dashboard.

## Running Tests
```bash
python -m unittest discover -s carebridge/tests -p "test_*.py"
```
