#!/usr/bin/env python3
import json
import os
import re
import subprocess
import sys
from dataclasses import dataclass
from urllib import request

DEFAULT_API_URL = "https://gopherguides.com/api/gopher-ai/review"


@dataclass
class Finding:
    severity: str
    confidence: str
    message: str


def run(cmd: list[str]) -> str:
    return subprocess.check_output(cmd, text=True)


def changed_files(base: str, head: str) -> list[str]:
    out = run(["git", "diff", "--name-only", f"{base}...{head}", "--", "*.go", "*.sql", "go.mod", "go.sum"]).strip()
    return [x.strip() for x in out.splitlines() if x.strip()]


def added_lines_only(diff: str) -> str:
    lines: list[str] = []
    for ln in diff.splitlines():
        if ln.startswith("+++") or ln.startswith("@@"):
            continue
        if ln.startswith("+"):
            lines.append(ln[1:])
    return "\n".join(lines)


def detect_heuristics(diff: str, added: str) -> list[Finding]:
    findings: list[Finding] = []

    if re.search(r"\bgo\s+func\b", added):
        findings.append(Finding("high", "medium", "Goroutine usage introduced; verify cancellation, lifetimes, and error propagation."))

    if re.search(r"\bcontext\.Background\(\)", added):
        findings.append(Finding("medium", "high", "`context.Background()` introduced; prefer request-scoped context in request paths."))

    if re.search(r"(SELECT|INSERT|UPDATE|DELETE)", added, re.I):
        findings.append(Finding("high", "medium", "SQL touched; verify parameterization, tx boundaries, and index impact."))

    if re.search(r"\bpanic\(", added):
        findings.append(Finding("high", "high", "`panic()` introduced in changed lines; confirm this cannot affect user-path behavior."))

    if re.search(r"\b(TODO|FIXME)\b", added):
        findings.append(Finding("low", "high", "TODO/FIXME in changed lines; confirm intentional and tracked."))

    if re.search(r"\btime\.Sleep\(", added):
        findings.append(Finding("medium", "medium", "`time.Sleep()` introduced; verify this is not masking race/timing issues in prod code."))

    if re.search(r"\b(err\s*:=\s*|err\s*=\s*)", added) and not re.search(r"\bif\s+err\s*!?=\s*nil", added):
        findings.append(Finding("medium", "low", "Potential new error paths; ensure all introduced errors are checked/handled."))

    return findings


def post_review(diff: str, key: str, api_url: str) -> dict:
    payload = json.dumps({"diff": diff}).encode("utf-8")
    req = request.Request(api_url, data=payload, method="POST")
    req.add_header("Authorization", f"Bearer {key}")
    req.add_header("Content-Type", "application/json")
    with request.urlopen(req, timeout=90) as resp:
        return json.loads(resp.read().decode("utf-8"))


def summarize_counts(findings: list[Finding]) -> tuple[int, int, int]:
    high = sum(1 for f in findings if f.severity == "high")
    med = sum(1 for f in findings if f.severity == "medium")
    low = sum(1 for f in findings if f.severity == "low")
    return high, med, low


def sanitize_guidance(text: str) -> str:
    """Remove noisy/irrelevant provenance lines from API output."""
    out: list[str] = []
    for line in text.splitlines():
        s = line.strip()
        if s.lower().startswith("source:"):
            continue
        # Drop simple markdown bullets that only contain source references.
        if s.startswith("-") and "source:" in s.lower():
            continue
        out.append(line)
    return "\n".join(out).strip()


def main() -> int:
    base = os.getenv("BASE_REF", "origin/main")
    head = os.getenv("HEAD_REF", "HEAD")
    out_path = os.getenv("OUT_PATH", "/tmp/agentic_review.md")
    api_url = os.getenv("REVIEW_API_URL", DEFAULT_API_URL).strip() or DEFAULT_API_URL

    diff = run(["git", "diff", f"{base}...{head}", "--", "*.go", "*.sql", "go.mod", "go.sum"]).strip()
    files = changed_files(base, head)

    if not diff:
        with open(out_path, "w") as f:
            f.write("No Go/SQL/module changes detected for agentic review.\n")
        return 0

    key = os.getenv("GOPHER_GUIDES_API_KEY", "").strip()
    if not key:
        with open(out_path, "w") as f:
            f.write("GOPHER_GUIDES_API_KEY not configured; skipping external review call.\n")
        return 0

    added = added_lines_only(diff)
    heuristics = detect_heuristics(diff, added)
    high, med, low = summarize_counts(heuristics)

    try:
        api = post_review(diff[:120000], key, api_url)
        content = sanitize_guidance((api.get("content") or ""))
    except Exception as e:
        content = f"API call failed: `{e}`"

    sections = [
        "## 🤖 Agentic Go Review",
        "",
        f"- Endpoint: `{api_url}`",
        f"- Files analyzed: **{len(files)}**",
        f"- Heuristic findings: **{len(heuristics)}** (high: {high}, medium: {med}, low: {low})",
        "",
        "### Changed files",
    ]

    if files:
        sections.extend([f"- `{f}`" for f in files[:30]])
        if len(files) > 30:
            sections.append(f"- … and {len(files) - 30} more")
    else:
        sections.append("- None")

    sections.extend(["", "### Severity/Confidence findings"])

    if heuristics:
        for f in heuristics:
            sections.append(f"- **[{f.severity.upper()} / {f.confidence.upper()}]** {f.message}")
    else:
        sections.append("- No heuristic red flags triggered.")

    sections.extend([
        "",
        "### Gopher Guides guidance",
        "",
        content[:9000] if content else "No guidance returned.",
        "",
        "---",
        "_This review is advisory. Human review and CI gates remain required._",
    ])

    with open(out_path, "w") as f:
        f.write("\n".join(sections) + "\n")

    return 0


if __name__ == "__main__":
    sys.exit(main())
