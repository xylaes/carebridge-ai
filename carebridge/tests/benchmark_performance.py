import time
import sys
from carebridge.analyzer import CareBridgeAnalyzer, ClinicalMetrics

def main():
    analyzer = CareBridgeAnalyzer(use_mock=True)
    sample_voice_note = (
        "Finished 4-hour shift with Mrs. Eleanor. Blood pressure was 120/80, pulse 72. "
        "Administered 5mg Lisinopril at 9 AM with water. She ate 80% of her oatmeal breakfast. "
        "Assisted with 15-minute walker gait exercise in garden. "
        "She reported mild left knee stiffness."
    )

    # Save original methods
    orig_transcribe = analyzer.transcribe_and_clean_audio
    orig_extract = analyzer.extract_clinical_metrics
    orig_summary = analyzer.generate_family_summary

    # Create wrapped methods with 1 second delay
    def mock_transcribe(*args, **kwargs):
        time.sleep(1.0)
        return orig_transcribe(*args, **kwargs)

    def mock_extract(*args, **kwargs):
        time.sleep(1.0)
        return orig_extract(*args, **kwargs)

    def mock_summary(*args, **kwargs):
        time.sleep(1.0)
        return orig_summary(*args, **kwargs)

    # Monkeypatch the analyzer instance
    analyzer.transcribe_and_clean_audio = mock_transcribe
    analyzer.extract_clinical_metrics = mock_extract
    analyzer.generate_family_summary = mock_summary

    print("Running performance benchmark on analyze_shift_note...")
    start_time = time.perf_counter()
    report = analyzer.analyze_shift_note(sample_voice_note)
    end_time = time.perf_counter()

    elapsed = end_time - start_time
    print(f"Elapsed time: {elapsed:.4f} seconds")

    # Assert correctness
    assert report.cleaned_transcript is not None, "Report transcription is None"
    assert report.clinical_log is not None, "Report clinical log is None"
    assert report.family_summary is not None, "Report family summary is None"
    assert report.billing_note is not None, "Report billing note is None"
    print("Benchmark completed successfully.")

if __name__ == "__main__":
    main()
