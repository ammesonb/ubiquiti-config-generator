"""
Contains the host node
"""

from ubiquiti_config_generator import type_checker
from ubiquiti_config_generator.nodes.validatable import Validatable


HOST_TYPES = {
    "name": type_checker.is_string,
    "address": type_checker.is_ip_address,
    "mac": type_checker.is_mac,
    "address-groups": lambda groups: all(
        [type_checker.is_string(group) for group in groups]
    ),
    "forward-ports": lambda ports: all(
        [
            type_checker.is_number(port) or type_checker.is_translated_port(port)
            for port in ports
        ]
    ),
    "hairpin-ports": lambda ports: all(
        [
            type_checker.is_number(port) or type_checker.is_translated_port(port)
            for port in ports
        ]
    ),
    "allow-connect-from": lambda hosts: all(
        [type_checker.is_address_and_or_port(host) for host in hosts]
    ),
    "allow-connect-to": lambda hosts: all(
        [type_checker.is_address_and_or_port(host) for host in hosts]
    ),
}


class Host(Validatable):
    """
    A host
    """

    def __init__(self, name: str, **kwargs):
        super().__init__(HOST_TYPES, ["name"])
        self.name = name
        self._add_keyword_attributes(kwargs)

    def is_consistent(self) -> bool:
        """
        Check configuration for consistency
        """

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "Host " + self.name
