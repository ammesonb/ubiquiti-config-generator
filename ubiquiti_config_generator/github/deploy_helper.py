"""
Functionality needed for deploying and checking configurations
"""
import shlex
from typing import List, Optional

import paramiko
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

    def add_or_append_list_value(self, component: dict, key: str, value: str) -> None:
        """
        Adds a given value for a key to a command configuration difference
        """
        if key not in component:
            component[key] = [value]
        else:
            component[key].append(value)

    def compare_commands(
        self, current_command: Optional[dict], previous_command: Optional[dict]
    ) -> None:
        """
        Compare two commands and add to the appropriate dictionary
        """
        is_list = (
            current_command and isinstance(list(current_command.values())[0], list)
        ) or (previous_command and isinstance(list(previous_command.values())[0], list))

        if is_list:
            self.compare_list_commands(current_command, previous_command)
        else:
            self.compare_simple_commands(current_command, previous_command)

    def compare_list_commands(
        self, current_command: Optional[dict], previous_command: Optional[dict]
    ) -> None:
        """
        Compare two commands with list values
        """
        command_key = (
            list(current_command.keys())[0]
            if current_command
            else list(previous_command.keys())[0]
        )
        current_values = list(current_command.values())[0] if current_command else []
        previous_values = list(previous_command.values())[0] if previous_command else []

        for each_value in current_values:
            if each_value in previous_values:
                self.add_or_append_list_value(self.preserved, command_key, each_value)
            else:
                self.add_or_append_list_value(self.added, command_key, each_value)

        for each_value in previous_values:
            # Don't need preserved, since already added in current
            if each_value not in current_values:
                self.add_or_append_list_value(self.removed, command_key, each_value)

    def compare_simple_commands(
        self, current_command: Optional[dict], previous_command: Optional[dict]
    ) -> None:
        """
        Compare two commands with single values and add to the appropriate dictionary
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
        command_key = get_command_key(command)
        command_value = shlex.split(command)[-1]
        # If command already found and it's a list,
        # add this to it for comparison purposes
        if command_key in current_commands_by_key and isinstance(
            current_commands_by_key[command_key], list
        ):
            current_commands_by_key[command_key].append(command_value)
        # Otherwise, if not a list, make it into a list with new value
        elif command_key in current_commands_by_key:
            current_commands_by_key[command_key] = [
                current_commands_by_key[command_key],
                command_value,
            ]
        else:
            current_commands_by_key[command_key] = command_value

    for command in previous_commands:
        command_key = get_command_key(command)
        command_value = shlex.split(command)[-1]
        # If command already found and it's a list,
        # add this to it for comparison purposes
        if command_key in previous_commands_by_key and isinstance(
            previous_commands_by_key[command_key], list
        ):
            previous_commands_by_key[command_key].append(command_value)
        # Otherwise, if not a list, make it into a list with new value
        elif command_key in previous_commands_by_key:
            previous_commands_by_key[command_key] = [
                previous_commands_by_key[command_key],
                command_value,
            ]
        else:
            previous_commands_by_key[command_key] = command_value

    difference = ConfigDifference()

    for command, value in current_commands_by_key.items():
        difference.compare_commands(
            {command: value},
            {command: previous_commands_by_key[command]}
            if command in previous_commands_by_key
            else None,
        )

    for command, value in previous_commands_by_key.items():
        # Lists already had a full diff done in the first pass for the current values
        # if there was a value for it
        if isinstance(value, list) and command in current_commands_by_key:
            continue

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

    apply_diff_only = only_return_diff or deploy_config["apply-difference-only"]
    run_commands = [[]]

    # Run deletes in a single batch first, since that _shouldn't_ cause any issues
    for key, value in difference.removed.items():
        if isinstance(value, list):
            for each_value in value:
                run_commands[0].append(
                    " ".join(["delete", key, shlex.quote(each_value)])
                )
        else:
            run_commands[0].append(" ".join(["delete", key, shlex.quote(value)]))

    # If applying whole diff, ensure nat and firewalls get cleared, since
    # rule re-ordering could mean something from a previous rule gets merged
    # e.g. before:
    # rule 20: source <addr> port 80
    # after preserves source, ADDING destination instead of overwriting the original:
    # rule 20 :destination <addr2> source <addr>
    if not apply_diff_only:
        # First get every firewall in the previous config
        previous_firewall_names = [
            firewall.name
            for network in previous_config.networks
            for firewall in network.firewalls
        ]
        # And then restrict the current ones to those in the previous, since
        # won't need to reset a net-new firewall configuration
        current_firewall_names = [
            firewall.name
            for network in current_config.networks
            for firewall in network.firewalls
            if firewall.name in previous_firewall_names
        ]
        run_commands[0].extend(
            [
                "delete service nat",
                *[
                    f"delete firewall name {firewall_name} rule"
                    for firewall_name in current_firewall_names
                ],
            ]
        )

        # Can have multiple DHCP networks set, but on updates
        # this will likely overlap (e.g. if shrinking or extending) address pool
        # need to delete any that are in both previous and current configs
        previous_dhcp_networks = [
            network.name
            for network in previous_config.networks
            if hasattr(network, "stop") and network.stop
        ]

        current_dhcp_networks = [
            network
            for network in current_config.networks
            if hasattr(network, "stop")
            and network.stop
            and network.name in previous_dhcp_networks
        ]

        run_commands[0].extend(
            [
                f"delete service dhcp-server shared-network-name {network.name} "
                f"subnet {network.cidr} start"
                for network in current_dhcp_networks
            ]
        )

    for command_set in current_ordered_commands:
        run_commands.append([])

        for command in command_set:
            command_prefix = (
                command if not apply_diff_only else get_command_key(command)
            )

            # Include commands only if applying the entire config
            should_include = not apply_diff_only
            if isinstance(
                difference.changed.get(command_prefix, None), list
            ) or isinstance(difference.added.get(command_prefix, None), list):
                command_value = shlex.split(command)[-1]
                should_include = should_include or (
                    command_value in difference.added.get(command_prefix, [])
                    or command_value in difference.changed.get(command_prefix, [])
                )
            else:
                # or the command's value is new or changed
                should_include = should_include or (
                    command_prefix in difference.changed
                    or command_prefix in difference.added
                )

            # the command's value changed
            if should_include:
                run_commands[-1].append("set " + command)

        if not run_commands[-1]:
            del run_commands[-1]

    return run_commands


def generate_bash_commands(commands: List[str], deploy_config: dict) -> str:
    """
    Creates the commands to execute for vbash to update the configuration
    """
    header = (
        "function check_command() {\n"
        "  status=$1\n"
        '  output="${2}"\n'
        '  command="${3}"\n'
        "\n"
        "  if [ $status -ne 0 ]; then\n"
        '    echo "Failed to execute command: ${command}" >&2\n'
        '    echo "${output}" >&2\n'
        f"    {deploy_config['script-cfg-path']} discard\n"
        "    kill -s TERM $TOP_PID\n"
        "  fi\n"
        "}\n\n"
    )

    output = "\n".join(['trap "exit 1" TERM', "export TOP_PID=$1", ""]) + "\n"

    output += (
        header if deploy_config["auto-rollback-on-failure"] else ""
    ) + f"{deploy_config['script-cfg-path']} begin\n\n"

    command_template = (
        f"{deploy_config['script-cfg-path']} {{0}}\n"
        if not deploy_config["auto-rollback-on-failure"]
        else (
            "command={1}\n"
            f'output=$({deploy_config["script-cfg-path"]} {{0}})\n'
            'check_command $? "${{output}}" "${{command}}"\n'
        )
    )

    for command in commands:
        output += command_template.format(command, shlex.quote(command))

    if deploy_config["reboot-after-minutes"]:
        output += (
            # commit-confirm isn't included in the script wrapper, for reasons??
            '\nsudo sg vyattacfg -c "'
            # Need to echo y to auto-confirm,
            # since there is a prompt that can't be automated
            "echo y | /opt/vyatta/sbin/vyatta-config-mgmt.pl "
            "--action=commit-confirm "
            f"--minutes={deploy_config['reboot-after-minutes']}\"\n"
            "if [ $? -ne 0 ]; then\n"
            '  echo "Failed to schedule reboot!"\n'
            "  kill -s TERM $TOP_PID\n"
            "fi\n"
        )

    output += "\n"

    output += (
        f'{deploy_config["script-cfg-path"]} commit\n'
        "if [ $? -ne 0 ]; then\n"
        '  echo "Failed to commit!"\n'
        "  kill -s TERM $TOP_PID\n"
        "fi\n\n"
    )

    if deploy_config["save-after-commit"]:
        output += (
            f'{deploy_config["script-cfg-path"]} save\n'
            "if [ $? -ne 0 ]; then\n"
            '  echo "Failed to save!"\n'
            "  kill -s TERM $TOP_PID\n"
            "fi\n\n"
        )

    output += "exit 0\n"

    return output


def get_router_connection(deploy_config: dict) -> paramiko.SSHClient:
    """
    Opens a connection to the configured router

    Can throw:
        - ValueError
        - BadHostKeyException
        - AuthenticationException
        - SSHException
        - socket.error
    """
    router_config = deploy_config["router"]

    client = paramiko.SSHClient()
    client.load_system_host_keys()
    auth = {}
    if router_config["keyfile"]:
        if router_config["password"]:
            auth["pkey"] = paramiko.RSAKey.from_private_key_file(
                router_config["keyfile"], router_config["password"]
            )
        else:
            auth["key_filename"] = router_config["keyfile"]
    elif router_config["password"]:
        auth["password"] = router_config["password"]
    else:
        # Can't connect if no keyfile or password specified
        raise ValueError("No password or keyfile provided")

    client.connect(
        router_config["address"],
        router_config.get("port", 22),
        router_config["user"],
        look_for_keys=False,
        **auth,
    )

    return client


def write_data_to_router_file(
    client: paramiko.SSHClient, file_path: str, file_data: str
) -> bool:
    """
    Writes data to a file on the router

    Throws ValueError
    """
    ftp = client.open_sftp()  # type: paramiko.sftp_client.SFTPClient
    remote_file = ftp.file(file_path, "w", -1)  # type: paramiko.sftp_file.SFTPFile
    if not remote_file:
        raise ValueError("Failed to open remote file")
    if not remote_file.writable():
        raise ValueError("Remote file not writable")

    remote_file.write(file_data)
    remote_file.flush()
    remote_file.close()

    written = ftp.open(file_path).read().decode()
    ftp.close()

    return written == file_data


def run_router_command(client: paramiko.SSHClient, command: str) -> paramiko.Channel:
    """
    Runs a given command on the router
    Returns the CLOSED channel with exit code, stdout/err, etc
    """
    command_session = client.get_transport().open_session()  # type: paramiko.Channel
    command_session.exec_command(command)
    return command_session
