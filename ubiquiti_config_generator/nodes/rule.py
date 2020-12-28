"""
A firewall rule
"""
from ubiquiti_config_generator import type_checker
from ubiquiti_config_generator.nodes.validatable import Validatable

# protocol
# state

RULE_TYPES = {
    "number": type_checker.is_number,
    "action": type_checker.is_action,
    "description": type_checker.is_string,
    "log": type_checker.is_string_boolean,
    "source": type_checker.is_address_and_or_port,
    "destination": type_checker.is_address_and_or_port,
    "protocol": type_checker.is_protocol,
    "state": type_checker.is_state,
}


class Rule(Validatable):
    """
    Represents a firewall rule
    """

    def __init__(self, number: int, **kwargs):
        super().__init__(RULE_TYPES, ["number"])
        self.number = number
        self._add_keyword_attributes(kwargs)
