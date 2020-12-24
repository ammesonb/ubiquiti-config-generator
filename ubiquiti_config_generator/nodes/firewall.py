"""
A firewall node
"""

from ubiquiti_config_generator import type_checker
from ubiquiti_config_generator.nodes.validatable import Validatable


FIREWALL_TYPES = {
    "name": type_checker.is_name,
    "direction": type_checker.is_firewall_direction,
    "default-action": type_checker.is_action,
    "auto-increment": type_checker.is_number,
}


# Disable for now
# pylint: disable=too-few-public-methods
class Firewall(Validatable):
    """
    The firewall object
    """

    def __init__(self, name: str, direction: str, **kwargs):
        super().__init__(FIREWALL_TYPES, ["name"])
        self.name = name
        self.direction = direction
        self._add_keyword_attributes(kwargs)

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "Firewall " + self.name
