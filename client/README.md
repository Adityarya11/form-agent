# python client

```bash
# defaults: CDP + local LLM
uv run python main.py 'FORM_URL'

# CDP + cloud LLM
uv run python main.py 'FORM_URL' --use-cloud-llm

# your Chrome profile + cloud LLM
uv run python main.py 'FORM_URL' --use-profile --use-cloud-llm

# anonymous + local (testing only)
uv run python main.py 'FORM_URL' --anonymous

# dry run with cloud
uv run python main.py 'FORM_URL' --use-cloud-llm --dry-run

```
