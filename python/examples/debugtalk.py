import logging
import sys
from typing import List

sys.path.insert(0, "/Users/debugtalk/MyProjects/HttpRunner-dev/plugin/python/")

import plugin


def sum(*args):
    result = 0
    for arg in args:
        result += arg
    return result

def sum_ints(*args: List[int]) -> int:
    result = 0
    for arg in args:
        result += arg
    return result

def sum_two_int(a: int, b: int) -> int:
    return a + b

def sum_two_string(a: str, b: str) -> str:
    return a + b

def sum_strings(*args: List[str]) -> str:
    result = ""
    for arg in args:
        result += arg
    return result

def concatenate(*args: List[str]) -> str:
    result = ""
    for arg in args:
        result += str(arg)
    return result

def setup_hook_example(name):
    logging.warn("setup_hook_example")
    return f"setup_hook_example: {name}"

def teardown_hook_example(name):
    logging.warn("teardown_hook_example")
    return f"teardown_hook_example: {name}"


if __name__ == '__main__':
    plugin.register("sum", sum)
    plugin.register("sum_ints", sum_ints)
    plugin.register("concatenate", concatenate)
    plugin.register("sum_two_int", sum_two_int)
    plugin.register("sum_two_string", sum_two_string)
    plugin.register("sum_strings", sum_strings)
    plugin.register("setup_hook_example", setup_hook_example)
    plugin.register("teardown_hook_example", teardown_hook_example)

    plugin.serve()
