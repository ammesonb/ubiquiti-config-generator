"""
Contains various constants for use in configurations
"""
import ipaddress
from typing import Union

ENABLED = "enabled"
DISABLED = "disabled"


def is_string_boolean(value: str) -> bool:
    """
    Checks if a given value is either enabled or disabled
    """
    return value in [ENABLED, DISABLED]


def is_ip_address(address: str) -> bool:
    """
    Check if an address is a valid ip
    """
    try:
        return bool(ipaddress.ip_address(address))
    except ValueError:
        return False


def is_subnet_mask(mask: str) -> bool:
    """
    Is input a subnet mask
    """
    return mask.isnumeric() and int(mask) in range(33)


def is_cidr(cidr: str) -> bool:
    """
    Is input of the form <address>/<subnet mask>
    """
    if "/" not in cidr:
        return False
    address, mask = cidr.split("/")

    return is_ip_address(address) and is_subnet_mask(mask)


def is_number(value: Union[int, str]) -> bool:
    """
    Is thing a number
    """
    return isinstance(value, int) or value.isnumeric()
