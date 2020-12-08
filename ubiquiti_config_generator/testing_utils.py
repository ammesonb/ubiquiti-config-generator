"""
Some utilities for testing
"""
import functools


def counter_wrapper(func):
    """
    Adds a "counter" variable to the function, incrementing each time it is called
    """

    @functools.wraps(func)  # pragma: no mutate
    def execute(*args, **kwargs):
        """
        Adds a "counter" variable to the function, incrementing each time it is called
        """
        execute.counter += 1
        return func(*args, **kwargs)

    execute.counter = 0

    return execute
