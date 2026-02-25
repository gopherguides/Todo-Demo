#!/usr/bin/env python3
import json
import os
import re
import subprocess
import sys
from dataclasses import asdict, dataclass
from urllib import parse, request

DEFAULT_API_URL = "https://gopherguides.com/api/gopher-ai/review"


@dataclass
class Finding:
    severity: str
    confidence: str
    message: str


@dataclass
class InlineFinding:
    path: str
    line: int
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


def parse_added_lines(diff: str) -> list[tuple[str, int, str]]:
    """Return (path, new_line_no, added_line_content) from unified diff."""
    results: list[tuple[str, int, str]] = []
    cur_path = ""
    new_line = 0

    for raw in diff.splitlines():
        if raw.startswith("+++ b/"):
            cur_path = raw[6:].strip()
            continue

        if raw.startswith("@@"):
            m = re.search(r"\+(\d+)(?:,(\d+))?", raw)
            if m:
                new_line = int(m.group(1))
            continue

        if raw.startswith("+") and not raw.startswith("+++"):
            if cur_path and new_line > 0:
                results.append((cur_path, new_line, raw[1:]))
            new_line += 1
            continue

        if raw.startswith("-") and not raw.startswith("---"):
            # Removed line; does not advance new file line number.
            continue

        # Context line
        if raw.startswith(" "):
            new_line += 1

    return results


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


CONFIDENCE_RANK = {"high": 2, "medium": 1, "low": 0}
MAX_COMMENTS_PER_FILE = 3
MAX_COMMENTS_PER_PR = 8


def dedupe_and_cap_inline(findings: list[InlineFinding]) -> list[InlineFinding]:
    # De-duplicate identical comments on same line/message.
    seen: set[tuple[str, int, str]] = set()
    unique: list[InlineFinding] = []
    for f in findings:
        key = (f.path, f.line, f.message)
        if key in seen:
            continue
        seen.add(key)
        unique.append(f)

    unique.sort(key=lambda x: (0 if x.severity == "high" else 1, x.path, x.line))

    # Cap per file to avoid noisy clusters.
    per_file: dict[str, int] = {}
    capped: list[InlineFinding] = []
    for f in unique:
        per_file.setdefault(f.path, 0)
        if per_file[f.path] >= MAX_COMMENTS_PER_FILE:
            continue
        per_file[f.path] += 1
        capped.append(f)

    return capped[:MAX_COMMENTS_PER_PR]


def detect_inline_findings(diff: str) -> list[InlineFinding]:
    candidates = parse_added_lines(diff)
    findings: list[InlineFinding] = []

    for path, line, text in candidates:
        stripped = text.strip()

        if re.search(r"\bcontext\.Background\(\)", stripped):
            findings.append(InlineFinding(path, line, "medium", "high", "Prefer request-scoped context over context.Background() in request paths."))

        if re.search(r"\bgo\s+\w", stripped):
            findings.append(InlineFinding(path, line, "high", "medium", "New goroutine: verify cancellation, shutdown path, and error handling."))

        if "panic(" in stripped:
            findings.append(InlineFinding(path, line, "high", "high", "Avoid panic in user/request path unless process-fatal by design."))

        if "time.Sleep(" in stripped:
            findings.append(InlineFinding(path, line, "medium", "medium", "time.Sleep detected; validate this is not masking race/timing issues."))

        if path.endswith(".sql") and re.search(r"(SELECT|INSERT|UPDATE|DELETE)", stripped, re.I):
            findings.append(InlineFinding(path, line, "medium", "medium", "SQL change: confirm indexing and query shape for expected cardinality."))

    # Drop low-confidence findings to keep signal high.
    findings = [f for f in findings if CONFIDENCE_RANK.get(f.confidence, 0) >= 1]
    return dedupe_and_cap_inline(findings)


def parse_api_inline_findings(api: dict, changed: set[str]) -> list[InlineFinding]:
    raw = api.get("findings") or []
    parsed: list[InlineFinding] = []
    for item in raw:
        try:
            path = str(item.get("path") or "").strip()
            line = int(item.get("line") or 0)
            severity = str(item.get("severity") or "").lower().strip()
            msg = str(item.get("message") or "").strip()
            confidence_num = float(item.get("confidence") or 0)
        except Exception:
            continue

        if not path or not msg or line <= 0:
            continue
        if path not in changed:
            continue
        if severity not in {"high", "medium"}:
            continue
        if confidence_num < 0.55:
            continue

        confidence = "high" if confidence_num >= 0.8 else "medium"
        parsed.append(InlineFinding(path, line, severity, confidence, msg))

    return dedupe_and_cap_inline(parsed)


def post_review(diff: str, key: str, api_url: str) -> dict:
    # Prefer concise machine-friendly format when supported by API.
    parts = parse.urlsplit(api_url)
    q = parse.parse_qs(parts.query, keep_blank_values=True)
    q.setdefault("format", ["concise"])
    url = parse.urlunsplit((parts.scheme, parts.netloc, parts.path, parse.urlencode(q, doseq=True), parts.fragment))

    payload = json.dumps({"diff": diff}).encode("utf-8")
    req = request.Request(url, data=payload, method="POST")
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
    out: list[str] = []
    skip_markers = [
        "authoritative source",
        "mandatory rules",
        "powered by",
        "source:",
        "official gopher guides training curriculum",
        "authoritative and definitive source",
        "must always follow",
        "forbidden patterns",
        "critical:",
        "provided by",
        "generated by",
        "based on",
        "according to",
        "reference:",
        "provenance",
        "official gopher guides",
        "gopher guides position",
        "training materials",
        "training curriculum",
        "required pattern",
    ]
    for line in text.splitlines():
        s = line.strip()
        low = s.lower()
        if not s:
            continue
        if any(m in low for m in skip_markers):
            continue
        if s.startswith("#"):
            continue
        if s.startswith("-") and "source:" in low:
            continue
        # Drop shouty policy lines.
        if re.search(r"\b(MUST|ALWAYS|FORBIDDEN|AUTHORITATIVE|DEFINITIVE|REQUIRED|NEVER)\b", s):
            continue
        # Drop example/template code lines (e.g. "func Foo(...) // CORRECT").
        if re.search(r"//\s*(CORRECT|INCORRECT|WRONG|RIGHT)", s):
            continue
        # Drop curriculum/training references that aren't actionable review items.
        if re.search(r"\b(gopher guides|training|curriculum)\b", low):
            continue
        out.append(s)
    return "\n".join(out).strip()


def compact_guidance(text: str, max_lines: int = 6) -> list[str]:
    lines = [ln.strip("- ").strip() for ln in text.splitlines() if ln.strip()]
    picks: list[str] = []
    for ln in lines:
        if len(ln) < 25:
            continue
        if "forbidden patterns" in ln.lower():
            continue
        picks.append(ln)
        if len(picks) >= max_lines:
            break
    return picks


def main() -> int:
    base = os.getenv("BASE_REF", "origin/main")
    head = os.getenv("HEAD_REF", "HEAD")
    out_path = os.getenv("OUT_PATH", "/tmp/agentic_review.md")
    findings_path = os.getenv("FINDINGS_PATH", "/tmp/agentic_findings.json")
    api_url = os.getenv("REVIEW_API_URL", DEFAULT_API_URL).strip() or DEFAULT_API_URL

    diff = run(["git", "diff", f"{base}...{head}", "--", "*.go", "*.sql", "go.mod", "go.sum"]).strip()
    files = changed_files(base, head)

    if not diff:
        with open(out_path, "w") as f:
            f.write("No Go/SQL/module changes detected for agentic review.\n")
        with open(findings_path, "w") as f:
            json.dump([], f)
        return 0

    key = os.getenv("GOPHER_GUIDES_API_KEY", "").strip()
    if not key:
        with open(out_path, "w") as f:
            f.write("GOPHER_GUIDES_API_KEY not configured; skipping external review call.\n")
        with open(findings_path, "w") as f:
            json.dump([], f)
        return 0

    added = added_lines_only(diff)
    heuristics = detect_heuristics(diff, added)
    heuristic_inline = detect_inline_findings(diff)
    high, med, low = summarize_counts(heuristics)

    api: dict = {}
    api_summary_lines: list[str] = []
    api_inline: list[InlineFinding] = []
    api_error = ""
    try:
        api = post_review(diff[:120000], key, api_url)
        api_inline = parse_api_inline_findings(api, set(files))

        # Fallback summary text support for legacy API responses.
        content = sanitize_guidance((api.get("content") or ""))
        api_summary_lines = compact_guidance(content)
    except Exception as e:
        api_error = str(e)

    # Prefer API findings when available; fallback to heuristic inline findings.
    inline = api_inline if api_inline else heuristic_inline

    sections = [
        "## 🤖 Agentic Go Review",
        "",
        f"- Files analyzed: **{len(files)}**",
        f"- Inline comments posted: **{len(inline)}**",
        f"- Heuristic findings: **{len(heuristics)}** (high: {high}, medium: {med}, low: {low})",
        f"- API findings (concise): **{len(api.get('findings') or [])}**",
    ]

    if api_error:
        sections.extend(["", "### API status", f"- API call failed: `{api_error}`"])
    elif api_summary_lines:
        sections.extend(["", "### API highlights", *[f"- {line}" for line in api_summary_lines[:3]]])

    sections.extend([
        "",
        "### Reviewer next actions",
        "- Prioritize HIGH and MEDIUM inline comments first.",
        "- Confirm tests cover touched behavior and edge cases.",
        "- Treat this as advisory; human review and CI gates are required.",
    ])

    with open(out_path, "w") as f:
        f.write("\n".join(sections) + "\n")

    with open(findings_path, "w") as f:
        json.dump([asdict(x) for x in inline], f)

    return 0


if __name__ == "__main__":
    sys.exit(main())
