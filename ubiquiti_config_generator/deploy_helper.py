"""
Functionality needed for deploying and checking configurations
"""
import shlex
from typing import List, Optional


class ConfigDifference:
    """
    Calculates and stores differences between configuration commands
    """

    def __init__(self):
        # These contain path to the key -> key value
        # Indexed by command to enable a quick lookup
        self.removed = {}
        self.added = {}
        self.changed = {}
        self.preserved = {}

    def remove(self, command: dict) -> None:
        """
        Add a command that was removed
        """
        self.removed.update(command)

    def add(self, command: dict) -> None:
        """
        Add a command that was added
        """
        self.added.update(command)

    def change(self, command: dict) -> None:
        """
        Add a command that was changed
        """
        self.changed.update(command)

    def preserve(self, command: dict) -> None:
        """
        Add a command that was preserved
        """
        self.preserved.update(command)

    def compare_commands(
        self, current_command: Optional[dict], previous_command: Optional[dict]
    ) -> None:
        """
        Compare two commands and add to the appropriate dictionary
        """
        if current_command and not previous_command:
            self.add(current_command)
        elif not current_command and previous_command:
            self.remove(previous_command)
        elif current_command == previous_command:
            self.preserve(current_command)
        else:
            self.change(current_command)


def diff_configurations(
    current_commands: List[str], previous_commands: List[str]
) -> ConfigDifference:
    """
    Diff a configuration against its previous, summarizing changes
    """
    current_commands_by_key = {}
    previous_commands_by_key = {}

    for command in current_commands:
        current_commands_by_key[" ".join(shlex.split(command)[:-1])] = shlex.split(
            command
        )[-1]

    for command in previous_commands:
        previous_commands_by_key[" ".join(shlex.split(command)[:-1])] = shlex.split(
            command
        )[-1]

    difference = ConfigDifference()

    for command, value in current_commands_by_key.items():
        difference.compare_commands(
            {command: value},
            {command: previous_commands_by_key[command]}
            if command in previous_commands_by_key
            else None,
        )

    for command, value in previous_commands_by_key.items():
        difference.compare_commands(
            {command: current_commands_by_key[command]}
            if command in current_commands_by_key
            else None,
            {command: value},
        )

    return difference
