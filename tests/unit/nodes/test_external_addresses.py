"""
Test external addresses
"""
from ubiquiti_config_generator.nodes import ExternalAddresses
from ubiquiti_config_generator import utility


def test_addresses_set():
    """
    .
    """
    addresses = ExternalAddresses(["1.1.1.1", "2.2.2.2"])
    assert addresses.addresses == ["1.1.1.1", "2.2.2.2"], "Addresses set"


def test_str():
    """
    .
    """
    assert str(ExternalAddresses([])) == "External address", "String returned"


def test_is_consistent(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(utility, "get_duplicates", lambda array: [])
    addresses = ExternalAddresses([])
    assert addresses.is_consistent(), "No duplicates is consistent"

    monkeypatch.setattr(utility, "get_duplicates", lambda array: ["10", "20"])

    duplicate_address = ExternalAddresses([])
    assert not duplicate_address.is_consistent(), "Duplicate address is not consistent"
    assert duplicate_address.validation_errors() == [
        str(duplicate_address) + " has duplicate addresses: 10, 20"
    ], "Validation error added"
