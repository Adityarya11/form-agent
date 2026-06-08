import httpx


MCP_URL = "http://localhost:8080/mcp"


def _call(tool: str, params: dict) -> dict:
    resp = httpx.post(MCP_URL, json={"tool": tool, "params": params}, timeout=60.0)
    resp.raise_for_status()
    data = resp.json()
    if "error" in data and data["error"]:
        raise RuntimeError(f"MCP error [{tool}]: {data['error']}")
    return data.get("result")


def search(query: str, max_results: int = 3) -> list[dict]:
    return _call("search", {"query": query, "max_results": max_results}) or []


def resolve_batch(jobs: list[dict]) -> list[dict]: 
    return _call("resolve_batch", jobs) or []

def ask_llm(question: str, context: str = "", options: list[str] = None) -> str:
    result = _call("ask_llm", {
        "question": question,
        "context": context,
        "options": options or [],
    })
    return result.get("answer", "")


def rank_answers(question: str, candidate_a: str, candidate_b: str) -> dict:
    return _call("rank_answers", {
        "question": question,
        "candidate_a": candidate_a,
        "candidate_b": candidate_b,
    })