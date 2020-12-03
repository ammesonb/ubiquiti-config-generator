"""
Contains port groups
"""
from typing import List


class PortGroup:
    """
    Represents a named grouping of ports
    """

    def __init__(self, name: str, ports: List[int] = None):
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
