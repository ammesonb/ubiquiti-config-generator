"""
Some utility functions
"""
import ipaddress


def get_duplicates(values: list) -> list:
    """
    Finds duplicates in a list
    """
    duplicates = list(
        set(filter(lambda value: value if values.count(value) > 1 else None, values))
    )

    duplicates.sort()
    return duplicates


def address_in_subnet(cidr: str, address: str) -> bool:
    """
    Check if a given address is in a subnet
    Wil return True if address is empty
    """
    return (not address) or ipaddress.ip_address(address) in ipaddress.ip_network(cidr)
