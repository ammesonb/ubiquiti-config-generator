"""
Test loading of secondary configs
"""
from ubiquiti_config_generator import secondary_configs, file_paths
from ubiquiti_config_generator.nodes import GlobalSettings, PortGroup, ExternalAddresses
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

    settings = secondary_configs.get_global_configuration(".")
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

    addresses = secondary_configs.get_external_addresses(".")
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

    groups = secondary_configs.get_port_groups(".")
    assert isinstance(groups, list), "Groups is list"
    assert isinstance(groups[0], PortGroup), "List contains port groups"
    assert groups[0].name == "test-port-group", "Port group name set"
    assert groups[0].ports == [80, 443], "Port group ports set"
