"""
Contains various constants for use in configurations
"""
import ipaddress
from typing import Union

ENABLED = "enabled"
DISABLED = "disabled"

AUTO = "auto"
FULL = "full"
HALF = "half"

ACCEPT = "accept"
DROP = "drop"
REJECT = "reject"

SPEEDS = [10, 100, 1000, 10000]


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


def is_duplex(value: str) -> bool:
    """
    Check duplex value
    """
    return value in [AUTO, FULL, HALF]


def is_speed(value: Union[int, str]) -> bool:
    """
    Is valid speed setting
    """
    return value in SPEEDS or value == AUTO


def is_action(value: str) -> bool:
    """
    Is an action for a packet
    """
    return value in [ACCEPT, DROP, REJECT]
