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


def test_address_in_subnet():
    """
    .
    """
    assert not utility.address_in_subnet(
        "10.0.0.0/8", "host-in-subnet"
    ), "String address group invalid"
    assert not utility.address_in_subnet(
        "10.0.0.0/8", None
    ), "Empty address is not valid"
    assert utility.address_in_subnet("10.0.0.0/8", "10.1.0.1"), "In subnet valid"
    assert utility.address_in_subnet("10.0.0.0/8", "10.0.0.1"), "In subnet valid"
    assert utility.address_in_subnet("10.0.0.0/8", "10.255.255.255"), "In subnet valid"
    assert not utility.address_in_subnet(
        "10.0.0.0/8", "11.0.0.0"
    ), "Outside subnet invalid"
    assert not utility.address_in_subnet(
        "10.0.0.0/8", "9.0.0.0"
    ), "Outside subnet invalid"
