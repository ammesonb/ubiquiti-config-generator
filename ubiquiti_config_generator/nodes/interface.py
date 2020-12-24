"""
An interface node
"""
import copy
from os import path
from typing import List

from ubiquiti_config_generator import type_checker, file_paths
from ubiquiti_config_generator.nodes.validatable import Validatable
from ubiquiti_config_generator.nodes import Firewall


# TODO: collapse this into the network, with firewalls under that
INTERFACE_TYPES = {
    "description": type_checker.is_string,
    "duplex": type_checker.is_duplex,
    "speed": type_checker.is_speed,
    "vif": type_checker.is_number,
    "name": type_checker.is_name,
    "network_name": type_checker.is_string,
    "firewalls": lambda firewalls: all([firewall.validate() for firewall in firewalls]),
}


class Interface(Validatable):
    """
    The interface node
    """

    def __init__(self, name: str, config_path: str, network_name: str, **kwargs):
        # Address is valid if it either directly corresponds to the parent network or
        # is in itself a valid CIDR address + mask
        validator_map = copy.deepcopy(INTERFACE_TYPES)
        validator_map[
            "address"
        ] = lambda value: value == network_name or type_checker.is_cidr(value)

        super().__init__(validator_map, ["name, network_name"])
        self.name = name
        self.network_name = network_name
        self.config_path = config_path

        self._add_keyword_attributes(kwargs)
        if "firewalls" not in kwargs:
            self._load_firewalls()

    def _load_firewalls(self):
        """
        Get firewalls for this interface
        """
        self.firewalls = [
            Firewall(
                name=firewall_path.split(path.sep)[-2],
                **(file_paths.load_yaml_from_file(firewall_path))
            )
            for firewall_path in file_paths.get_folders_with_config(
                file_paths.get_path(
                    [
                        self.config_path,
                        file_paths.NETWORK_FOLDER,
                        self.network_name,
                        file_paths.INTERFACE_FOLDER,
                        self.name,
                        file_paths.FIREWALL_FOLDER,
                    ]
                )
            )
        ]
        self._add_validate_attribute("firewalls")

    def validation_failures(self) -> List[str]:
        """
        Get all validation failures
        """
        failures = self.validation_errors()
        for firewall in self.firewalls:
            failures.extend(firewall.validation_errors())
        return failures

    def is_consistent(self) -> bool:
        """
        Check configuration for consistency
        """
        consistent = True

        for firewall_index in range(len(self.firewalls)):
            first_firewall = self.firewalls[firewall_index]
            matches = [
                second_firewall
                for second_firewall in self.firewalls[firewall_index + 1 :]
                if first_firewall.direction == second_firewall.direction
            ]
            if matches:
                self.add_validation_error(
                    "{0} shares direction with {1}".format(
                        str(first_firewall),
                        ", ".join([str(firewall) for firewall in matches]),
                    )
                )
                consistent = False

        return consistent

    def validate(self) -> bool:
        """
        Is the root node valid
        """
        return super().validate() and all(
            [firewall.validate() for firewall in self.firewalls]
        )

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "Interface " + self.name
