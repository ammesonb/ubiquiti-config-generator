"""
External IP addresses
"""

from typing import List

from ubiquiti_config_generator.nodes.validatable import Validatable
from ubiquiti_config_generator import type_checker

EXTERNAL_ADDRESS_TYPES = {
    "addresses": lambda addresses: all(
        [type_checker.is_ip_address(addr) for addr in addresses]
    )
}


class ExternalAddresses(Validatable):
    """
    External IP addressses
    """

    def __init__(self, addresses: List[str]):
        super().__init__(EXTERNAL_ADDRESS_TYPES, ["addresses"])
        self.addresses = addresses

    def is_consistent(self) -> bool:
        """
        Check configuration for consistency
        """

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "External address"
