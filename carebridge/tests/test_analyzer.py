import unittest
from unittest.mock import patch, MagicMock

# Import will be tested once analyzer module is defined
from carebridge.analyzer import CareBridgeAnalyzer, ClinicalMetrics, CareShiftReport


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
        nutrition = self.analyzer._heuristic_extract_nutrition(
            "had oatmeal for breakfast"
        )
        self.assertIn("oatmeal", nutrition)

        # Breakfast/eat/%
        nutrition = self.analyzer._heuristic_extract_nutrition(
            "finished 70% of breakfast"
        )
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

    def test_heuristic_family_summary(self):
        # Test with named person
        summary = self.analyzer._heuristic_family_summary("Shift report for Mrs. Smith")
        self.assertIn("Smith", summary)

        # Test without named person
        summary_no_name = self.analyzer._heuristic_family_summary(
            "Shift report with no name"
        )
        self.assertIn("your loved one", summary_no_name)

    def test_generate_family_summary_fallback(self):
        # Setup analyzer to not use mock, but mock the client so generate_content raises Exception
        analyzer = CareBridgeAnalyzer(use_mock=False)
        mock_client = MagicMock()
        mock_client.models.generate_content.side_effect = Exception("API Error")
        analyzer.client = mock_client
        analyzer.use_mock = False

        metrics = ClinicalMetrics()
        # Should fallback gracefully to _heuristic_family_summary
        summary = analyzer.generate_family_summary(
            "Shift report for Mrs. Eleanor", metrics
        )
        self.assertIn("Eleanor", summary)
        mock_client.models.generate_content.assert_called_once()


if __name__ == "__main__":
    unittest.main()
