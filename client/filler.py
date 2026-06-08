from playwright.sync_api import sync_playwright, Page
from scraper import FormField


def fill_form(url: str, answers: dict[int, str | list[str]]):
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=False)
        page = browser.new_page()
        page.goto(url, wait_until="networkidle")

        current_page = 0
        page_fields = group_by_page(answers)

        while True:
            if current_page in page_fields:
                fill_page(page, page_fields[current_page])

            if not click_next(page):
                break
            current_page += 1

        submit(page)
        browser.close()


def group_by_page(answers: dict[int, str | list[str]]) -> dict[int, dict]:
    return answers


def fill_page(page: Page, fields: list[tuple[FormField, str | list[str]]]):
    for form_field, answer in fields:
        try:
            fill_field(page, form_field, answer)
        except Exception as e:
            print(f"failed to fill field [{form_field.index}]: {e}")


def fill_field(page: Page, form_field: FormField, answer: str | list[str]):
    items = page.query_selector_all("div[role='listitem']")

    target = None
    for item in items:
        heading = item.query_selector("div[role='heading']")
        if heading and heading.inner_text().strip() == form_field.question:
            target = item
            break

    if target is None:
        print(f"question not found on page: {form_field.question}")
        return

    match form_field.field_type:
        case "short_text" | "long_text":
            inp = target.query_selector("input, textarea")
            if inp:
                inp.click()
                inp.fill(str(answer))

        case "radio":
            radios = target.query_selector_all("div[role='radio']")
            for radio in radios:
                label = radio.get_attribute("data-value") or radio.inner_text().strip()
                if label.strip().lower() == str(answer).strip().lower():
                    radio.click()
                    break

        case "checkbox":
            selected = [a.strip().lower() for a in (answer if isinstance(answer, list) else [answer])]
            boxes = target.query_selector_all("div[role='checkbox']")
            for box in boxes:
                label = box.get_attribute("data-value") or box.inner_text().strip()
                if label.strip().lower() in selected:
                    box.click()

        case "dropdown":
            listbox = target.query_selector("div[role='listbox']")
            if listbox:
                listbox.click()
                page.wait_for_selector("div[role='option']", timeout=3000)
                options = page.query_selector_all("div[role='option']")
                for opt in options:
                    if opt.inner_text().strip().lower() == str(answer).strip().lower():
                        opt.click()
                        break


def click_next(page: Page) -> bool:
    for sel in ["div[role='button'] span:has-text('Next')", "span:has-text('Next')"]:
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


def submit(page: Page):
    for sel in ["div[role='button'] span:has-text('Submit')", "span:has-text('Submit')"]:
        btn = page.query_selector(sel)
        if btn:
            btn.click()
            page.wait_for_load_state("networkidle")
            print("form submitted")
            return
    print("submit button not found")