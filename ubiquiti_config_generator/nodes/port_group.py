"""
Contains port groups
"""
import shlex
from typing import List

from ubiquiti_config_generator.nodes.validatable import Validatable
from ubiquiti_config_generator import type_checker, utility

PORT_GROUP_TYPES = {
    "name": type_checker.is_name,
    "description": type_checker.is_description,
    "ports": lambda ports: ports
    and all([type_checker.is_number(port) for port in ports]),
}


class PortGroup(Validatable):
    """
    Represents a named grouping of ports
    """

    def __init__(self, name: str, ports: List[int] = None, description: str = None):
        super().__init__(PORT_GROUP_TYPES, ["ports"])
        self.__name = name
        self.__ports = ports or []
        if description:
            self._add_validate_attribute("description")
            self.description = description

    def add_port(self, port: int) -> None:
        """
        Adds a port
        """
        self.__ports.append(port)

    def add_ports(self, ports: List[int]) -> None:
        """
        Adds multiple ports
        """
        self.__ports.extend(ports)

    @property
    def name(self) -> str:
        """
        Returns the name of the group
        """
        return self.__name

    @property
    def ports(self) -> List[int]:
        """
        Returns ports in the group
        """
        return self.__ports

    def is_consistent(self) -> bool:
        """
        Check configuration for consistency
        """
        duplicates = utility.get_duplicates(self.ports)
        if duplicates:
            self.add_validation_error(
                "{0} has duplicate ports: {1}".format(
                    str(self), ", ".join([str(dup) for dup in duplicates])
                )
            )

        return not bool(duplicates)

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "Port group " + self.name

    def commands(self) -> List[str]:
        """
        Commands to generate the port group
        """
        base_command = "firewall group port-group {0}".format(self.name)
        return [base_command + " port {0}".format(port) for port in self.ports] + (
            [base_command + " description {0}".format(shlex.quote(self.description))]
            if hasattr(self, "description")
            else []
        )
