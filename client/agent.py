import mcp_client as mcp
from doc_loader import load_docs, get_relevant_context
from scraper import scrape_form, FormField
from filler import fill_form


def build_search_query(question: str, options: list[str]) -> str:
    if options:
        return f"{question} {' '.join(options[:2])}"
    return question


def answer_from_personal(question: str, options: list[str], context: str) -> str:
    relevant = get_relevant_context(context, question)
    if not relevant:
        return ""
    return mcp.ask_llm(question, context=relevant, options=options)


def answer_from_web(question: str, options: list[str]) -> str:
    query = build_search_query(question, options)
    results = mcp.search(query, max_results=3)
    if not results:
        return ""
    web_context = "\n".join(
        f"{r['title']}: {r['snippet']}" for r in results if r.get("title")
    )
    return mcp.ask_llm(question, context=web_context, options=options)


def resolve_answer(field: FormField, personal_context: str) -> str | list[str]:
    print(f"  [personal] querying...")
    ans_personal = answer_from_personal(field.question, field.options, personal_context)

    print(f"  [web]      querying...")
    ans_web = answer_from_web(field.question, field.options)

    if not ans_personal and not ans_web:
        return field.options[0] if field.options else ""

    if not ans_personal:
        print(f"  [decision] only web answer available")
        return ans_web

    if not ans_web:
        print(f"  [decision] only personal answer available")
        return ans_personal

    print(f"  [rank]     personal='{ans_personal}' | web='{ans_web}'")
    ranked = mcp.rank_answers(field.question, ans_personal, ans_web)
    winner = ranked.get("winner", ans_personal)
    source = ranked.get("source", "unknown")
    print(f"  [decision] winner='{winner}' source={source}")

    if field.field_type == "checkbox":
        return [winner]
    return winner


def run(url: str, docs_dir: str = "../docs"):
    print(f"loading personal docs from {docs_dir}")
    personal_context = load_docs(docs_dir)

    print(f"scraping form: {url}")
    fields = scrape_form(url)
    print(f"found {len(fields)} fields across pages\n")

    page_fields: dict[int, list[tuple[FormField, str | list[str]]]] = {}

    for field in fields:
        print(f"[field {field.index}] page={field.page} type={field.field_type}")
        print(f"  Q: {field.question}")
        if field.options:
            print(f"  options: {field.options}")

        answer = resolve_answer(field, personal_context)
        print(f"  => answer: {answer}\n")

        if field.page not in page_fields:
            page_fields[field.page] = []
        page_fields[field.page].append((field, answer))

    print("filling form...")
    fill_form(url, page_fields)
    print("done")


if __name__ == "__main__":
    import sys
    if len(sys.argv) < 2:
        print("usage: uv run python agent.py <google_form_url>")
        sys.exit(1)
    run(sys.argv[1])