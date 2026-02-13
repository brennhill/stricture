# empty-constructs.py — Stress test: empty and minimal Python constructs.

class Empty:
    pass

class SingleField:
    x: int

def no_body():
    pass

def no_params() -> None:
    return None

def only_docstring():
    """This function has only a docstring."""

async def async_empty():
    pass

def many_params(a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p) -> bool:
    return True

class NestedEmpty:
    class Inner:
        class InnerInner:
            pass

def generator():
    yield

async def async_generator():
    yield

lambda_var = lambda: None
nested_lambda = lambda x: lambda y: lambda z: x + y + z
