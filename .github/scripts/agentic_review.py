#!/usr/bin/env python3
import json
import os
import re
import subprocess
import sys
from urllib import request

API_URL = "https://gopherguides.com/api/gopher-ai/review"


def run(cmd: list[str]) -> str:
    out = subprocess.check_output(cmd, text=True)
    return out


def detect_heuristics(diff: str) -> list[str]:
    findings = []
    if re.search(r"\bgo\s+func\b", diff):
        findings.append("Potential goroutine usage change; verify context cancellation and error handling.")
    if re.search(r"\bcontext\.Background\(\)", diff):
        findings.append("`context.Background()` introduced; ensure request-scoped context is preferred where appropriate.")
    if re.search(r"SELECT|INSERT|UPDATE|DELETE", diff, re.I):
        findings.append("SQL changes detected; verify parameterization and transaction boundaries.")
    if re.search(r"panic\(", diff):
        findings.append("`panic()` introduced; confirm this is intentional and not user-path behavior.")
    if re.search(r"TODO|FIXME", diff):
        findings.append("TODO/FIXME markers found in diff; verify if acceptable before merge.")
    return findings


def post_review(diff: str, key: str) -> dict:
    payload = json.dumps({"diff": diff}).encode("utf-8")
    req = request.Request(API_URL, data=payload, method="POST")
    req.add_header("Authorization", f"Bearer {key}")
    req.add_header("Content-Type", "application/json")
    with request.urlopen(req, timeout=60) as resp:
        return json.loads(resp.read().decode("utf-8"))


def main() -> int:
    base = os.getenv("BASE_REF", "origin/main")
    head = os.getenv("HEAD_REF", "HEAD")
    out_path = os.getenv("OUT_PATH", "/tmp/agentic_review.md")

    diff = run(["git", "diff", f"{base}...{head}", "--", "*.go", "*.sql", "go.mod", "go.sum"]).strip()
    if not diff:
        with open(out_path, "w") as f:
            f.write("No Go/SQL/module changes detected for agentic review.\n")
        return 0

    key = os.getenv("GOPHER_GUIDES_API_KEY", "").strip()
    if not key:
        with open(out_path, "w") as f:
            f.write("GOPHER_GUIDES_API_KEY not configured; skipping external review call.\n")
        return 0

    api = post_review(diff[:120000], key)
    content = (api.get("content") or "").strip()
    heuristics = detect_heuristics(diff)

    sections = [
        "## 🤖 Agentic Go Review (Gopher Guides API)",
        "",
        "### Heuristic checks",
    ]
    if heuristics:
        sections.extend([f"- {h}" for h in heuristics])
    else:
        sections.append("- No heuristic red flags triggered.")

    sections.extend([
        "",
        "### Gopher Guides guidance",
        "",
        content[:8000] if content else "No guidance returned.",
    ])

    with open(out_path, "w") as f:
        f.write("\n".join(sections) + "\n")

    return 0


if __name__ == "__main__":
    sys.exit(main())
