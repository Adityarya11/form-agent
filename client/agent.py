import mcp_client as mcp
from doc_loader import load_docs, get_relevant_context
from scraper import scrape_form, FormField
from filler import fill_form


def run(url: str, docs_dir: str = "../docs", concurrency: int = 2, dry_run: bool = False,use_profile: bool = False, 
        profile_path: str = "", use_cdp: bool = False):
    
    print(f"loading personal docs from {docs_dir}")
    personal_context = load_docs(docs_dir)

    print(f"scraping form: {url}")
    fields = scrape_form(url, use_cdp=use_cdp)
    print(f"found {len(fields)} fields across pages\n")

    jobs = [
        {
            "Question": f.question,
            "Options":  f.options,
            "Context":  get_relevant_context(personal_context, f.question),
        }
        for f in fields
    ]

    print(f"resolving {len(jobs)} fields (batch={concurrency})...")
    results = mcp.resolve_batch(jobs)

    answer_map = {r["Question"]: r["Winner"] for r in results}

    for r in results:
        print(f"Q: {r['Question']}")
        print(f"   winner='{r['Winner']}' source={r['Source']}\n")

    if dry_run:
        print("dry-run: skipping fill")
        return

    page_fields: dict[int, list[tuple[FormField, str | list[str]]]] = {}
    for field in fields:
        answer = answer_map.get(field.question, "")
        if field.page not in page_fields:
            page_fields[field.page] = []
        page_fields[field.page].append((field, answer))

    print("filling form...")
    fill_form(url, page_fields, use_profile=use_profile, profile_path=profile_path, use_cdp=use_cdp)
    print("done")