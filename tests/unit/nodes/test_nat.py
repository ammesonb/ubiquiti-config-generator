"""
Test firewall
"""
from typing import List

from ubiquiti_config_generator import file_paths
from ubiquiti_config_generator.nodes import NAT, NATRule
from ubiquiti_config_generator.nodes.validatable import Validatable
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_nat_calls_methods(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_load_rules(self):
        """
        .
        """

    monkeypatch.setattr(NAT, "_load_rules", fake_load_rules)
    nat = NAT(".")

    assert fake_load_rules.counter == 1, "Load rules called"

    assert str(nat) == "NAT", "Name returned"
    assert getattr(nat, "auto-increment") == 10, "Auto increment set to default"
    assert getattr(nat, "rules") == [], "Rules set to empty array"


def test_load_rules(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(
        file_paths,
        "get_config_files",
        lambda path: ["nat/config.yaml", "nat/10.yaml", "nat/20.yaml", "nat/30.yaml",],
    )
    monkeypatch.setattr(file_paths, "load_yaml_from_file", lambda path: {})

    nat = NAT(".")
    assert len(nat.rules) == 3, "Three rules loaded"
    assert [rule.number for rule in nat.rules] == [
        "10",
        "20",
        "30",
    ], "Rule numbers correct"


def test_get_next_number():
    """
    .
    """
    properties = {
        "rules": [NATRule(11, "."), NATRule(100, ".")],
        "auto-increment": 100,
    }
    nat = NAT(".", **properties)
    assert getattr(nat, "auto-increment") == 100, "Auto increment overridden"
    assert nat.next_rule_number() == 200, "Next rule number correct"

    nat.add_rule({"number": 200, "config_path": "."})
    assert nat.next_rule_number() == 300, "Next rule number updated after adding rule"


def test_rules():
    """
    .
    """
    nat = NAT(".")
    nat.add_rule({"number": "10", "config_path": "."})
    nat.add_rule({"config_path": "."})
    assert [rule.number for rule in nat.rules] == ["10", 20,], "Rule numbers returned"


def test_validate(monkeypatch):
    """
    .
    """

    @counter_wrapper
    def fake_validate(self):
        """
        .
        """
        return True

    monkeypatch.setattr(NATRule, "validate", fake_validate)

    nat = NAT(".", rules=[NATRule(1, "."), NATRule(2, ".")],)
    assert nat.validate(), "NAT is valid"
    assert fake_validate.counter == 2, "Validation called for each rule"

    monkeypatch.setattr(NATRule, "validate", lambda self: False)
    assert not nat.validate(), "Rule validation fails"


def test_validation_failures(monkeypatch):
    """
    .
    """

    def fake_validation_errors(self) -> List[str]:
        """
        .
        """
        return ["problem"] if self.number == 10 else ["bad", "wrong"]

    monkeypatch.setattr(NATRule, "validation_errors", fake_validation_errors)
    monkeypatch.setattr(NAT, "validation_errors", lambda self: ["an error"])

    nat = NAT(".", rules=[NATRule(10, "."), NATRule(20, ".")],)
    assert nat.validation_failures() == [
        "an error",
        "problem",
        "bad",
        "wrong",
    ], "Validation failures are correct"


def test_commands(monkeypatch):
    """
    .
    """

    @counter_wrapper
    def rule_commands(self) -> List[str]:
        commands = []
        if rule_commands.counter == 1:
            commands = ["rule1-command1", "rule1-command2"]
        else:
            commands = ["rule2-command1", "rule2-command2", "rule2-command3"]

        return commands

    monkeypatch.setattr(NATRule, "commands", rule_commands)

    nat_properties = {
        "rules": [NATRule(10, "."), NATRule(20, ".")],
    }
    nat = NAT(".", **nat_properties)
    assert nat.commands() == [
        "rule1-command1",
        "rule1-command2",
        "rule2-command1",
        "rule2-command2",
        "rule2-command3",
    ], "Commands generated correctly"


def test_consistency():
    """
    .
    """
    nat = NAT(
        ".",
        rules=[
            NATRule(10, "."),
            NATRule(10, "."),
            NATRule(20, "."),
            NATRule(20, "."),
            NATRule(30, "."),
        ],
    )

    assert not nat.is_consistent(), "Not consistent"
    assert nat.validation_errors() == ["NAT has duplicate rules: 10, 20"]
