"""
Deploy helper functionality testing
"""

from ubiquiti_config_generator import root_parser, file_paths
from ubiquiti_config_generator.github import deploy_helper
from ubiquiti_config_generator.nodes import GlobalSettings, ExternalAddresses, NAT
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_get_command_key():
    """
    .
    """
    assert (
        deploy_helper.get_command_key("a simple key value") == "a simple key"
    ), "basic string correct"
    assert (
        deploy_helper.get_command_key("something more-complex 40")
        == "something more-complex"
    ), "With dash"
    assert (
        deploy_helper.get_command_key(
            "a much longer key-value config-pair network name foo"
        )
        == "a much longer key-value config-pair network name"
    ), "Very long"
    assert (
        deploy_helper.get_command_key(
            'firewall name test description "this is a firewall"'
        )
        == "firewall name test description"
    ), "With quoted argument"


def test_compare_commands():
    """
    .
    """
    current = [
        "command 1",
        "firewall 2",
        'description "a command description"',
        "some stuff",
        "foo bar",
    ]

    previous = [
        "some thing",
        'description "a command description"',
        "foo baz",
        "firewall 2",
        "obsolete interface",
    ]

    difference = deploy_helper.diff_configurations(current, previous)
    assert difference.added == {"command": "1"}, "Added correct"
    assert difference.removed == {"obsolete": "interface"}, "Removed correct"
    assert difference.changed == {"some": "stuff", "foo": "bar"}, "Changed correct"
    assert difference.preserved == {
        "description": "a command description",
        "firewall": "2",
    }, "Preserved correct"


def test_get_commands_to_run(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(
        root_parser.RootNode,
        "create_from_configs",
        lambda *args, **kwargs: root_parser.RootNode(
            GlobalSettings(), [], ExternalAddresses([]), [], NAT(".", [])
        ),
    )

    # pylint: disable=unused-argument
    @counter_wrapper
    def get_commands(self):
        """
        .
        """
        if get_commands.counter % 2:
            commands = (
                [
                    ["command bar", "firewall baz",],
                    ["network foo",],
                    ["description test",],
                    ["interface wan",],
                ],
                [
                    "command bar",
                    "firewall baz",
                    "network foo",
                    "description test",
                    "interface wan",
                ],
            )
        else:
            commands = (
                [
                    ["firewall ipsum",],
                    ["nat lorem", "command bar",],
                    ["interface wan", "address 192.168.0.1"],
                ],
                [
                    "firewall ipsum",
                    "nat lorem",
                    "command bar",
                    "interface wan",
                    "address 192.168.0.1",
                ],
            )
        return commands

    monkeypatch.setattr(root_parser.RootNode, "get_commands", get_commands)

    monkeypatch.setattr(
        file_paths,
        "load_yaml_from_file",
        lambda file_path: {"apply-difference-only": True},
    )
    commands = deploy_helper.get_commands_to_run(".", ".")
    assert commands == [
        ["delete nat lorem", "delete address 192.168.0.1"],
        ["set firewall baz"],
        ["set network foo"],
        ["set description test"],
    ], "Command to run (only difference) correct"

    monkeypatch.setattr(
        file_paths,
        "load_yaml_from_file",
        lambda file_path: {"apply-difference-only": False},
    )
    commands = deploy_helper.get_commands_to_run(".", ".")
    assert commands == [
        ["delete nat lorem", "delete address 192.168.0.1"],
        ["set command bar", "set firewall baz"],
        ["set network foo"],
        ["set description test"],
        ["set interface wan"],
    ], "Command to run (full config) correct"


def test_generate_bash_commands():
    """
    .
    """
    commands = [
        "set value 1",
        'set firewall description "a description"',
        "delete network foo",
    ]
    deploy_config = {
        "script-cfg-path": "/bin/cmd-wrapper",
        "auto-rollback-on-failure": True,
        "reboot-after-minutes": 10,
        "save-after-commit": False,
    }
    bash_commands = deploy_helper.generate_bash_commands(commands, deploy_config)
    assert bash_commands == "\n".join(
        [
            'trap "exit 1" TERM',
            "export TOP_PID=$1",
            "",
            "function check_command() {",
            "  status=$1",
            '  output="${2}"',
            "",
            "  if [ $status -ne 0 ]; then",
            '    echo "Failed to execute command:"',
            '    echo "${output}"',
            "    /bin/cmd-wrapper discard",
            "    kill -s TERM $TOP_PID",
            "  fi",
            "}",
            "",
            "/bin/cmd-wrapper begin",
            "",
            "output=$(/bin/cmd-wrapper set value 1)",
            'check_output $? "${output}"',
            'output=$(/bin/cmd-wrapper set firewall description "a description")',
            'check_output $? "${output}"',
            "output=$(/bin/cmd-wrapper delete network foo)",
            'check_output $? "${output}"',
            "",
            'sudo sg vyattacfg -c "/opt/vyatta/sbin/vyatta-config-mgmt.pl --action=commit-confirm --minutes=10"',
            "if [ $? -ne 0 ]; then",
            '  echo "Failed to schedule reboot!"',
            "  kill -s TERM $TOP_PID",
            "fi",
            "",
            "/bin/cmd-wrapper commit",
            "if [ $? -ne 0 ]; then",
            '  echo "Failed to commit!"',
            "  kill -s TERM $TOP_PID",
            "fi",
            "",
            "exit 0",
            "",
        ]
    ), "Commands generated as expected"

    deploy_config.update(
        {
            "reboot-after-minutes": None,
            "save-after-commit": True,
            "auto-rollback-on-failure": False,
        }
    )

    bash_commands = deploy_helper.generate_bash_commands(commands, deploy_config)
    assert bash_commands == "\n".join(
        [
            'trap "exit 1" TERM',
            "export TOP_PID=$1",
            "",
            "/bin/cmd-wrapper begin",
            "",
            "/bin/cmd-wrapper set value 1",
            '/bin/cmd-wrapper set firewall description "a description"',
            "/bin/cmd-wrapper delete network foo",
            "",
            "/bin/cmd-wrapper commit",
            "if [ $? -ne 0 ]; then",
            '  echo "Failed to commit!"',
            "  kill -s TERM $TOP_PID",
            "fi",
            "",
            "/bin/cmd-wrapper save",
            "if [ $? -ne 0 ]; then",
            '  echo "Failed to save!"',
            "  kill -s TERM $TOP_PID",
            "fi",
            "",
            "exit 0",
            "",
        ]
    ), "Commands generated as expected"
