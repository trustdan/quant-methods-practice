"""Check backlog objectives against approved files and matching review records."""
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def main():
    text = (ROOT / "docs/CONTENT-BACKLOG.md").read_text(encoding="utf-8")
    # Only the objective checklist defines the catalog (not wave-summary tables).
    items = text.split("## Question items", 1)[1].split("## Open course conventions", 1)[0]
    current = None
    rows = {}
    seen = set()
    for line in items.splitlines():
        heading = re.match(r"#### (\d+\.\d+) ", line)
        if heading:
            current = heading.group(1)
            rows[current] = []
        item = re.match(r"- \[[ x]\].*?`([a-z][a-z0-9-]+)`", line)
        if not item:
            continue
        slug = item.group(1)
        assert current, f"Question outside objective: {slug}"
        assert slug not in seen, f"Duplicate backlog item: {slug}"
        seen.add(slug)
        path = ROOT / "curriculum/approved" / (slug + ".json")
        q = json.loads(path.read_text(encoding="utf-8"))
        assert q["status"] == "active" and q["approval"], f"Unapproved: {slug}"
        assert q["id"] == slug.replace("-", "_"), f"ID mismatch: {slug}"
        review = json.loads((ROOT / "curriculum/reviews" / (slug + ".review.json")).read_text(encoding="utf-8"))
        # Older immutable reviews use status/verdict and omit version. They
        # apply only to their original v1 template, never to a later revision.
        version = review.get("template_version", 1)
        status = review.get("review_status", review.get("status", review.get("verdict")))
        assert (review["template_id"], version, status) == (q["id"], q["version"], "approved"), f"Review mismatch: {slug}"
        assert review["reviewer"] == q["approval"]["reviewer"], f"Reviewer mismatch: {slug}"
        rows[current].append(slug)
    assert len(rows) == 22, f"Expected 22 objectives, got {len(rows)}"
    assert len(seen) == 121, f"Expected 121 backlog questions, got {len(seen)}"
    for row, slugs in rows.items():
        assert len(slugs) >= 5, f"Insufficient breadth: {row}"
        status_line = next(line for line in text.splitlines() if line.startswith(f"| {row} |"))
        completed, target = map(int, status_line.split("|")[-2].strip().split("/"))
        assert completed == len(slugs) == target, f"Status table mismatch for {row}"
        print(f"{row}: {len(slugs)} approved catalog questions")
    bank = [json.loads(p.read_text(encoding="utf-8")) for p in (ROOT / "curriculum/approved").glob("*.json")]
    assert len(bank) == 137 and len({q["id"] for q in bank}) == 137
    print("Coverage OK: 121/121 backlog questions, 22/22 objectives, 137 unique active bank questions.")


if __name__ == "__main__":
    main()
