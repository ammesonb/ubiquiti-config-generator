"""
Test host
"""
import copy

from ubiquiti_config_generator import secondary_configs
from ubiquiti_config_generator.nodes import (
    Host,
    PortGroup,
    Network,
    NAT,
    Firewall,
    Rule,
    NATRule,
)
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
    host = Host("host", None, ".", "192.168.0.1")

    assert host.name == "host", "Name set"
    assert fake_set_attrs.counter == 1, "Set attrs called"

    assert str(host) == "Host host", "Name returned"


def test_is_consistent(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(Host, "add_firewall_rules", lambda self: None)
    network = Network("network", None, ".", "192.168.0.0/24", "eth0")
    host = Host("host", network, ".", "192.168.0.1")
    monkeypatch.setattr(secondary_configs, "get_port_groups", lambda config_path: [])
    assert host.is_consistent(), "Empty host is consistent"

    monkeypatch.setattr(
        secondary_configs, "get_port_groups", lambda config_path: [PortGroup("group1")]
    )
    attrs = {"forward-ports": ["group1", "group2"]}
    host = Host("host", network, ".", "192.168.0.1", **attrs)
    assert not host.is_consistent(), "Missing forward port group inconsistent"
    assert host.validation_errors() == [
        "Port Group group2 not defined for forwarding in Host host"
    ], "Forward port group error set"

    attrs = {"forward-ports": ["group1", 80]}
    host = Host("host", network, ".", "192.168.0.1", **attrs)
    assert host.is_consistent(), "Forward port groups consistent"
    assert not host.validation_errors(), "No errors for forward port"

    attrs = {"hairpin-ports": ["group1", "group2"]}
    host = Host("host", network, ".", "192.168.0.1", **attrs)
    assert not host.is_consistent(), "Missing hairpin port group inconsistent"
    assert host.validation_errors() == [
        "Port Group group2 not defined for hairpin in Host host"
    ], "Hairpin port group error set"

    attrs = {"hairpin-ports": ["group1", 80]}
    host = Host("host", network, ".", "192.168.0.2", **attrs)
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
                "source": {"address": "192.168.0.0/24"},
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
    host = Host("host", network, ".", "192.168.0.1", **attrs)
    assert not host.is_consistent(), "Connections inconsistent"
    assert host.validation_errors() == [
        "Source Port Group group2 not defined for blocked connection in Host host",
        "Destination Port Group group3 not defined for blocked connection in Host host",
        "Destination Port Group web not defined for allowed connection in Host host",
    ], "Error set for missing port group in connections"

    del attrs["connections"][-2:]
    host = Host("host", network, ".", "192.168.0.1", **attrs)
    assert host.is_consistent(), "Connections are consistent"
    assert not host.validation_errors(), "No connection errors"


def test_host_firewall_consistency(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(secondary_configs, "get_port_groups", lambda config_path: [])
    monkeypatch.setattr(Host, "add_firewall_rules", lambda self: None)

    network = Network("network", None, ".", "192.168.0.0/24", "eth0")

    attrs = {
        "address-groups": ["an-address"],
        "connections": [
            {
                "allow": True,
                "rule": 10,
                "source": {"address": "an-address"},
                "destination": {"port": 80},
            },
            {
                "allow": False,
                "rule": 10,
                "source": {"address": "12.10.12.12", "port": 8080},
            },
            {
                "allow": False,
                "rule": 20,
                "source": {"address": "an-address"},
                "destination": {"port": 443},
            },
            {
                "allow": False,
                "rule": 20,
                "source": {"address": "an-address"},
                "destination": {"port": 4443},
            },
        ],
    }
    host = Host("host", network, ".", "12.10.12.12", **attrs)
    assert not host.is_consistent(), "Duplicate rules inconsistent"
    assert host.validation_errors() == [
        "Host host has duplicate firewall rules: 10, 20"
    ], "Duplicate rules errors set"

    firewall_in = Firewall(
        "firewall-in",
        "in",
        "network",
        ".",
        rules=[Rule(10, "firewall-in", "."), Rule(30, "firewall-in", ".")],
    )
    firewall_out = Firewall(
        "firewall-out",
        "out",
        "network",
        ".",
        rules=[Rule(20, "firewall-out", "."), Rule(40, "firewall-out", ".")],
    )
    network = Network(
        "network",
        None,
        ".",
        "192.168.0.0/24",
        "eth0",
        firewalls=[firewall_in, firewall_out],
    )
    attrs = {
        "address-groups": ["an-address"],
        "connections": [
            {
                "allow": True,
                "rule": 10,
                "source": {"address": "an-address"},
                "destination": {"port": 80},
            },
            {
                "allow": False,
                "rule": 20,
                "source": {"address": "an-address"},
                "destination": {"port": 443},
            },
        ],
    }
    host = Host("host", network, ".", "192.168.0.1", **attrs)
    assert not host.is_consistent(), "Rules conflicting with firewall is inconsistent"
    assert host.validation_errors() == [
        "Host host has conflicting connection rule with Firewall firewall-in, "
        "rule number 10",
        "Host host has conflicting connection rule with Firewall firewall-out, "
        "rule number 20",
    ], "Firewall rule conflict errors set"


def test_connection_consistency(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(
        secondary_configs, "get_port_groups", lambda config_path: [PortGroup("a-group")]
    )
    monkeypatch.setattr(Host, "add_firewall_rules", lambda self: None)

    network = Network("network", None, ".", "192.168.0.0/24", "eth0")

    attributes = {
        "connections": [
            {"allow": True, "rule": 10},
            {
                "allow": True,
                "rule": 20,
                "source": {"address": "123.123.123.123"},
                "destination": {"address": "a-group"},
            },
        ]
    }
    host = Host("host", network, ".", "192.168.0.1", **attributes)

    assert not host.is_consistent(), "Host is not consistent"
    assert host.validation_errors() == [
        "Host host has connection with no source address or destination address",
        "Host host has connection where its address is not used "
        "in source or destination",
    ]


def test_add_firewall_rules(monkeypatch):
    """
    .
    """
    host_properties = {
        "address-groups": ["devices"],
        "forward-ports": [
            80,
            "443",
            "server-ports",
            {"8080": 80},
            {4443: 443},
            {222: "22"},
        ],
        "hairpin-ports": [
            {
                "connection": {"destination": {"address": "8.8.8.8", "port": 53}},
                "interface": "eth1.10",
            },
            {
                "connection": {"destination": {"port": "server-ports"}},
                "description": "Hairpin server ports back to host",
                "interface": "eth0",
            },
        ],
        "connections": [
            {
                "allow": True,
                "rule": 20,
                "log": False,
                "source": {"address": "192.168.0.0/24",},
                "destination": {"port": "printer-ports"},
            },
            {
                "allow": False,
                "log": True,
                "source": {"address": "bad-things",},
                "destination": {"address": "devices", "port": 22},
            },
            {
                "allow": True,
                "source": {"address": "devices",},
                "destination": {"port": "allowed-ports"},
            },
            {
                "allow": False,
                "log": True,
                "source": {"address": "devices",},
                "destination": {"port": "blocked-ports"},
            },
            {
                "allow": True,
                "log": False,
                "source": {"port": "allowed-ports",},
                "destination": {"address": "devices"},
            },
        ],
    }

    host = Host(
        "host1",
        Network("network", NAT("."), ".", "192.168.0.0/24", "eth0"),
        ".",
        "192.168.0.100",
        **host_properties
    )

    firewall_in = host.network.firewalls_by_direction["in"]
    firewall_out = host.network.firewalls_by_direction["out"]

    rule_20 = {
        "number": 20,
        "firewall_name": "network-IN",
        "config_path": ".",
        "action": "accept",
        "protocol": "tcp_udp",
        "log": "disable",
        "source": {"address": "192.168.0.0/24"},
        "destination": {"port": "printer-ports"},
    }

    rule_10 = copy.deepcopy(rule_20)
    rule_10["number"] = 10
    rule_10["source"]["address"] = "devices"
    rule_10["destination"]["port"] = "allowed-ports"

    rule_30 = copy.deepcopy(rule_10)
    rule_30["number"] = 30
    rule_30["action"] = "drop"
    rule_30["log"] = "enable"
    rule_30["destination"]["port"] = "blocked-ports"

    assert firewall_in.rules == [
        Rule(**rule_20),
        Rule(**rule_10),
        Rule(**rule_30),
    ], "Rules created for in firewall correctly"

    rule_10["action"] = "drop"
    rule_10["source"]["address"] = "bad-things"
    rule_10["destination"]["address"] = "devices"
    rule_10["destination"]["port"] = 22
    rule_10["log"] = "enable"

    rule_20 = copy.deepcopy(rule_10)
    rule_20["number"] = 20
    rule_20["action"] = "accept"
    rule_20["log"] = "disable"
    rule_20["source"] = {"port": "allowed-ports"}
    rule_20["destination"] = {"address": "devices"}

    assert firewall_out.rules == [
        Rule(**rule_10),
        Rule(**rule_20),
    ], "Rules created for out firewall correctly"

    base_rule_properties = {
        "number": 10,
        "description": "Forward port 80 to host1",
        "config_path": ".",
        "type": "destination",
        "protocol": "tcp_udp",
        "inbound-interface": "eth0",
        "inside-address": {"address": "192.168.0.100"},
        "destination": {"port": 80},
    }

    ssl_rule = copy.deepcopy(base_rule_properties)
    ssl_rule["number"] = 20
    ssl_rule["destination"]["port"] = "443"
    ssl_rule["description"] = ssl_rule["description"].replace("80", "443")

    server_rule = copy.deepcopy(base_rule_properties)
    server_rule["number"] = 30
    server_rule["destination"]["port"] = "server-ports"
    server_rule["description"] = server_rule["description"].replace(
        "80", "server-ports"
    )

    str_translate_rule = copy.deepcopy(base_rule_properties)
    str_translate_rule["number"] = 40
    str_translate_rule["destination"]["port"] = "8080"
    str_translate_rule["inside-address"]["port"] = 80
    str_translate_rule["description"] = str_translate_rule["description"].replace(
        "80", "8080"
    )

    int_translate_rule = copy.deepcopy(base_rule_properties)
    int_translate_rule["number"] = 50
    int_translate_rule["destination"]["port"] = 4443
    int_translate_rule["inside-address"]["port"] = 443
    int_translate_rule["description"] = int_translate_rule["description"].replace(
        "80", "4443"
    )

    other_translate_rule = copy.deepcopy(base_rule_properties)
    other_translate_rule["number"] = 60
    other_translate_rule["destination"]["port"] = 222
    other_translate_rule["inside-address"]["port"] = "22"
    other_translate_rule["description"] = other_translate_rule["description"].replace(
        "80", "222"
    )

    hairpin_rule = {
        "number": 70,
        "description": "",
        "config_path": ".",
        "type": "destination",
        "protocol": "tcp_udp",
        "inside-address": {"address": "192.168.0.100"},
        "source": {},
        "destination": {"address": "8.8.8.8", "port": 53},
        "inbound-interface": "eth1.10",
    }

    hairpin_rule_2 = copy.deepcopy(hairpin_rule)
    hairpin_rule_2["number"] = 80
    hairpin_rule_2["description"] = "Hairpin server ports back to host"
    hairpin_rule_2["destination"] = {"port": "server-ports"}
    hairpin_rule_2["inbound-interface"] = "eth0"

    assert host.network.nat.rules == [
        NATRule(**base_rule_properties),
        NATRule(**ssl_rule),
        NATRule(**server_rule),
        NATRule(**str_translate_rule),
        NATRule(**int_translate_rule),
        NATRule(**other_translate_rule),
        NATRule(**hairpin_rule),
        NATRule(**hairpin_rule_2),
    ], "NAT rules created"
