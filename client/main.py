import argparse
import time
import os
from dotenv import load_dotenv
from agent import run

load_dotenv()


def main():
    parser = argparse.ArgumentParser(description="form-agent: fill Google Forms using local LLM")
    parser.add_argument("url", help="Google Form URL")
    parser.add_argument("--docs", default="../docs", help="path to personal docs directory")
    parser.add_argument("--concurrency", type=int, default=2, help="parallel LLM batch size")
    parser.add_argument("--dry-run", action="store_true", help="resolve answers but do not fill")

    browser_group = parser.add_mutually_exclusive_group()
    browser_group.add_argument("--use-cdp", action="store_true", default=True, help="attach to existing Chrome via CDP (default)")
    browser_group.add_argument("--use-profile", action="store_true", help="launch Chrome with your profile path from .env")
    browser_group.add_argument("--anonymous", action="store_true", help="launch a fresh anonymous browser")

    llm_group = parser.add_mutually_exclusive_group()
    llm_group.add_argument("--use-local-llm", action="store_true", default=True, help="use local Ollama/Qwen (default)")
    llm_group.add_argument("--use-cloud-llm", action="store_true", help="use Gemini 2.0 Flash via API key in .env")

    args = parser.parse_args()

    profile_path = os.getenv("CHROME_PROFILE_PATH", "")
    if args.use_profile and not profile_path:
        print("error: --use-profile requires CHROME_PROFILE_PATH in .env")
        raise SystemExit(1)

    if args.use_cloud_llm:
        if not os.getenv("GEMINI_API_KEY"):
            print("error: --use-cloud-llm requires GEMINI_API_KEY in .env")
            raise SystemExit(1)
        os.environ["USE_CLOUD_LLM"] = "true"
        llm_label = "gemini-2.0-flash"
    else:
        os.environ["USE_CLOUD_LLM"] = "false"
        llm_label = "qwen2.5:3b (local)"

    if args.anonymous:
        browser_label = "anonymous"
    elif args.use_profile:
        browser_label = f"profile ({profile_path})"
    else:
        browser_label = "cdp (existing Chrome)"

    print(f"browser : {browser_label}")
    print(f"llm     : {llm_label}")
    print()

    start = time.time()
    run(
        url=args.url,
        docs_dir=args.docs,
        concurrency=args.concurrency,
        dry_run=args.dry_run,
        use_profile=args.use_profile,
        profile_path=profile_path,
        use_cdp=not args.anonymous and not args.use_profile,
    )
    print(f"total time: {time.time() - start:.1f}s")


if __name__ == "__main__":
    main()