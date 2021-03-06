"""
NAT for all networks
"""
from os import path
from typing import Tuple, List

from ubiquiti_config_generator import type_checker, file_paths, utility
from ubiquiti_config_generator.nodes.nat_rule import NATRule
from ubiquiti_config_generator.nodes.validatable import Validatable


NAT_TYPES = {
    "auto-increment": type_checker.is_number,
    "rules": lambda rules: all([rule.validate() for rule in rules]),
}


class NAT(Validatable):
    """
    The NAT object
    """

    def __init__(self, config_path: str, rules: List[NATRule] = None, **kwargs):
        super().__init__(NAT_TYPES, ["rules"])
        self.config_path = config_path
        setattr(self, "auto-increment", kwargs.get("auto-increment", 10))

        self.rules = rules or []
        if not rules:
            self._load_rules()

    def _load_rules(self):
        """
        Load rules for this firewall
        """
        for rule_path in file_paths.get_config_files(
            [self.config_path, file_paths.NAT_FOLDER,]
        ):
            if type_checker.is_number(rule_path.split(path.sep)[-1].rstrip(".yaml")):
                self.add_rule(
                    {
                        "number": rule_path.split(path.sep)[-1].rstrip(".yaml"),
                        "config_path": self.config_path,
                        **(file_paths.load_yaml_from_file(rule_path)),
                    }
                )

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "NAT"

    def commands(self) -> Tuple[List[List[str]], List[str]]:
        """
        Commands to create this firewall
        """
        command_list = []

        for rule in self.rules:
            command_list.extend(rule.commands())

        return command_list

    def add_rule(self, rule_properties: dict):
        """
        Add a rule to the list
        """
        if "number" not in rule_properties:
            rule_properties["number"] = self.next_rule_number()

        self.rules.append(NATRule(**rule_properties))

    def next_rule_number(self) -> int:
        """
        Find the next number usable for a rule
        """
        next_number = None
        to_check = getattr(self, "auto-increment")
        while next_number is None:
            if int(to_check) in [int(rule.number) for rule in self.rules]:
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

    def is_consistent(self) -> bool:
        """
        Are the NAT rules consistent
        """
        consistent = True
        duplicate_numbers = utility.get_duplicates([rule.number for rule in self.rules])
        if duplicate_numbers:
            self.add_validation_error(
                "NAT has duplicate rules: "
                + ", ".join([str(rule) for rule in duplicate_numbers])
            )
            consistent = False

        return consistent
