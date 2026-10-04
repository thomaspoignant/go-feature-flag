"""Shared fixtures and builders for the WASM evaluation test suites."""

BOOL_FLAG = {
    "variations": {"on": True, "off": False},
    "defaultRule": {"variation": "on"},
    "trackEvents": True,
}


def nested_ctx(depth: int) -> dict:
    """Evaluation context whose attributes nest `depth` dict levels deep."""
    ctx: dict = {"targetingKey": "user-1"}
    cur = ctx
    for _ in range(depth):
        cur["a"] = {}
        cur = cur["a"]
    cur["x"] = 1
    return ctx


def deep_paren_query(depth: int) -> str:
    """nikunjy query wrapped in `depth` nested parentheses."""
    return "(" * depth + 'targetingKey eq "user-1"' + ")" * depth


def int_in_list_query(count: int, attr: str = "age") -> str:
    """
    nikunjy `in` list of `count` integers. List parsing is right-recursive,
    so each item costs one parser stack frame — the production trigger of
    issue #5651 was a flat allow-list of ~200 integers.
    """
    return f"{attr} in [" + ",".join(str(i + 1) for i in range(count)) + "]"


def or_chain_query(count: int, attr: str = "age") -> str:
    """
    Flat bracket-less or-chain of `count` conditions. Logical expressions are
    binary and recursive in the nikunjy parser, so each operator costs stack
    frames while showing none of the bracket nesting or list commas the other
    recursion drivers do.
    """
    return " or ".join(f"{attr} eq {i + 1}" for i in range(count))


def split_int_in_list_query(count: int, chunk: int, attr: str = "age") -> str:
    """
    The same allow-list split into or-joined `in` chunks of at most `chunk`
    items, so the parser recursion never exceeds `chunk` frames per list.
    """
    parts = []
    for start in range(0, count, chunk):
        items = ",".join(str(i + 1) for i in range(start, min(start + chunk, count)))
        parts.append(f"({attr} in [{items}])")
    return " or ".join(parts)


def query_flag(query: str) -> dict:
    """Boolean flag whose single targeting rule uses `query`."""
    return {
        "variations": {"on": True, "off": False},
        "targeting": [{"query": query, "variation": "on"}],
        "defaultRule": {"variation": "off"},
    }
