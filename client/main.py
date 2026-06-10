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
    parser.add_argument("--use-profile", action="store_true", help="use your Chrome profile for authenticated submission")
    parser.add_argument(
    "--use-cdp",
    action="store_true",
    help="attach to existing Chrome via CDP"
)
    args = parser.parse_args()

    profile_path = os.getenv("CHROME_PROFILE_PATH", "")

    if args.use_profile and not profile_path:
        print("error: --use-profile requires CHROME_PROFILE_PATH in .env")
        print("  example: CHROME_PROFILE_PATH=/home/youruser/.config/google-chrome")
        raise SystemExit(1)

    start = time.time()
    run(
        url=args.url,
        docs_dir=args.docs,
        concurrency=args.concurrency,
        dry_run=args.dry_run,
        use_profile=args.use_profile,
        profile_path=profile_path,
        use_cdp=args.use_cdp,
    )
    print(f"total time: {time.time() - start:.1f}s")


if __name__ == "__main__":
    main()