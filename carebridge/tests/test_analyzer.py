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

if __name__ == "__main__":
    unittest.main()
