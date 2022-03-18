# Python plugin over gRPC

## install SDK

Before you develop your python plugin, you need to install an dependency as SDK.

```bash
$ python3 -m pip install funppy
```

## create plugin functions

Then you can write your plugin functions in python. The functions can be very flexible, only the following restrictions should be complied with.

- function should return at most one value and one error.
- `funppy.register()` must be called to register plugin functions and `funppy.serve()` must be called to start a plugin server process.

Here is some plugin functions as example.

```python
import logging
from typing import List

import funppy


def sum_two_int(a: int, b: int) -> int:
    return a + b

def sum_ints(*args: List[int]) -> int:
    result = 0
    for arg in args:
        result += arg
    return result

def Sum(*args):
    result = 0
    for arg in args:
        result += arg
    return result


if __name__ == '__main__':
    funppy.register("sum_two_int", sum_two_int)
    funppy.register("sum_ints", sum_ints)
    funppy.register("sum", Sum)
    funppy.serve()
```

You can get more examples at [funppy/examples/].

## build plugin

Python plugins do not need to be complied, just make sure its file suffix is `.py` by convention and should not be changed.

## use plugin functions

Finally, you can use `Init` to initialize plugin via the `xxx.py` path, and you can call the plugin API to handle plugin functionality.


[funppy/examples/]: ../funppy/examples/
