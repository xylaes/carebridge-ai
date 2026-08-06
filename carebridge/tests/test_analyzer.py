import unittest
from unittest.mock import patch, MagicMock

# Import will be tested once analyzer module is defined
from carebridge.analyzer import (
    CareBridgeAnalyzer,
    ClinicalMetrics,
    CareShiftReport
)

class TestCareBridgeAnalyzer(unittest.TestCase):

    def setUp(self):
        self.analyzer = CareBridgeAnalyzer(use_mock=True)
        self.sample_voice_note = (
            "Finished 4-hour shift with Mrs. Eleanor. Blood pressure was 120/80, pulse 72. "
            "Administered 5mg Lisinopril at 9 AM with water. She ate 80% of her oatmeal breakfast. "
            "Assisted with 15-minute walker gait exercise in garden. "
            "She reported mild left knee stiffness."
        )

    def test_clean_transcript(self):
        result = self.analyzer.transcribe_and_clean_audio(self.sample_voice_note)
        self.assertIsInstance(result, str)
        self.assertIn("Mrs. Eleanor", result)

    def test_extract_clinical_metrics(self):
        metrics = self.analyzer.extract_clinical_metrics(self.sample_voice_note)
        self.assertIsInstance(metrics, ClinicalMetrics)
        self.assertEqual(metrics.vitals.get("blood_pressure"), "120/80")
        self.assertEqual(metrics.vitals.get("pulse"), 72)
        self.assertTrue(len(metrics.medications) > 0)
        self.assertIn("walker", metrics.mobility.lower())
        self.assertIn("oatmeal", metrics.nutrition.lower())

    def test_generate_family_summary(self):
        metrics = self.analyzer.extract_clinical_metrics(self.sample_voice_note)
        summary = self.analyzer.generate_family_summary(self.sample_voice_note, metrics)
        self.assertIsInstance(summary, str)
        self.assertIn("Eleanor", summary)

    def test_generate_billing_note(self):
        metrics = self.analyzer.extract_clinical_metrics(self.sample_voice_note)
        billing = self.analyzer.generate_billing_note(self.sample_voice_note, metrics)
        self.assertIsInstance(billing, str)
        self.assertIn("Medicaid", billing)
        self.assertIn("4-hour", billing)

    def test_full_analyze_shift_note(self):
        report = self.analyzer.analyze_shift_note(self.sample_voice_note)
        self.assertIsInstance(report, CareShiftReport)
        self.assertIsNotNone(report.raw_transcript)
        self.assertIsNotNone(report.clinical_log)
        self.assertTrue(len(report.family_summary) > 0)
        self.assertTrue(len(report.billing_note) > 0)

    def test_heuristic_extract_vitals(self):
        vitals = self.analyzer._heuristic_extract_vitals(self.sample_voice_note)
        self.assertEqual(vitals.get("blood_pressure"), "120/80")
        self.assertEqual(vitals.get("pulse"), 72)

        # Test empty input
        self.assertEqual(self.analyzer._heuristic_extract_vitals(""), {})

    def test_heuristic_extract_medications(self):
        meds = self.analyzer._heuristic_extract_medications(self.sample_voice_note)
        self.assertEqual(len(meds), 1)
        self.assertEqual(meds[0]["name"], "Lisinopril")
        self.assertEqual(meds[0]["dosage"], "5mg")
        self.assertEqual(meds[0]["status"], "Administered")

        # Test fallback
        fallback_meds = self.analyzer._heuristic_extract_medications("")
        self.assertEqual(len(fallback_meds), 1)
        self.assertIn("note", fallback_meds[0])

    def test_heuristic_extract_mobility(self):
        # Walker
        mobility = self.analyzer._heuristic_extract_mobility("Using walker today")
        self.assertIn("walker", mobility)

        # Wheelchair
        mobility = self.analyzer._heuristic_extract_mobility("assisted into wheelchair")
        self.assertIn("Wheelchair", mobility)

        # Walk/gait
        mobility = self.analyzer._heuristic_extract_mobility("went for a walk")
        self.assertIn("Ambulatory", mobility)

        # Unspecified
        mobility = self.analyzer._heuristic_extract_mobility("")
        self.assertEqual(mobility, "Unspecified mobility")

    def test_heuristic_extract_nutrition(self):
        # Oatmeal
        nutrition = self.analyzer._heuristic_extract_nutrition("had oatmeal for breakfast")
        self.assertIn("oatmeal", nutrition)

        # Breakfast/eat/%
        nutrition = self.analyzer._heuristic_extract_nutrition("finished 70% of breakfast")
        self.assertIn("breakfast", nutrition)

        # Fallback
        nutrition = self.analyzer._heuristic_extract_nutrition("")
        self.assertEqual(nutrition, "Tolerated meal well.")

    def test_heuristic_extract_alerts(self):
        # With alert keyword
        alerts = self.analyzer._heuristic_extract_alerts("reported pain in hip")
        self.assertEqual(len(alerts), 1)
        self.assertIn("stiffness", alerts[0])

        # No alerts
        alerts = self.analyzer._heuristic_extract_alerts("feeling fine")
        self.assertEqual(len(alerts), 0)

    @patch('carebridge.analyzer.genai.Client')
    def test_extract_clinical_metrics_non_mock_success(self, mock_client_class):
        # Setup mock client
        mock_client = MagicMock()
        mock_client_class.return_value = mock_client

        # Setup mock response with valid JSON
        mock_response = MagicMock()
        mock_response.text = (
            '{"vitals": {"blood_pressure": "110/70", "pulse": 68}, '
            '"medications": [{"name": "Aspirin", "dosage": "81mg", "time": "8 AM"}], '
            '"mobility": "Assisted walks", "nutrition": "Ate 100%", '
            '"alerts": ["Some alert"]}'
        )
        mock_client.models.generate_content.return_value = mock_response

        # Initialize analyzer with non-mock setting
        analyzer = CareBridgeAnalyzer(api_key="fake-key", use_mock=False)
        self.assertFalse(analyzer.use_mock)
        self.assertEqual(analyzer.client, mock_client)

        # Call the method
        metrics = analyzer.extract_clinical_metrics(self.sample_voice_note)

        # Verify call arguments
        mock_client.models.generate_content.assert_called_once()
        args, kwargs = mock_client.models.generate_content.call_args
        self.assertEqual(kwargs.get('model'), 'gemini-2.5-flash')
        self.assertIn(self.sample_voice_note, kwargs.get('contents'))
        self.assertEqual(kwargs.get('config').response_mime_type, "application/json")

        # Verify parsed clinical metrics
        self.assertIsInstance(metrics, ClinicalMetrics)
        self.assertEqual(metrics.vitals, {"blood_pressure": "110/70", "pulse": 68})
        self.assertEqual(metrics.medications, [{"name": "Aspirin", "dosage": "81mg", "time": "8 AM"}])
        self.assertEqual(metrics.mobility, "Assisted walks")
        self.assertEqual(metrics.nutrition, "Ate 100%")
        self.assertEqual(metrics.alerts, ["Some alert"])

    @patch('carebridge.analyzer.genai.Client')
    def test_extract_clinical_metrics_non_mock_malformed_json_fallback(self, mock_client_class):
        # Setup mock client
        mock_client = MagicMock()
        mock_client_class.return_value = mock_client

        # Setup mock response with malformed/invalid JSON
        mock_response = MagicMock()
        mock_response.text = '{"vitals": {"blood_pressure": "110/70", ' # Unclosed JSON
        mock_client.models.generate_content.return_value = mock_response

        # Initialize analyzer
        analyzer = CareBridgeAnalyzer(api_key="fake-key", use_mock=False)

        # Call the method
        metrics = analyzer.extract_clinical_metrics(self.sample_voice_note)

        # Verify call was made
        mock_client.models.generate_content.assert_called_once()

        # Verify fallback to heuristic extraction occurred
        self.assertIsInstance(metrics, ClinicalMetrics)
        self.assertEqual(metrics.vitals.get("blood_pressure"), "120/80")
        self.assertEqual(metrics.vitals.get("pulse"), 72)
        self.assertTrue(len(metrics.medications) > 0)
        self.assertIn("walker", metrics.mobility.lower())
        self.assertIn("oatmeal", metrics.nutrition.lower())

    @patch('carebridge.analyzer.genai.Client')
    def test_extract_clinical_metrics_non_mock_api_error_fallback(self, mock_client_class):
        # Setup mock client
        mock_client = MagicMock()
        mock_client_class.return_value = mock_client

        # Setup mock response to raise an exception
        mock_client.models.generate_content.side_effect = Exception("Google GenAI API connection failed")

        # Initialize analyzer
        analyzer = CareBridgeAnalyzer(api_key="fake-key", use_mock=False)

        # Call the method
        metrics = analyzer.extract_clinical_metrics(self.sample_voice_note)

        # Verify fallback to heuristic extraction occurred
        self.assertIsInstance(metrics, ClinicalMetrics)
        self.assertEqual(metrics.vitals.get("blood_pressure"), "120/80")
        self.assertEqual(metrics.vitals.get("pulse"), 72)

if __name__ == "__main__":
    unittest.main()
