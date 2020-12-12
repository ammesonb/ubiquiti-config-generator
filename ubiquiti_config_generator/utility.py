"""
Some utility functions
"""


def get_duplicates(values: list) -> list:
    """
    Finds duplicates in a list
    """
    duplicates = list(
        set(filter(lambda value: value if values.count(value) > 1 else None, values))
    )

    duplicates.sort()
    return duplicates
