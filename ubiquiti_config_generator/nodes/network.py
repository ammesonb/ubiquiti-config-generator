"""
Contains the network node
"""
from ubiquiti_config_generator.nodes.validatable import Validatable
from ubiquiti_config_generator import type_checker

# Allow too few public methods, for now
# pylint: disable=too-few-public-methods

NETWORK_TYPES = {
    "name": lambda value: True,
    "subnet": type_checker.is_cidr,
    "default-router": type_checker.is_ip_address,
    "dns-server": type_checker.is_ip_address,
    "dns-servers": lambda servers: all(
        [type_checker.is_ip_address(addr) for addr in servers]
    ),
    # Don't think this can be invalid?
    "domain-name": lambda value: True,
    "lease": type_checker.is_number,
    "start": type_checker.is_ip_address,
    "stop": type_checker.is_ip_address,
}


class Network(Validatable):
    """
    A network to be created
    """

    def __init__(self, name: str, cidr: str, **kwargs):
        super().__init__(NETWORK_TYPES, ["name", "cidr"])
        self.name = name
        self.cidr = cidr
        self._add_keyword_attributes(kwargs)
