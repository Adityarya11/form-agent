from playwright.sync_api import sync_playwright, Page, BrowserContext
from scraper import FormField


def fill_form(
    url: str,
    page_fields: dict[int, list[tuple[FormField, str | list[str]]]],
    use_profile: bool = False,
    profile_path: str = "",
    use_cdp: bool = False,
):
    with sync_playwright() as p:
        browser = None
        context = None

        if use_cdp:
            print("connecting to existing Chrome via CDP...")

            browser = p.chromium.connect_over_cdp(
                "http://127.0.0.1:9222"
            )

            if not browser.contexts:
                raise RuntimeError(
                    "No browser contexts found. "
                    "Is Chrome running with --remote-debugging-port=9222 ?"
                )

            context = browser.contexts[0]

            page = context.new_page()

        elif use_profile and profile_path:

            print("launching Chrome profile...")

            context = p.chromium.launch_persistent_context(
                user_data_dir=profile_path,
                channel="chrome",
                headless=False,
                args=[
                    "--profile-directory=Profile 13",
                ],
            )

            page = context.pages[0] if context.pages else context.new_page()

        else:

            browser = p.chromium.launch(headless=False)

            page = browser.new_page()

        page.goto(url, wait_until="networkidle")

        ## Wait for the login.
        if "accounts.google.com" in page.url:
            print()
            print("Google login required.")
            print("Complete login in the browser.")
            print()

            page.wait_for_url(
                "**docs.google.com/forms/**",
                timeout=300000
            )

            print("Login detected.")

        current_page = 0
        while True:
            page.wait_for_selector("div[role='listitem']", timeout=10000)
            page.wait_for_timeout(500)

            if current_page in page_fields:
                fill_page(page, page_fields[current_page])

            if not click_next(page):
                break
            current_page += 1

        print()
        print("=" * 60)
        print("Form has been filled.")
        print("Review the answers in Chrome before submission.")
        print()
        print("Press ENTER to submit.")
        print("Close the browser window to cancel.")
        print("=" * 60)
        print()

        input()

        submitted = submit(page)

        if submitted:
            print("Detaching from browser...")

        if use_cdp:
            try: 
                browser.close()
            except Exception: 
                pass

        elif use_profile and context:

            context.close()

        else:

            page.context.browser.close()


def fill_page(page: Page, fields: list[tuple[FormField, str | list[str]]]):
    for form_field, answer in fields:
        if not answer:
            print(f"skipping [{form_field.index}] no answer available")
            continue
        try:
            fill_field(page, form_field, answer)
        except Exception as e:
            print(f"failed [{form_field.index}] '{form_field.question[:50]}': {e}")


def fill_field(page: Page, form_field: FormField, answer: str | list[str]):
    items = page.query_selector_all("div[role='listitem']")

    target = None
    for item in items:
        heading = item.query_selector("div[role='heading']")
        if not heading:
            continue
        if heading.inner_text().strip() == form_field.question.strip():
            target = item
            break

    if target is None:
        print(f"not found: '{form_field.question[:60]}'")
        return

    print(f"filling [{form_field.index}] type={form_field.field_type} answer='{answer}'")

    match form_field.field_type:
        case "short_text" | "long_text":
            inp = target.query_selector("input, textarea")
            if inp:
                inp.click()
                inp.fill(str(answer))

        case "radio":
            radios = target.query_selector_all("div[role='radio']")
            matched = False
            for radio in radios:
                data_val = radio.get_attribute("data-value") or ""
                span = radio.query_selector("span")
                label = data_val.strip() or (span.inner_text().strip() if span else "") or radio.inner_text().strip()
                if label.lower() == str(answer).strip().lower():
                    radio.click()
                    matched = True
                    break
            if not matched:
                labels = []
                for r in radios:
                    dv = r.get_attribute("data-value") or ""
                    sp = r.query_selector("span")
                    labels.append(dv.strip() or (sp.inner_text().strip() if sp else "") or r.inner_text().strip())
                print(f"  no radio match for '{answer}' in {labels}")

        case "checkbox":
            selected = [a.strip().lower() for a in (answer if isinstance(answer, list) else [answer])]
            boxes = target.query_selector_all("div[role='checkbox']")
            for box in boxes:
                label = (box.get_attribute("data-value") or box.inner_text()).strip().lower()
                if label in selected:
                    box.click()

        case "dropdown":
            listbox = target.query_selector("div[role='listbox']")
            if listbox:
                listbox.click()
                page.wait_for_selector("div[role='option']", timeout=3000)
                for opt in page.query_selector_all("div[role='option']"):
                    if opt.inner_text().strip().lower() == str(answer).strip().lower():
                        opt.click()
                        break


def click_next(page: Page) -> bool:
    for sel in ["div[role='button'] span:has-text('Next')", "span:has-text('Next')"]:
        btn = page.query_selector(sel)
        if btn:
            btn.click()
            try:
                page.wait_for_selector("div[role='listitem']", timeout=8000)
                page.wait_for_timeout(400)
                return True
            except Exception:
                return False
    return False


def submit(page: Page):

    for sel in [
        "div[role='button'] span:has-text('Submit')",
        "span:has-text('Submit')",
    ]:

        btn = page.query_selector(sel)

        if btn:

            btn.click()

            print("→→→→→→→→→→→→→→→→→→→→→→→→→→    ✓ Form submitted.")
            return True

    print("✗ Submit button not found.")
    return False


def debug_headings(url: str):
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=False)
        page = browser.new_page()
        page.goto(url, wait_until="networkidle")
        page.wait_for_selector("div[role='listitem']", timeout=10000)
        items = page.query_selector_all("div[role='listitem']")
        for i, item in enumerate(items):
            heading = item.query_selector("div[role='heading']")
            if heading:
                print(f"[{i}] repr: {repr(heading.inner_text().strip())}")
        browser.close()