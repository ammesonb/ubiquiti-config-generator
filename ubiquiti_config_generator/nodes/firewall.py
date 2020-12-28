"""
A firewall node
"""
from os import path
from typing import Tuple, List

from ubiquiti_config_generator import type_checker, file_paths
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
        super().__init__(FIREWALL_TYPES, ["name", "rules"])
        self.name = name
        self.direction = direction
        self.network_name = network_name
        self.config_path = config_path
        if "auto-increment" not in kwargs:
            setattr(self, "auto-increment", 10)

        self.rules = []
        if not "rules" in kwargs:
            self._load_rules()

        self._add_keyword_attributes(kwargs)

    def _load_rules(self):
        """
        Load rules for this firewall
        """
        for rule_path in file_paths.get_config_files(
            file_paths.get_path(
                [
                    self.config_path,
                    file_paths.NETWORK_FOLDER,
                    self.network_name,
                    file_paths.FIREWALL_FOLDER,
                    self.name,
                ]
            )
        ):
            if type_checker.is_number(rule_path.split(path.sep)[-1].rstrip(".yaml")):
                self.add_rule(
                    {
                        "number": rule_path.split(path.sep)[-1].rstrip(".yaml"),
                        **(file_paths.load_yaml_from_file(rule_path)),
                    }
                )

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

    def add_rule(self, rule_properties: dict):
        """
        Add a rule to the list
        """
        if "number" not in rule_properties:
            rule_properties["number"] = self.next_rule_number()
        if "firewall_name" not in rule_properties:
            rule_properties["firewall_name"] = self.name

        self.rules.append(Rule(**rule_properties))

    def next_rule_number(self) -> int:
        """
        Find the next number usable for a rule
        """
        next_number = None
        to_check = getattr(self, "auto-increment")
        while next_number is None:
            if to_check in [rule.number for rule in self.rules]:
                to_check += getattr(self, "auto-increment")
            else:
                next_number = to_check

        return next_number

    def validation_failures(self) -> List[str]:
        """
        Get all validation failures
        """
        failures = self.validation_errors()
        for rule in self.rules:
            failures.extend(rule.validation_errors())
        return failures

    def validate(self) -> bool:
        """
        Is the firewall valid
        """
        return super().validate() and all([rule.validate() for rule in self.rules])
