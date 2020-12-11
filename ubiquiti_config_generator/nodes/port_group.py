"""
Contains port groups
"""
from typing import List

from ubiquiti_config_generator.nodes.validatable import Validatable
from ubiquiti_config_generator import type_checker

PORT_GROUP_TYPES = {
    "ports": lambda ports: all([type_checker.is_number(port) for port in ports])
}


# Allow too few public methods, for now
# pylint: disable=too-few-public-methods


class PortGroup(Validatable):
    """
    Represents a named grouping of ports
    """

    def __init__(self, name: str, ports: List[int] = None):
        super().__init__(PORT_GROUP_TYPES, ["ports"])
        self.__name = name
        self.__ports = ports or []

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

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "Port group " + self.name
