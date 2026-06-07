from dataclasses import dataclass, field
from playwright.sync_api import sync_playwright, Page


@dataclass
class FormField:
    index: int
    question: str
    field_type: str
    page: int
    options: list[str] = field(default_factory=list)
    required: bool = False


def scrape_form(url: str) -> list[FormField]:
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        page.goto(url, wait_until="networkidle")

        all_fields = []
        field_index = 0
        page_number = 0

        while True:
            page.wait_for_selector("div[role='listitem']", timeout=15000)
            fields, field_index = scrape_page(page, page_number, field_index)
            all_fields.extend(fields)

            if not navigate_next(page):
                break
            page_number += 1

        browser.close()
        return all_fields


def scrape_page(page: Page, page_number: int, start_index: int) -> tuple[list[FormField], int]:
    raw_items = page.query_selector_all("div[role='listitem']")
    fields = []
    idx = start_index

    for item in raw_items:
        question_el = item.query_selector("div[role='heading']")
        if not question_el:
            continue

        question_text = question_el.inner_text().strip()
        required = "*" in (item.inner_text() or "")
        field_type, options = detect_field(item)

        if field_type is None:
            continue

        fields.append(FormField(
            index=idx,
            question=question_text,
            field_type=field_type,
            options=options,
            required=required,
            page=page_number,
        ))
        idx += 1

    return fields, idx


def navigate_next(page: Page) -> bool:
    next_selectors = [
        "div[role='button'] span:has-text('Next')",
        "span:has-text('Next')",
    ]

    for sel in next_selectors:
        btn = page.query_selector(sel)
        if btn:
            btn.click()
            page.wait_for_load_state("networkidle")
            try:
                page.wait_for_selector("div[role='listitem']", timeout=5000)
                return True
            except Exception:
                return False

    return False


def detect_field(item) -> tuple[str | None, list[str]]:
    if item.query_selector("input[type='text'], input[type='email'], input[type='number'], input[type='url']"):
        return "short_text", []

    if item.query_selector("textarea"):
        return "long_text", []

    radio_els = item.query_selector_all("div[role='radio']")
    if radio_els:
        options = [el.get_attribute("data-value") or el.inner_text().strip() for el in radio_els]
        return "radio", [o for o in options if o]

    checkbox_els = item.query_selector_all("div[role='checkbox']")
    if checkbox_els:
        options = [el.get_attribute("data-value") or el.inner_text().strip() for el in checkbox_els]
        return "checkbox", [o for o in options if o]

    dropdown_el = item.query_selector("div[role='listbox']")
    if dropdown_el:
        option_els = item.query_selector_all("div[role='option']")
        return "dropdown", [el.inner_text().strip() for el in option_els]

    return None, []


def print_fields(fields: list[FormField]):
    current_page = -1
    for f in fields:
        if f.page != current_page:
            current_page = f.page
            print(f"\n=== Page {f.page} ===")
        print(f"[{f.index}] {f.field_type.upper()} | required={f.required}")
        print(f"     Q: {f.question}")
        if f.options:
            print(f"     Options: {f.options}")