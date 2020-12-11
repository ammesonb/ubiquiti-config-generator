"""
Tests the root parser
"""

# pylint: disable=protected-access

from ubiquiti_config_generator import root_parser, file_paths
from ubiquiti_config_generator.nodes import (
    GlobalSettings,
    ExternalAddresses,
    PortGroup,
    Network,
)
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_get_global_settings(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_load_yaml(file_path: str):
        """
        .
        """
        return {}

    monkeypatch.setattr(file_paths, "load_yaml_from_file", fake_load_yaml)

    settings = root_parser._get_global_configuration()
    assert isinstance(settings, GlobalSettings), "Global settings returned"
    assert fake_load_yaml.counter == 1, "Yaml loaded from file"


def test_get_external_addresses(monkeypatch):
    """
    .
    """
    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_load_yaml(file_path: str):
        """
        .
        """
        return {}

    monkeypatch.setattr(file_paths, "load_yaml_from_file", fake_load_yaml)

    addresses = root_parser._get_external_addresses()
    assert isinstance(addresses, ExternalAddresses), "External addresses returned"
    assert fake_load_yaml.counter == 1, "Yaml loaded from file"


def test_get_port_groups(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    monkeypatch.setattr(file_paths, "load_yaml_from_file", lambda path: [80, 443])
    monkeypatch.setattr(
        file_paths, "get_config_files", lambda folder: ["test-port-group"]
    )

    groups = root_parser._get_port_groups()
    assert isinstance(groups, list), "Groups is list"
    assert isinstance(groups[0], PortGroup), "List contains port groups"
    assert groups[0].name == "test-port-group", "Port group name set"
    assert groups[0].ports == [80, 443], "Port group ports set"


def test_create_from_config(monkeypatch):
    """
    .
    """

    @counter_wrapper
    def fake_load_global_settings():
        """
        .
        """
        return GlobalSettings()

    @counter_wrapper
    def fake_load_external_addresses():
        """
        .
        """
        return ExternalAddresses([])

    @counter_wrapper
    def fake_load_port_groups():
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
        root_parser, "_get_global_configuration", fake_load_global_settings
    )
    monkeypatch.setattr(
        root_parser, "_get_external_addresses", fake_load_external_addresses
    )
    monkeypatch.setattr(root_parser, "_get_port_groups", fake_load_port_groups)
    monkeypatch.setattr(
        file_paths, "get_folders_with_config", fake_get_folders_with_config
    )

    root_node = root_parser.RootNode.create_from_configs()
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
        [Network("Network", "1.1.1.1/24")],
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
        [Network("Network", "1.1.1.1/24")],
    )
    result = node.validation_failures()
    print(result)
    assert result == ["failure"] * 4, "Validation failures returned"
