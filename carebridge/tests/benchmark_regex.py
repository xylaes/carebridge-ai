import time
import sys
from carebridge.analyzer import CareBridgeAnalyzer

def main():
    analyzer = CareBridgeAnalyzer(use_mock=True)
    sample_voice_note = (
        "Finished 4-hour shift with Mrs. Eleanor. Blood pressure was 120/80, pulse 72. "
        "Administered 5mg Lisinopril at 9 AM with water. She ate 80% of her oatmeal breakfast. "
        "Assisted with 15-minute walker gait exercise in garden. "
        "She reported mild left knee stiffness."
    )

    iterations = 50000
    print(f"Running {iterations} iterations of regex-intensive tasks...")

    # Warmup
    for _ in range(100):
        metrics = analyzer._heuristic_extract_clinical_metrics(sample_voice_note)
        analyzer.generate_family_summary(sample_voice_note, metrics)
        analyzer.generate_billing_note(sample_voice_note, metrics)

    start_time = time.perf_counter()
    for _ in range(iterations):
        metrics = analyzer._heuristic_extract_clinical_metrics(sample_voice_note)
        analyzer.generate_family_summary(sample_voice_note, metrics)
        analyzer.generate_billing_note(sample_voice_note, metrics)
    end_time = time.perf_counter()

    elapsed = end_time - start_time
    print(f"Total time for {iterations} iterations: {elapsed:.6f} seconds")
    print(f"Average time per iteration: {elapsed / iterations * 1000000:.3f} microseconds")

if __name__ == "__main__":
    main()
