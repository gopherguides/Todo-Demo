#!/usr/bin/env python3
import unittest

from agentic_review import (
    InlineFinding,
    detect_inline_findings,
    sanitize_guidance,
    compact_guidance,
    MAX_COMMENTS_PER_FILE,
    MAX_COMMENTS_PER_PR,
)


def _make_diff(path: str, lines: list[str], start: int = 1) -> str:
    header = f"diff --git a/{path} b/{path}\n--- a/{path}\n+++ b/{path}\n"
    hunk = f"@@ -1,0 +{start},{len(lines)} @@\n"
    body = "\n".join(f"+{ln}" for ln in lines)
    return header + hunk + body + "\n"


class TestConfidenceThreshold(unittest.TestCase):
    def test_low_confidence_filtered(self):
        diff = _make_diff("main.go", [
            'err := doStuff()',
        ])
        findings = detect_inline_findings(diff)
        for f in findings:
            self.assertNotEqual(f.confidence, "low",
                                "low-confidence findings should be filtered")

    def test_medium_confidence_kept(self):
        diff = _make_diff("main.go", ['go worker(ctx)'])
        findings = detect_inline_findings(diff)
        self.assertTrue(len(findings) > 0, "medium-confidence goroutine finding should be kept")

    def test_high_confidence_kept(self):
        diff = _make_diff("main.go", ['panic("fatal")'])
        findings = detect_inline_findings(diff)
        self.assertTrue(len(findings) > 0, "high-confidence panic finding should be kept")


class TestPerFileCap(unittest.TestCase):
    def test_caps_at_limit(self):
        lines = [f'go handler{i}(ctx)' for i in range(10)]
        diff = _make_diff("server.go", lines)
        findings = detect_inline_findings(diff)
        per_file = sum(1 for f in findings if f.path == "server.go")
        self.assertLessEqual(per_file, MAX_COMMENTS_PER_FILE)


class TestPRWideCap(unittest.TestCase):
    def test_total_cap(self):
        diffs = []
        for i in range(5):
            lines = [f'go worker{j}(ctx)' for j in range(5)]
            diffs.append(_make_diff(f"pkg{i}/server.go", lines))
        combined = "\n".join(diffs)
        findings = detect_inline_findings(combined)
        self.assertLessEqual(len(findings), MAX_COMMENTS_PER_PR)


class TestSanitizeGuidance(unittest.TestCase):
    def test_strips_provenance(self):
        text = "Good advice here.\nSource: some model\nProvenance info.\nMore good advice."
        result = sanitize_guidance(text)
        self.assertNotIn("Source:", result)
        self.assertNotIn("Provenance", result)
        self.assertIn("Good advice", result)

    def test_strips_policy_boilerplate(self):
        text = "You MUST ALWAYS do X.\nPrefer interfaces for testability."
        result = sanitize_guidance(text)
        self.assertNotIn("MUST", result)
        self.assertIn("Prefer interfaces", result)


class TestCompactGuidance(unittest.TestCase):
    def test_skips_short_lines(self):
        text = "OK\nThis is a longer actionable suggestion for the reviewer."
        result = compact_guidance(text)
        self.assertEqual(len(result), 1)
        self.assertIn("actionable", result[0])

    def test_max_lines_cap(self):
        text = "\n".join([f"Line {i}: this is long enough to pass" for i in range(20)])
        result = compact_guidance(text, max_lines=6)
        self.assertLessEqual(len(result), 6)


if __name__ == "__main__":
    unittest.main()
