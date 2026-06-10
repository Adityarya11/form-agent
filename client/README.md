# python client

````bash
# anonymous, dry run to preview answers
$ uv run python main.py 'FORM_URL' --dry-run

# anonymous, fills and waits for human review before submit
$ uv run python main.py 'FORM_URL'

# using Google profile, fills as you(the user), waits for review
$ uv run python main.py 'FORM_URL' --use-profile```

````
