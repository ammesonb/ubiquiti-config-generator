"""
Test host
"""
from ubiquiti_config_generator import secondary_configs
from ubiquiti_config_generator.nodes import Host, PortGroup
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_host_calls_methods(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_set_attrs(self, attrs: dict):
        """
        .
        """

    monkeypatch.setattr(Host, "_add_keyword_attributes", fake_set_attrs)
    host = Host("host", ".")

    assert host.name == "host", "Name set"
    assert fake_set_attrs.counter == 1, "Set attrs called"

    assert str(host) == "Host host", "Name returned"


def test_is_consistent(monkeypatch):
    """
    .
    """
    host = Host("host", ".")
    monkeypatch.setattr(secondary_configs, "get_port_groups", lambda config_path: [])
    assert host.is_consistent(), "Empty host is consistent"

    monkeypatch.setattr(
        secondary_configs, "get_port_groups", lambda config_path: [PortGroup("group1")]
    )
    attrs = {"forward-ports": ["group1", "group2"]}
    host = Host("host", ".", **attrs)
    assert not host.is_consistent(), "Missing forward port group inconsistent"
    assert host.validation_errors() == [
        "Port Group group2 not defined for forwarding in Host host"
    ], "Forward port group error set"

    attrs = {"forward-ports": ["group1", 80]}
    host = Host("host", ".", **attrs)
    assert host.is_consistent(), "Forward port groups consistent"
    assert not host.validation_errors(), "No errors for forward port"

    attrs = {"hairpin-ports": ["group1", "group2"]}
    host = Host("host", ".", **attrs)
    assert not host.is_consistent(), "Missing hairpin port group inconsistent"
    assert host.validation_errors() == [
        "Port Group group2 not defined for hairpin in Host host"
    ], "Hairpin port group error set"

    attrs = {"hairpin-ports": ["group1", 80]}
    host = Host("host", ".", **attrs)
    assert host.is_consistent(), "Hairpin port groups consistent"
    assert not host.validation_errors(), "No errors for hairpin port"

    attrs = {
        "connections": [
            {
                "allow": True,
                "source": {"address": "an-address"},
                "destination": {"port": 80},
            },
            {
                "allow": False,
                "source": {"address": "12.10.12.12", "port": 8080},
                "destination": {"port": "443"},
            },
            {
                "allow": False,
                "source": {"address": "group"},
                "destination": {"port": "group1"},
            },
            {
                "allow": False,
                "source": {"port": "group2"},
                "destination": {"port": "group3"},
            },
            {
                "allow": True,
                "source": {"address": "Amazon"},
                "destination": {"port": "web"},
            },
        ]
    }
    host = Host("host", ".", **attrs)
    assert not host.is_consistent(), "Connections inconsistent"
    assert host.validation_errors() == [
        "Source Port Group group2 not defined for blocked connection in Host host",
        "Destination Port Group group3 not defined for blocked connection in Host host",
        "Destination Port Group web not defined for allowed connection in Host host",
    ], "Error set for missing port group in connections"

    del attrs["connections"][-2:]
    host = Host("host", ".", **attrs)
    assert host.is_consistent(), "Connections are consistent"
    assert not host.validation_errors(), "No connection errors"
