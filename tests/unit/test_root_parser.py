"""
Tests the root parser
"""

# pylint: disable=protected-access

from ubiquiti_config_generator import (
    root_parser,
    file_paths,
    secondary_configs,
)
from ubiquiti_config_generator.nodes import (
    GlobalSettings,
    ExternalAddresses,
    PortGroup,
    Network,
)
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_create_from_config(monkeypatch):
    """
    .
    """

    @counter_wrapper
    def fake_load_global_settings(config_path: str):
        """
        .
        """
        return GlobalSettings()

    @counter_wrapper
    def fake_load_external_addresses(config_path: str):
        """
        .
        """
        return ExternalAddresses([])

    @counter_wrapper
    def fake_load_port_groups(config_path: str):
        """
        .
        """
        return []

    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_get_folders_with_config(folder: str):
        """
        .
        """
        return []

    monkeypatch.setattr(
        secondary_configs, "get_global_configuration", fake_load_global_settings
    )
    monkeypatch.setattr(
        secondary_configs, "get_external_addresses", fake_load_external_addresses
    )
    monkeypatch.setattr(secondary_configs, "get_port_groups", fake_load_port_groups)
    monkeypatch.setattr(
        file_paths, "get_folders_with_config", fake_get_folders_with_config
    )

    root_node = root_parser.RootNode.create_from_configs(".")
    assert isinstance(root_node, root_parser.RootNode), "Root node returned"
    assert fake_load_global_settings.counter == 1, "Global settings retrieved"
    assert fake_load_external_addresses.counter == 1, "External addresses retrieved"
    assert fake_load_port_groups.counter == 1, "Port groups retrieved"
    assert fake_get_folders_with_config.counter == 1, "Networks retrieved"


def test_is_valid(monkeypatch):
    """
    .
    """
    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_validate():
        """
        .
        """
        return True

    @counter_wrapper
    def fake_consistent():
        """
        .
        """
        return False

    node = root_parser.RootNode(None, None, None, None)
    monkeypatch.setattr(node, "is_valid", fake_validate)
    monkeypatch.setattr(node, "is_consistent", fake_consistent)

    assert not node.validate(), "Validation fails if consistency fails"
    assert fake_validate.counter == 1, "Validity checked"
    assert fake_consistent.counter == 1, "Consistency checked"


def test_validate(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_validate(self):
        """
        .
        """
        return True

    monkeypatch.setattr(GlobalSettings, "validate", fake_validate)
    monkeypatch.setattr(ExternalAddresses, "validate", fake_validate)
    monkeypatch.setattr(PortGroup, "validate", fake_validate)
    monkeypatch.setattr(Network, "validate", fake_validate)

    node = root_parser.RootNode(
        GlobalSettings(),
        [PortGroup("group")],
        ExternalAddresses([]),
        [Network("Network", ".", "1.1.1.1/24")],
    )
    assert node.is_valid(), "Node is valid"
    assert fake_validate.counter == 4, "All things validated"


def test_validation_failures(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_validate(self):
        """
        .
        """
        return ["failure"]

    monkeypatch.setattr(GlobalSettings, "validation_errors", fake_validate)
    monkeypatch.setattr(ExternalAddresses, "validation_errors", fake_validate)
    monkeypatch.setattr(PortGroup, "validation_errors", fake_validate)
    monkeypatch.setattr(Network, "validation_failures", fake_validate)

    node = root_parser.RootNode(
        GlobalSettings(),
        [PortGroup("group")],
        ExternalAddresses([]),
        [Network("Network", ".", "1.1.1.1/24")],
    )
    result = node.validation_failures()
    print(result)
    assert result == ["failure"] * 4, "Validation failures returned"


def test_consistency_checks_called(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def check_consistency(self):
        """
        .
        """
        return True

    monkeypatch.setattr(GlobalSettings, "is_consistent", check_consistency)
    monkeypatch.setattr(ExternalAddresses, "is_consistent", check_consistency)
    monkeypatch.setattr(PortGroup, "is_consistent", check_consistency)
    monkeypatch.setattr(Network, "is_consistent", check_consistency)

    monkeypatch.setattr(file_paths, "get_folders_with_config", lambda folder: [])

    root = root_parser.RootNode(
        GlobalSettings(),
        [PortGroup("Ports", [80])],
        ExternalAddresses(["1.1.1.1"]),
        [Network("Network", ".", "10.0.0.0/24")],
    )

    assert root.is_consistent(), "Node is consistent"
    assert check_consistency.counter == 4, "Consistency checked for each object"


def test_network_overlap_consistency(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(GlobalSettings, "is_consistent", lambda self: True)
    monkeypatch.setattr(ExternalAddresses, "is_consistent", lambda self: True)
    monkeypatch.setattr(PortGroup, "is_consistent", lambda self: True)
    monkeypatch.setattr(Network, "is_consistent", lambda self: True)

    root = root_parser.RootNode(
        GlobalSettings(),
        [PortGroup("Ports", [80])],
        ExternalAddresses(["1.1.1.1"]),
        [
            Network("Network 1", ".", "10.0.0.0/24"),
            Network("Network 2", ".", "10.0.1.0/24"),
            Network("Network 3", ".", "10.0.2.0/24"),
        ],
    )

    assert root.is_consistent(), "Networks do not overlap"

    overlap_root = root_parser.RootNode(
        GlobalSettings(),
        [PortGroup("Ports", [80])],
        ExternalAddresses(["1.1.1.1"]),
        [
            # First network contains all the others
            Network("Network 1", ".", "10.0.0.0/22"),
            # This network has no collisions inside it
            Network("Network 2", ".", "10.0.1.0/24"),
            # This network collides with the next
            Network("Network 2", ".", "10.0.2.0/23"),
            # This network has no collisions inside it
            Network("Network 3", ".", "10.0.2.0/24"),
        ],
    )

    assert not overlap_root.is_consistent(), "Networks overlap"
    networks = overlap_root.networks
    assert networks[0].validation_errors() == [
        "{0} overlaps with {1}".format(networks[0], networks[1]),
        "{0} overlaps with {1}".format(networks[0], networks[2]),
        "{0} overlaps with {1}".format(networks[0], networks[3]),
    ], "Network 0 contains collisions"

    assert networks[1].validation_errors() == [], "Network 1 contains no collisions"
    assert networks[2].validation_errors() == [
        "{0} overlaps with {1}".format(networks[2], networks[3]),
    ], "Network 2 contains collision with 3"
    assert networks[3].validation_errors() == [], "Network 3 contains no collisions"


def test_get_commands(monkeypatch):
    """
    .
    """

    @counter_wrapper
    def get_port_group_commands(self):
        """
        .
        """
        if get_port_group_commands.counter == 1:
            return ["group1-command"]
        else:
            return ["group2-command"]

    @counter_wrapper
    def get_network_commands(self):
        """
        .
        """
        if get_network_commands.counter == 1:
            return (
                [["network1-command", "network1-command2"], ["network1-command3"]],
                ["network1-command", "network1-command2", "network1-command3"],
            )
        else:
            return (
                [["network2-command", "network2-command2"], ["network2-command3"]],
                ["network2-command", "network2-command2", "network2-command3"],
            )

    monkeypatch.setattr(GlobalSettings, "commands", lambda self: ["settings-command"])
    monkeypatch.setattr(
        ExternalAddresses, "commands", lambda self: ["addresses-command"]
    )
    monkeypatch.setattr(PortGroup, "commands", get_port_group_commands)
    monkeypatch.setattr(Network, "commands", get_network_commands)

    parser = root_parser.RootNode(
        GlobalSettings(),
        [PortGroup("group1"), PortGroup("group2")],
        ExternalAddresses([]),
        [
            Network("network1", ".", "1.1.1.1/24"),
            Network("network2", ".", "2.2.2.2/24"),
        ],
    )
    commands = parser.get_commands()
    print(commands[0])
    print(commands[1])
    assert len(commands) == 2, "Two entries in tuple"
    assert commands[0] == [
        ["addresses-command", "group1-command", "group2-command"],
        ["settings-command"],
        [
            "network1-command",
            "network1-command2",
            "network2-command",
            "network2-command2",
        ],
        ["network1-command3", "network2-command3"],
    ], "Ordered commands correct"
    assert commands[1] == [
        "addresses-command",
        "group1-command",
        "group2-command",
        "settings-command",
        "network1-command",
        "network1-command2",
        "network1-command3",
        "network2-command",
        "network2-command2",
        "network2-command3",
    ], "Command list correct"
