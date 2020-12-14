"""
Test utility functions
"""

from ubiquiti_config_generator import utility


def test_get_duplicates():
    """
    .
    """
    assert not utility.get_duplicates([1, 2, 3]), "No duplicates in list"
    assert not utility.get_duplicates(["1", "2", "3"]), "No duplicates in list"
    assert utility.get_duplicates([1, 2, 3, 1]) == [1], "Duplicate returned"
    assert utility.get_duplicates(["a", "bc", "ab", "ab", "cd", "ab", "a"]) == [
        "a",
        "ab",
    ], "Duplicates returned"
