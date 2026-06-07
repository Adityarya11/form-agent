from pathlib import Path


def load_docs(docs_dir: str = "../docs") -> str:
    base = Path(docs_dir)
    if not base.exists():
        return ""

    chunks = []
    for path in sorted(base.iterdir()):
        if path.suffix in {".md", ".txt"} and path.is_file():
            text = path.read_text(encoding="utf-8").strip()
            if text:
                chunks.append(f"=== {path.name} ===\n{text}")

    return "\n\n".join(chunks)


def get_relevant_context(full_context: str, question: str, max_chars: int = 1500) -> str:
    if not full_context:
        return ""

    question_words = set(question.lower().split())
    lines = full_context.splitlines()

    scored = []
    for i, line in enumerate(lines):
        line_words = set(line.lower().split())
        score = len(question_words & line_words)
        scored.append((score, i, line))

    scored.sort(key=lambda x: x[0], reverse=True)

    selected_indices = sorted(idx for _, idx, _ in scored[:10])
    selected_lines = [lines[i] for i in selected_indices]
    result = "\n".join(selected_lines)

    return result[:max_chars]