"""
A firewall node
"""
from typing import Tuple, List

from ubiquiti_config_generator import type_checker
from ubiquiti_config_generator.nodes.rule import Rule
from ubiquiti_config_generator.nodes.validatable import Validatable


FIREWALL_TYPES = {
    "name": type_checker.is_name,
    "direction": type_checker.is_firewall_direction,
    "default-action": type_checker.is_action,
    "auto-increment": type_checker.is_number,
}


class Firewall(Validatable):
    """
    The firewall object
    """

    def __init__(
        self, name: str, direction: str, network_name: str, config_path: str, **kwargs
    ):
        super().__init__(FIREWALL_TYPES, ["name"])
        self.name = name
        self.network_name = network_name
        self.config_path = config_path
        self.direction = direction
        self._add_keyword_attributes(kwargs)
        self._rules = []

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "Firewall " + self.name

    def commands(self) -> Tuple[List[List[str]], List[str]]:
        """
        Commands to create this firewall
        """
        pass

    @property
    def rules(self) -> List[Rule]:
        """
        Get rules for the firewall
        """
        return self._rules
