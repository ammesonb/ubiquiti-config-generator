"""
External IP addresses
"""

from typing import List

from ubiquiti_config_generator.nodes.validatable import Validatable
from ubiquiti_config_generator import type_checker, utility

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
        duplicates = utility.get_duplicates(self.addresses)
        if duplicates:
            self.add_validation_error(
                "{0} has duplicate addresses: {1}".format(
                    str(self), ", ".join(duplicates)
                )
            )

        return not bool(duplicates)

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "External address"
