"""
Test firewall rule
"""

from ubiquiti_config_generator import secondary_configs
from ubiquiti_config_generator.nodes import Rule, PortGroup
from ubiquiti_config_generator.nodes.validatable import Validatable
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_non_address_commands():
    """
    .
    """
    rule_properties = {
        "action": "reject",
        "description": "Its a firewall rule",
        "log": "disable",
        "protocol": "tcp",
        "state": {
            "new": "disable",
            "invalid": "disable",
            "related": "enable",
            "established": "enable",
        },
    }
    commands = Rule(10, "firewall1", ".", **rule_properties).commands()

    rule_base = "firewall name firewall1 rule 10 "
    assert commands == [
        rule_base + "action reject",
        rule_base + "description 'Its a firewall rule'",
        rule_base + "log disable",
        rule_base + "protocol tcp",
        rule_base + "state new disable",
        rule_base + "state invalid disable",
        rule_base + "state related enable",
        rule_base + "state established enable",
    ], "Rule commands correct"


def test_non_group_address_commands():
    """
    .
    """
    rule_properties = {
        "description": "rule",
        "source": {"address": "10.10.10.10", "port": 8080},
        "destination": {"address": "8.8.8.8", "port": "443"},
    }
    commands = Rule(10, "firewall1", ".", **rule_properties).commands()
    rule_base = "firewall name firewall1 rule 10 "

    assert commands == [
        rule_base + "action accept",
        rule_base + 'description "rule"',
        rule_base + "source address 10.10.10.10",
        rule_base + "source port 8080",
        rule_base + "destination address 8.8.8.8",
        rule_base + "destination port 443",
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
    commands = Rule(10, "firewall1", ".", **rule_properties).commands()
    rule_base = "firewall name firewall1 rule 10 "

    assert commands == [
        rule_base + "action accept",
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

    # pylint: disable=unused-argument
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
    rule = Rule(10, "firewall", ".")
    assert not rule.validate(), "Validation fails if parent fails"
    assert fake_validate_false.counter == 1, "Parent validation called"

    monkeypatch.setattr(Validatable, "validate", fake_validate_true)
    rule = Rule(10, "firewall", ".")
    assert rule.validate(), "Validation can succeed if parent succeeds"
    assert fake_validate_true.counter == 1, "Parent validation called"

    rule = Rule(
        10, "firewall", ".", source={"port": "group1"}, destination={"port": "group2"}
    )
    assert rule.validate(), "Rule is valid if groups are valid"
    rule = Rule(
        10, "firewall", ".", source={"port": "group3"}, destination={"port": "group4"}
    )
    assert not rule.validate(), "Rule is invalid with nonexistent groups"
    assert rule.validation_errors() == [
        "Rule 10 has nonexistent source port group group3",
        "Rule 10 has nonexistent destination port group group4",
    ], "Errors added"


def test_string():
    """
    .
    """
    rule = Rule(123, "a-firewall", ".")
    assert str(rule) == "Firewall a-firewall rule 123", "String correct"
