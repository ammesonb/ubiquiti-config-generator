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
    # TODO: commands for this
    "forward-ports": lambda ports: all(
        [
            type_checker.is_number(port)
            or type_checker.is_string(port)
            or type_checker.is_translated_port(port)
            for port in ports
        ]
    ),
    # TODO: commands for this
    "hairpin-ports": lambda hosts: all(
        [type_checker.is_address_and_or_port(host) for host in hosts]
    ),
    # This is a dictionary for connections to allow/block, with properties:
    # allow: bool
    # rule: optional[int] - rule number for use in firewall
    # log: optional[bool] - log this connection
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

    def __init__(
        self, name: str, network: "Network", config_path: str, address: str, **kwargs
    ):
        super().__init__(HOST_TYPES, ["name"])
        self.name = name
        self.network = network
        self.config_path = config_path
        self.address = address
        self.connections = []
        self._add_keyword_attributes(kwargs)
        self.add_firewall_rules()

    # pylint: disable=too-many-branches
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

        for connection in self.connections:
            rule = connection.get("rule", None)
            if not rule:
                continue

            # Ensure the firewall rule numbers don't conflict with the numbers
            # set in a host
            for firewall in self.network.firewalls_by_direction.values():
                if any(
                    [firewall_rule.number == rule for firewall_rule in firewall.rules]
                ):
                    self.add_validation_error(
                        "{0} has conflicting connection rule with {1}, "
                        "rule number {2}".format(str(self), str(firewall), rule)
                    )
                    consistent = False

            # TODO: test this
            # Ensure either the source or destination contains this host
            # in each connection, otherwise can't know what firewall to add a rule to
            source = None
            destination = None
            if "source" in connection and "address" in connection["source"]:
                source = connection["source"]["address"]
            if "destination" in connection and "address" in connection["destination"]:
                destination = connection["destination"]["address"]

            if source is None and destination is None:
                self.add_validation_error(
                    str(self)
                    + " has connection with no source address or destination address!"
                )
                consistent = False
            elif self.address not in [source, destination] and not any(
                [
                    group
                    for group in getattr(self, "address-groups", [])
                    if group in [source, destination]
                ]
            ):
                self.add_validation_error(
                    str(self)
                    + " has connection where its address is not used in source or destination!"
                )
                consistent = False

        return consistent

    # TODO: test this
    def add_firewall_rules(self):
        """
        Add rules to the firewalls in the network for the host's connections
        """
        for connection in self.connections:
            # Must be either source or destination to be valid, so
            # only have to check for the source to match
            connection_is_source = "source" in connection and (
                connection["source"].get("address", "") == self.address
                or any(
                    [
                        group == connection["source"].get("address", "")
                        for group in getattr(self, "address-groups", [])
                    ]
                )
            )

            rule_properties = {
                "action": "accept" if connection["allow"] else "drop",
                "protocol": connection.get("protocol", "tcp_udp"),
                "log": "enable" if connection.get("log", False) else "disable",
            }

            if "rule" in connection:
                rule_properties["rule"] = connection["rule"]
            if "source" in connection:
                rule_properties["source"] = connection["source"]
            if "destination" in connection:
                rule_properties["destination"] = connection["destination"]

            self.network.firewalls_by_direction[
                "in" if connection_is_source else "out"
            ].add_rule(rule_properties)

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "Host " + self.name

    def commands(self) -> Tuple[List[List[str]], List[str]]:
        """
        Generates commands to create this host
        """
