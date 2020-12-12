"""
Some utility functions
"""


def get_duplicates(values: list) -> list:
    """
    Finds duplicates in a list
    """
    return list(
        set(filter(lambda value: value if values.count(value) > 1 else None, values))
    )
