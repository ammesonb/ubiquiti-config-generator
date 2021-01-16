"""
Functionality needed for deploying and checking configurations
"""
import shlex
from typing import List, Optional

from ubiquiti_config_generator import root_parser, file_paths


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


def get_command_key(command: str) -> str:
    """
    Gets the command key, everything except the last space-separated value
    """
    return " ".join(shlex.split(command)[:-1])


def diff_configurations(
    current_commands: List[str], previous_commands: List[str]
) -> ConfigDifference:
    """
    Diff a configuration against its previous, summarizing changes
    """
    current_commands_by_key = {}
    previous_commands_by_key = {}

    for command in current_commands:
        current_commands_by_key[get_command_key(command)] = shlex.split(command)[-1]

    for command in previous_commands:
        previous_commands_by_key[get_command_key(command)] = shlex.split(command)[-1]

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


# Most excess locals are convenience, and improve readability
# pylint: disable=too-many-locals
def get_commands_to_run(
    current_config_path: str, previous_config_path: str, only_return_diff: bool = False
) -> List[List[str]]:
    """
    Given two sets of configurations, returns the ordered command sets to execute
    """
    deploy_config = file_paths.load_yaml_from_file("deploy.yaml")

    current_config = root_parser.RootNode.create_from_configs(current_config_path)
    previous_config = root_parser.RootNode.create_from_configs(previous_config_path)

    current_ordered_commands, current_command_list = current_config.get_commands()
    # The previous ordered commands are unused, but need the list
    # pylint: disable=unused-variable
    previous_ordered_commands, previous_command_list = previous_config.get_commands()

    difference = diff_configurations(current_command_list, previous_command_list)

    run_commands = [[]]

    # Run deletes in a single batch first, since that _shouldn't_ cause any issues
    for key, value in difference.removed.items():
        run_commands[0].append(" ".join(["delete", key, shlex.quote(value)]))

    apply_diff_only = only_return_diff or deploy_config["apply-difference-only"]
    for command_set in current_ordered_commands:
        run_commands.append([])

        for command in command_set:
            command_prefix = (
                command if not apply_diff_only else get_command_key(command)
            )

            # Include commands only if applying the entire config, or
            # the command's value changed
            if (
                not apply_diff_only
                or command_prefix in difference.changed
                or command_prefix in difference.added
            ):
                run_commands[-1].append("set " + command)

        if not run_commands[-1]:
            del run_commands[-1]

    return run_commands
