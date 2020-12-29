"""
Contains the host node
"""
from typing import Tuple, List

from ubiquiti_config_generator import secondary_configs, type_checker, utility
from ubiquiti_config_generator.nodes.validatable import Validatable


HOST_TYPES = {
    "name": type_checker.is_name,
    "address": type_checker.is_ip_address,
    "mac": type_checker.is_mac,
    "address-groups": lambda groups: all(
        [type_checker.is_string(group) for group in groups]
    ),
    "forward-ports": lambda ports: all(
        [
            type_checker.is_number(port)
            or type_checker.is_string(port)
            or type_checker.is_translated_port(port)
            for port in ports
        ]
    ),
    "hairpin-ports": lambda hosts: all(
        [type_checker.is_address_and_or_port(host) for host in hosts]
    ),
    # This is a dictionary for connections to allow/block, with properties:
    # allow: bool
    # rule: optional[int] - rule number for use in firewall
    # source:
    #    address: IP/address group
    #    port: port/port group
    # destination:
    #    address: IP/address group
    #    port: port/port group
    "connections": lambda hosts: all(
        [type_checker.is_source_destination(host) for host in hosts]
    ),
}


class Host(Validatable):
    """
    A host
    """

    def __init__(self, name: str, network: "Network", config_path: str, **kwargs):
        super().__init__(HOST_TYPES, ["name"])
        self.name = name
        self.network = network
        self.config_path = config_path
        self.connections = []
        self._add_keyword_attributes(kwargs)

    def is_consistent(self) -> bool:
        """
        Check configuration for consistency
        """
        consistent = True
        port_groups = secondary_configs.get_port_groups(self.config_path)
        port_group_names = [group.name for group in port_groups]
        for port in getattr(self, "forward-ports", []):
            if not type_checker.is_number(port) and port not in port_group_names:
                self.add_validation_error(
                    "Port Group {0} not defined for forwarding in {1}".format(
                        port, str(self)
                    )
                )
                consistent = False

        for port in getattr(self, "hairpin-ports", []):
            if not type_checker.is_number(port) and port not in port_group_names:
                self.add_validation_error(
                    "Port Group {0} not defined for hairpin in {1}".format(
                        port, str(self)
                    )
                )
                consistent = False

        for connection in getattr(self, "connections", []):
            source_port = connection.get("source", {}).get("port", 0)
            if (
                not type_checker.is_number(source_port)
                and source_port not in port_group_names
            ):
                self.add_validation_error(
                    "Source Port Group {0} not defined for "
                    "{1} connection in {2}".format(
                        source_port,
                        "allowed" if connection["allow"] else "blocked",
                        str(self),
                    )
                )
                consistent = False

            destination_port = connection.get("destination", {}).get("port", 0)
            if (
                not type_checker.is_number(destination_port)
                and destination_port not in port_group_names
            ):
                self.add_validation_error(
                    "Destination Port Group {0} not defined for "
                    "{1} connection in {2}".format(
                        destination_port,
                        "allowed" if connection["allow"] else "blocked",
                        str(self),
                    )
                )
                consistent = False

        # Check for duplicate rule values
        duplicate_rules = list(
            filter(
                lambda rule: rule is not None,
                utility.get_duplicates(
                    [
                        connection["rule"] if "rule" in connection else None
                        for connection in self.connections
                    ]
                ),
            )
        )
        if duplicate_rules:
            self.add_validation_error(
                str(self)
                + " has duplicate firewall rules: "
                + ", ".join([str(rule) for rule in duplicate_rules])
            )
            consistent = False

        # Ensure the firewall rule numbers don't conflict with the numbers
        # set in a host
        all_firewall_rules = (
            self.network.firewalls_by_direction["in"].rules
            + self.network.firewalls_by_direction["out"].rules
        )
        for connection in self.connections:
            rule = connection.get("rule", None)
            if not rule:
                continue

            for (
                firewall_direction,
                firewall,
            ) in self.network.firewalls_by_direction.items():
                if any(
                    [firewall_rule.number == rule for firewall_rule in firewall.rules]
                ):
                    self.add_validation_error(
                        "{0} has conflicting connection rule with {1}, rule number {2}".format(
                            str(self), str(firewall), rule
                        )
                    )
                    consistent = False

        return consistent

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "Host " + self.name

    def commands(self) -> Tuple[List[List[str]], List[str]]:
        """
        Generates commands to create this host
        """
