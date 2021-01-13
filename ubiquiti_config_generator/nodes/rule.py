"""
A firewall rule
"""
import shlex
from typing import List

from ubiquiti_config_generator import type_checker, secondary_configs
from ubiquiti_config_generator.nodes.validatable import Validatable


RULE_TYPES = {
    "number": type_checker.is_number,
    "action": type_checker.is_action,
    "description": type_checker.is_description,
    "log": type_checker.is_string_boolean,
    "source": type_checker.is_address_and_or_port,
    "destination": type_checker.is_address_and_or_port,
    "protocol": type_checker.is_protocol,
    "state": type_checker.is_state,
}

# Maybe someday this disabling of duplicate will actually be honored
# pylint: disable=duplicate-code


class Rule(Validatable):
    """
    Represents a firewall rule
    """

    def __init__(self, number: int, firewall_name: str, config_path: str, **kwargs):
        super().__init__(RULE_TYPES, ["number"])
        self.number = number
        self.firewall_name = firewall_name
        self.config_path = config_path
        self._add_keyword_attributes(kwargs)

    # pylint: disable=too-many-branches
    def commands(self) -> List[str]:
        """
        Get the command for this rule
        """
        command_base = "firewall name {0} rule {1} ".format(
            self.firewall_name, self.number
        )

        commands = [command_base + "action " + getattr(self, "action", "accept")]

        if hasattr(self, "description"):
            # pylint: disable=no-member
            description = shlex.quote(self.description)
            if description[0] not in ["'", '"']:
                description = '"{0}"'.format(description)

            commands.append(command_base + "description " + description)

        for part in ["log", "protocol"]:
            if hasattr(self, part):
                commands.append(command_base + part + " " + str(getattr(self, part)))

        if hasattr(self, "state"):
            # pylint: disable=no-member
            for state, enabled in self.state.items():
                commands.append(command_base + "state {0} {1}".format(state, enabled))

        connections = ["source", "destination"]
        for connection in connections:
            if hasattr(self, connection):
                data = getattr(self, connection)

                if "address" in data:
                    if type_checker.is_ip_address(
                        data["address"]
                    ) or type_checker.is_cidr(data["address"]):
                        commands.append(
                            command_base + connection + " address " + data["address"]
                        )
                    else:
                        # Address groups defined by hosts or statically, so can't know
                        # this exhaustively, unlike port groups
                        # So have to assume user knows what they're doing on this one
                        commands.append(
                            command_base
                            + connection
                            + " group address-group "
                            + data["address"]
                        )

                if "port" in data:
                    if type_checker.is_number(data["port"]):
                        commands.append(
                            command_base + connection + " port " + str(data["port"])
                        )
                    else:
                        commands.append(
                            command_base
                            + connection
                            + " group port-group "
                            + data["port"]
                        )

        return commands

    # pylint: disable=duplicate-code
    def validate(self) -> bool:
        """
        Is the rule valid
        """
        valid = super().validate()

        port_groups = [
            group.name for group in secondary_configs.get_port_groups(self.config_path)
        ]
        for connection in ["source", "destination"]:
            if hasattr(self, connection) and "port" in getattr(self, connection):
                if (
                    not type_checker.is_number(getattr(self, connection)["port"])
                    and getattr(self, connection)["port"] not in port_groups
                ):
                    self.add_validation_error(
                        "Rule {0} has nonexistent {1} port group {2}".format(
                            self.number, connection, getattr(self, connection)["port"]
                        )
                    )
                    valid = False

        return valid

    def __str__(self) -> str:
        """
        A string representation of the rule
        """
        return "Firewall {0} rule {1}".format(self.firewall_name, self.number)

    def print(self) -> None:
        """
        Print the rule
        """
        for attr in self.attributes():
            print("{0} | {1}".format(attr, getattr(self, attr)))
