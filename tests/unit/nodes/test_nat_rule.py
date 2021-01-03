"""
Test firewall rule
"""

from ubiquiti_config_generator import secondary_configs
from ubiquiti_config_generator.nodes import NATRule, PortGroup
from ubiquiti_config_generator.nodes.validatable import Validatable
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_non_address_commands():
    """
    .
    """
    rule_properties = {
        "description": "Its a rule",
        "log": "disable",
        "protocol": "tcp",
        "type": "source",
        "inbound-interface": "eth1.10",
        "outbound-interface": "eth0",
    }
    commands = NATRule(10, ".", **rule_properties).commands()

    rule_base = "service nat rule 10 "
    assert commands == [
        rule_base + "description 'Its a rule'",
        rule_base + "log disable",
        rule_base + "protocol tcp",
        rule_base + "type source",
        rule_base + "inbound-interface eth1.10",
        rule_base + "outbound-interface eth0",
    ], "Rule commands correct"


def test_non_group_address_commands():
    """
    .
    """
    rule_properties = {
        "description": "rule",
        "source": {"address": "10.10.10.10", "port": 8080},
        "destination": {"address": "8.8.8.8", "port": "443"},
        "inside-address": {"address": "192.168.0.2", "port": 80},
    }
    commands = NATRule(10, ".", **rule_properties).commands()
    rule_base = "service nat rule 10 "

    assert commands == [
        rule_base + 'description "rule"',
        rule_base + "source address 10.10.10.10",
        rule_base + "source port 8080",
        rule_base + "destination address 8.8.8.8",
        rule_base + "destination port 443",
        rule_base + "inside-address address 192.168.0.2",
        rule_base + "inside-address port 80",
    ], "Rule commands correct"


def test_group_address_commands():
    """
    .
    """
    rule_properties = {
        "description": "rule",
        "source": {"address": "a-group", "port": "web-ports"},
        "destination": {"address": "the.whole.internet", "port": "any"},
    }
    commands = NATRule(10, ".", **rule_properties).commands()
    rule_base = "service nat rule 10 "

    assert commands == [
        rule_base + 'description "rule"',
        rule_base + "source group address-group a-group",
        rule_base + "source group port-group web-ports",
        rule_base + "destination group address-group the.whole.internet",
        rule_base + "destination group port-group any",
    ], "Rule commands correct"


def test_validate(monkeypatch):
    """
    .
    """

    @counter_wrapper
    def fake_validate_false(self) -> bool:
        """
        .
        """
        return False

    @counter_wrapper
    def fake_validate_true(self) -> bool:
        """
        .
        """
        return True

    monkeypatch.setattr(
        secondary_configs,
        "get_port_groups",
        lambda config_path: [PortGroup("group1"), PortGroup("group2")],
    )

    monkeypatch.setattr(Validatable, "validate", fake_validate_false)
    rule = NATRule(10, ".")
    assert not rule.validate(), "Validation fails if parent fails"
    assert fake_validate_false.counter == 1, "Parent validation called"

    monkeypatch.setattr(Validatable, "validate", fake_validate_true)
    rule = NATRule(10, ".")
    assert not rule.validate(), "Validation fails without inside address"
    assert fake_validate_true.counter == 1, "Parent validation called"
    assert rule.validation_errors() == [
        "NAT rule 10 does not have inside address"
    ], "Validation errors set"

    properties = {
        "source": {"port": "group1"},
        "destination": {"port": "group2"},
        "inside-address": {"address": "192.168.0.2"},
    }
    rule = NATRule(10, ".", **properties)

    rule.validate()
    print(rule.validation_errors())
    assert rule.validate(), "Rule is valid if groups are valid"
    rule = NATRule(10, ".", **properties)

    properties["source"]["port"] = "group3"
    properties["destination"]["port"] = "group4"
    assert not rule.validate(), "Rule is invalid with nonexistent groups"
    assert rule.validation_errors() == [
        "NAT rule 10 has nonexistent source port group group3",
        "NAT rule 10 has nonexistent destination port group group4",
    ], "Errors added"
