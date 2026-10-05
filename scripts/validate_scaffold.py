"""Check move-ready docs and draft basics; not a full JSON Schema/app validator."""

import json
import math
import re
from pathlib import Path
from urllib.parse import unquote


ROOT = Path(__file__).resolve().parents[1]
REQUIRED_FILES = [
    "README.md", "AGENTS.md", "CLAUDE.md", "OVERVIEW.md", "REQUIREMENTS.md",
    "PLAN.md", "CONTRIBUTING.md", ".gitignore", ".editorconfig",
    "docs/AGENT-CONTRACT.md", "docs/HANDOFF.md", "docs/ARCHITECTURE.md",
    "docs/BOOTSTRAP.md", "docs/DIRECTORY-STRUCTURE.md", "docs/FEATURE-PARITY.md",
    "docs/COURSE-MAP.md", "docs/PEDAGOGY.md", "docs/NUMERICS.md",
    "docs/NAVIGATION.md", "docs/STORAGE.md", "docs/MASTERY.md",
    "docs/QUESTION-BANK.md", "docs/API.md", "docs/AI-TUTOR.md",
    "docs/PROVIDERS.md", "docs/MATH-AND-VISUALS.md", "docs/ARCADE.md",
    "docs/SECURITY.md", "docs/TESTING.md", "docs/RELEASE.md",
    "docs/DECISIONS.md", "docs/SOURCES.md", "course-materials/README.md",
    "schemas/question-template.schema.json", "schemas/saved-explanation.schema.json",
    "curriculum/examples/binomial-seven-stage.draft.json",
]
REQUIRED_DIRS = [
    "cmd/quant-practice", "curriculum/approved", "curriculum/reviews",
    "course-materials/private", ".github/workflows", "tests/fixtures", "tests/e2e",
] + ["internal/" + name for name in (
    "app", "auth", "assets", "bank", "domain", "drill", "exam", "httpapi",
    "mastery", "mathengine", "providers", "storage", "tutor",
)] + ["web/src/" + name for name in (
    "app", "components", "navigation", "visuals", "arcade",
)] + ["web/src/features/" + name for name in (
    "practice", "exams", "mastery", "tutor", "notes", "providers",
    "candidates", "reference", "cases",
)] + ["web/public"]


def main():
    errors = []
    for name in REQUIRED_FILES:
        if not (ROOT / name).is_file():
            errors.append("Missing file: " + name)
    for name in REQUIRED_DIRS:
        if not (ROOT / name).is_dir():
            errors.append("Missing directory: " + name)

    ignored_parts = {"node_modules", ".git", "dist", "bin", ".system_generated"}
    markdown_files = sorted(p for p in ROOT.rglob("*.md") if not any(part in p.parts for part in ignored_parts))
    local_links = 0
    for path in markdown_files:
        content = path.read_text(encoding="utf-8")
        if "C:/Users/" in content or "C:\\Users\\" in content:
            errors.append(f"Machine-specific path: {path.relative_to(ROOT)}")
        for target in re.findall(r"\[[^\]]+\]\(([^)]+)\)", content):
            target = target.strip().strip("<>")
            if re.match(r"^[a-zA-Z][a-zA-Z0-9+.-]*:", target) or target.startswith("#"):
                continue
            target = unquote(target.split("#", 1)[0])
            resolved = (path.parent / target).resolve()
            if not resolved.is_relative_to(ROOT):
                errors.append(f"Outside-root link: {path.relative_to(ROOT)} -> {target}")
            elif not resolved.exists():
                errors.append(f"Broken link: {path.relative_to(ROOT)} -> {target}")
            local_links += 1

    json_files = sorted(p for p in ROOT.rglob("*.json") if not any(part in p.parts for part in ignored_parts))
    parsed = {}
    for path in json_files:
        try:
            parsed[path.relative_to(ROOT).as_posix()] = json.loads(
                path.read_text(encoding="utf-8"),
                parse_constant=lambda value: (_ for _ in ()).throw(ValueError(value)),
            )
        except (ValueError, OSError) as exc:
            errors.append(f"Invalid JSON: {path.relative_to(ROOT)}: {exc}")

    fixture = parsed.get("curriculum/examples/binomial-seven-stage.draft.json")
    if fixture:
        if fixture.get("status") != "draft" or fixture.get("approval") is not None:
            errors.append("Example must remain unapproved draft")
        stages = fixture.get("stages", [])
        if len(stages) != 7 or len({s["id"] for s in stages}) != 7:
            errors.append("Draft must have seven distinct stages")
        evidence = []
        for stage in stages:
            options = stage["options"]
            ids = [option["id"] for option in options]
            if len(ids) != len(set(ids)) or len(ids) > 4:
                errors.append("Invalid options: " + stage["id"])
            expected = stage["expected_answer"]
            if stage["kind"] != expected["kind"]:
                errors.append("Answer kind mismatch: " + stage["id"])
            if expected["kind"] == "choice" and expected["option_id"] not in ids:
                errors.append("Missing expected choice: " + stage["id"])
            if expected["kind"] == "numeric":
                params = fixture["parameters"]
                value = math.comb(params["n"], params["k"]) * params["p"] ** params["k"] * (1 - params["p"]) ** (params["n"] - params["k"])
                if expected["value"] != value:
                    errors.append("Draft binomial expectation disagrees with exact small fixture")
                if options or stage["numeric_policy"] is None:
                    errors.append("Invalid numeric stage shape")
            evidence.extend(stage["evidence_concept_ids"])
        if len(evidence) != len(set(evidence)) or set(evidence) != set(fixture["concept_ids"]):
            errors.append("Draft evidence concepts must each have one designated stage")

    if errors:
        print("Scaffold validation FAILED:\n" + "\n".join(errors))
        return 1
    print(f"Scaffold OK: {len(markdown_files)} Markdown files, {local_links} local links, "
          f"{len(json_files)} JSON files, {len(REQUIRED_DIRS)} source directories; draft basics checked.")
    print("No full JSON Schema, application, provider, browser, or release validation performed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
