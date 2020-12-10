"""
Test external addresses
"""
from ubiquiti_config_generator.nodes import ExternalAddresses


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
