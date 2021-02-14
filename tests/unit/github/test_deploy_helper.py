"""
Deploy helper functionality testing
"""
import paramiko
import pytest

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


def test_compare_commands(monkeypatch):
    """
    .
    """
    # pylint: disable=unused-argument
    @counter_wrapper
    def compare_simple(*args, **kwargs):
        """
        .
        """

    @counter_wrapper
    def compare_list(*args, **kwargs):
        """
        .
        """

    monkeypatch.setattr(
        deploy_helper.ConfigDifference, "compare_simple_commands", compare_simple
    )
    monkeypatch.setattr(
        deploy_helper.ConfigDifference, "compare_list_commands", compare_list
    )

    diff = deploy_helper.ConfigDifference()
    diff.compare_commands({"abc": "def"}, {"abc": "ghi"})
    assert compare_simple.counter == 1, "Simple comparison"
    assert compare_list.counter == 0, "No list comparison"

    diff.compare_commands({"abc": ["def"]}, {"abc": "ghi"})
    assert compare_simple.counter == 1, "Not simple comparison"
    assert compare_list.counter == 1, "Current uses list comparison"

    diff.compare_commands({"abc": ["def"]}, {"abc": ["ghi"]})
    assert compare_simple.counter == 1, "Not simple comparison"
    assert compare_list.counter == 2, "Both use list comparison"

    diff.compare_commands({"abc": "def"}, {"abc": ["ghi"]})
    assert compare_simple.counter == 1, "Not simple comparison"
    assert compare_list.counter == 3, "Previous uses list comparison"


def test_compare_list_commands():
    """
    .
    """
    current = [
        "firewall port 1",
        "firewall port 3",
        "firewall port 2",
    ]

    previous = [
        "firewall port 2",
        "firewall port 4",
        "firewall port 3",
    ]

    difference = deploy_helper.diff_configurations(current, previous)
    print(difference.added)
    print(difference.removed)
    print(difference.preserved)
    assert difference.added == {"firewall port": ["1"]}, "Port one added"
    assert difference.removed == {"firewall port": ["4"]}, "Port four removed"
    assert difference.preserved == {
        "firewall port": ["3", "2"]
    }, "Ports two and three preserved"

    difference = deploy_helper.ConfigDifference()
    difference.compare_list_commands({"command": ["1", "2"]}, {})
    assert difference.added == {
        "command": ["1", "2"]
    }, "Commands added without previous"
    assert not difference.removed, "No removed"
    assert not difference.changed, "No changed"
    assert not difference.preserved, "No preserved"

    difference = deploy_helper.ConfigDifference()
    difference.compare_list_commands({}, {"command": ["1", "2"]})
    assert difference.removed == {
        "command": ["1", "2"]
    }, "Commands removed without current"
    assert not difference.added, "No added"
    assert not difference.changed, "No changed"
    assert not difference.preserved, "No preserved"


def test_compare_simple_commands():
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
                    ["firewall port 1", "firewall port 2", "firewall port 4",],
                ],
                [
                    "command bar",
                    "firewall baz",
                    "network foo",
                    "description test",
                    "interface wan",
                    "firewall port 1",
                    "firewall port 2",
                    "firewall port 4",
                ],
            )
        else:
            commands = (
                [
                    ["firewall ipsum",],
                    ["nat lorem", "command bar",],
                    ["interface wan", "address 192.168.0.1"],
                    ["firewall port 1", "firewall port 2", "firewall port 3",],
                ],
                [
                    "firewall ipsum",
                    "nat lorem",
                    "command bar",
                    "interface wan",
                    "address 192.168.0.1",
                    "firewall port 1",
                    "firewall port 2",
                    "firewall port 3",
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
        ["delete firewall port 3", "delete nat lorem", "delete address 192.168.0.1"],
        ["set firewall baz"],
        ["set network foo"],
        ["set description test"],
        ["set firewall port 4"],
    ], "Command to run (only difference) correct"

    monkeypatch.setattr(
        file_paths,
        "load_yaml_from_file",
        lambda file_path: {"apply-difference-only": False},
    )
    commands = deploy_helper.get_commands_to_run(".", ".")
    assert commands == [
        ["delete firewall port 3", "delete nat lorem", "delete address 192.168.0.1"],
        ["set command bar", "set firewall baz"],
        ["set network foo"],
        ["set description test"],
        ["set interface wan"],
        ["set firewall port 1", "set firewall port 2", "set firewall port 4"],
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
            'check_command $? "${output}"',
            'output=$(/bin/cmd-wrapper set firewall description "a description")',
            'check_command $? "${output}"',
            "output=$(/bin/cmd-wrapper delete network foo)",
            'check_command $? "${output}"',
            "",
            # pylint: disable=line-too-long
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


def test_router_connection(monkeypatch):
    """
    .
    """

    class Client:
        """
        Mock SSH client
        """

        expected_connection = {}

        # pylint: disable=unused-argument
        @counter_wrapper
        def load_system_host_keys(self):
            """
            .
            """

        def with_connection_auth(self, details: dict):
            """
            Inject expected connection auth, using shared dictionary
            to bypass issues with accessing instance
            """
            self.expected_connection.clear()
            self.expected_connection.update(details)

        def connect(self, host, post, username, **kwargs):
            """
            .
            """
            assert not kwargs["look_for_keys"], "Don't check system for keys"
            for key, value in self.expected_connection.items():
                assert kwargs[key] == value, "Auth details set"

    # pylint: disable=unused-argument
    @counter_wrapper
    def from_keyfile(file_path, passphrase):
        """
        .
        """
        return "bytes"

    monkeypatch.setattr(paramiko, "SSHClient", Client)
    monkeypatch.setattr(paramiko.RSAKey, "from_private_key_file", from_keyfile)

    config = {
        "router": {
            "address": "1.1.1.",
            "port": 22,
            "user": "admin",
            "keyfile": None,
            "password": None,
        }
    }

    with pytest.raises(ValueError):
        deploy_helper.get_router_connection(config)

    assert Client.load_system_host_keys.counter == 1, "Known hosts loaded"

    config["router"]["password"] = "passwd"
    Client().with_connection_auth({"password": "passwd"})
    deploy_helper.get_router_connection(config)

    config["router"]["keyfile"] = "filepath"
    Client().with_connection_auth({"pkey": "bytes"})
    deploy_helper.get_router_connection(config)

    config["router"]["password"] = None
    Client().with_connection_auth({"key_filename": "filepath"})
    deploy_helper.get_router_connection(config)


def test_write_data(monkeypatch):
    """
    .
    """

    # pylint: disable=too-few-public-methods,unused-argument
    class FakeFile:
        """
        Fake SFTP file
        """

        called = []

        def writable(self):
            """
            .
            """
            self.called.append("writable")

        def write(self, data):
            """
            .
            """
            self.called.append("write")

        def flush(self):
            """
            .
            """
            self.called.append("flush")

        def close(self):
            """
            .
            """
            self.called.append("close")

        def read(self):
            """
            .
            """
            self.called.append("read")
            return "foo".encode()

    class FailSFTP:
        """
        Will fail to return file
        """

        # pylint: disable=no-self-use
        def file(self, path, mode, buffer):
            """
            Opens file
            """
            return None

    # pylint: disable=no-self-use,too-few-public-methods
    class FakeSFTP:
        """
        Fake SFTP client
        """

        def file(self, path, mode, buffer):
            """
            Opens file
            """
            return FakeFile()

        def open(self, path):
            """
            .
            """
            return FakeFile()

        @counter_wrapper
        def close(self):
            """
            .
            """

    class FakeClient:
        """
        Fake client
        """

        def __init__(self, client):
            self.client = client

        def open_sftp(self):
            """
            .
            """
            return self.client

    with pytest.raises(ValueError):
        deploy_helper.write_data_to_router_file(FakeClient(FailSFTP()), "file", "data")

    with pytest.raises(ValueError):
        deploy_helper.write_data_to_router_file(FakeClient(FakeSFTP()), "file", "data")

    assert FakeFile.called == ["writable"], "Writable called, but failed"

    FakeFile.writable = lambda self: self.called.append("writable") or True
    assert not deploy_helper.write_data_to_router_file(
        FakeClient(FakeSFTP()), "file", "data"
    ), "Fails due to data mismatch"
    assert FakeFile.called == [
        "writable",
        "writable",
        "write",
        "flush",
        "close",
        "read",
    ], "Functions called in right order"
    assert FakeSFTP.close.counter == 1, "Close called"

    assert deploy_helper.write_data_to_router_file(
        FakeClient(FakeSFTP()), "file", "foo"
    ), "Successful emulated write"
    assert FakeFile.called == [
        "writable",
        "writable",
        "write",
        "flush",
        "close",
        "read",
        "writable",
        "write",
        "flush",
        "close",
        "read",
    ], "Functions called in right order"
    assert FakeSFTP.close.counter == 2, "Close called"


def test_run_command():
    """
    .
    """

    # pylint: disable=too-few-public-methods,no-self-use
    class FakeSession:
        """
        Fake session
        """

        command = None

        def exec_command(self, command):
            """
            .
            """
            self.command = command

    # pylint: disable=too-few-public-methods,unused-argument
    class FakeTransport:
        """
        Fake transport
        """

        @counter_wrapper
        def open_session(self):
            """
            .
            """
            return FakeSession()

    class FakeClient:
        """
        Fake client
        """

        @counter_wrapper
        def get_transport(self):
            """
            .
            """
            return FakeTransport()

    assert (
        deploy_helper.run_router_command(FakeClient(), "something").command
        == "something"
    ), "Command returned"
