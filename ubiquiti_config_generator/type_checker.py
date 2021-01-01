"""
Contains various constants for use in configurations
"""
import ipaddress
import re
from typing import Union

ENABLED = "enabled"
DISABLED = "disabled"

AUTO = "auto"
FULL = "full"
HALF = "half"

ACCEPT = "accept"
DROP = "drop"
REJECT = "reject"

ADDRESS = "address"
PORT = "port"

IN = "in"
OUT = "out"
LOCAL = "local"

TCP = "tcp"
UDP = "udp"
TCP_UDP = "tcp_udp"
IP = "ip"

NEW = "new"
INVALID = "invalid"
ESTABLISHED = "established"
RELATED = "related"

SPEEDS = [10, 100, 1000, 10000]

SOURCE = "source"
DESTINATION = "destination"
MASQUERADE = "masquerade"


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
        # Technically an integer will be converted to the octet form more commonly used
        # but probably isn't what _most_ people would expect to use so require a string
        return bool(ipaddress.ip_address(address)) and isinstance(address, str)
    except ValueError:
        return False


def is_subnet_mask(mask: Union[str, int]) -> bool:
    """
    Is input a subnet mask
    """
    return type(mask) in [str, int] and mask.isnumeric() and int(mask) in range(33)


def is_cidr(cidr: str) -> bool:
    """
    Is input of the form <address>/<subnet mask>
    """
    if "/" not in cidr:
        return False
    address, mask = cidr.split("/")

    return is_ip_address(address) and is_subnet_mask(mask)


def is_name(value: str) -> bool:
    """
    Is the value suitable for a field name
    """
    return is_string(value) and bool(re.match(r"^[a-zA-Z0-9\-_]+$", value))


def is_description(value: str) -> bool:
    """
    Is the value suitable for a description
    """
    # Can't contain quotes
    return is_string(value) and not "'" in value and not '"' in value


def is_string(value) -> bool:
    """
    Is the value a string
    """
    return isinstance(value, str)


def is_number(value: Union[int, str]) -> bool:
    """
    Is thing a number
    """
    return isinstance(value, int) or (isinstance(value, str) and value.isnumeric())


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


def is_mac(value: str) -> bool:
    """
    Is a MAC address
    """
    return isinstance(value, str) and bool(
        re.match(r"^[0-9a-fA-F]{2}([:\-][0-9a-fA-F]{2}){5}$", value)
    )


def is_translated_port(value: dict) -> bool:
    """
    Check if the dictionary is a single instance of one integer mapped to another
    """
    return (
        isinstance(value, dict)
        and len(value.items()) == 1
        and is_number(list(value.keys())[0])
        and is_number(list(value.values())[0])
    )


def is_address_and_or_port(value: dict) -> bool:
    """
    Does the value contain exclusively the fields address and/or port,
    with expected values for them
    """
    return (
        isinstance(value, dict)
        and list(value.keys()) in [[ADDRESS], [PORT], [ADDRESS, PORT]]
        and len(value.keys())
        and (ADDRESS not in value or is_string(value.get(ADDRESS, None)))
        and (
            PORT not in value
            or is_number(value.get(PORT, None))
            or is_string(value.get(PORT, None))
        )
    )


def is_source_destination(connections: dict) -> bool:
    """
    This is a dictionary with a source and/or destination property,
    containing addresses and ports
    """
    keys = list(connections.keys()) if isinstance(connections, dict) else []
    return (
        isinstance(connections, dict)
        # Only keys permissible are these three
        and not any(
            [key not in ["destination", "source", "allow", "rule"] for key in keys]
        )
        and isinstance(connections.get("source", {}), dict)
        and isinstance(connections.get("destination", {}), dict)
        and is_number(connections.get("rule", 0))
        # A source or destination must be set
        and (connections.get("source", {}) or connections.get("destination", {}))
        and is_string(connections.get("source", {}).get("address", ""))
        and is_string(connections.get("destination", {}).get("address", ""))
        and (
            is_string(connections.get("source", {}).get("port", ""))
            or is_number(connections.get("source", {}).get("port", ""))
        )
        and (
            is_string(connections.get("destination", {}).get("port", ""))
            or is_number(connections.get("destination", {}).get("port", ""))
        )
    )


def is_firewall_direction(value: str) -> bool:
    """
    Is the firewall direction valid
    """
    return value in [IN, OUT, LOCAL]


def is_protocol(value: str) -> bool:
    """
    Is the value a protocol
    """
    return value in [TCP, UDP, TCP_UDP, IP]


def is_state(value: dict) -> bool:
    """
    Is the value a set of connection states
    """
    keys = list(value.keys()) if isinstance(value, dict) else []
    return (
        isinstance(value, dict)
        and not any([key not in [NEW, ESTABLISHED, RELATED, INVALID] for key in keys])
        and all(
            [
                is_string_boolean(value.get(key, ENABLED))
                for key in [NEW, ESTABLISHED, RELATED, INVALID]
            ]
        )
    )


def is_nat_type(value: str) -> bool:
    """
    Is the value a NAT type
    """
    return value in [SOURCE, DESTINATION, MASQUERADE]
