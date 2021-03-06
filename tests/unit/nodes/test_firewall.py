"""
Test firewall
"""
from typing import List

from ubiquiti_config_generator import file_paths
from ubiquiti_config_generator.nodes import Firewall, Rule
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_firewall_calls_methods(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_set_attrs(self, attrs: dict):
        """
        .
        """

    monkeypatch.setattr(Firewall, "_add_keyword_attributes", fake_set_attrs)
    firewall = Firewall("firewall", "in", "network", ".")

    assert firewall.name == "firewall", "Name set"
    assert fake_set_attrs.counter == 1, "Set attrs called"

    assert str(firewall) == "Firewall firewall", "Name returned"
    assert getattr(firewall, "auto-increment") == 10, "Auto increment set to default"


def test_load_rules(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(
        file_paths,
        "get_config_files",
        lambda path: [
            "networks/network1/firewalls/firewall1/config.yaml",
            "networks/network1/firewalls/firewall1/10.yaml",
            "networks/network1/firewalls/firewall1/20.yaml",
            "networks/network1/firewalls/firewall1/30.yaml",
        ],
    )
    monkeypatch.setattr(file_paths, "load_yaml_from_file", lambda path: {})

    firewall = Firewall("firewall", "in", "network", ".")
    assert len(firewall.rules) == 3, "Three rules loaded"
    assert [rule.number for rule in firewall.rules] == [
        "10",
        "20",
        "30",
    ], "Rule numbers correct"


def test_get_next_number():
    """
    .
    """
    properties = {
        "rules": [Rule(11, "firewall", "."), Rule(100, "firewall", ".")],
        "auto-increment": 100,
    }
    firewall = Firewall("firewall", "in", "network", ".", **properties)
    assert getattr(firewall, "auto-increment") == 100, "Auto increment overridden"
    assert firewall.next_rule_number() == 200, "Next rule number correct"

    firewall.add_rule({"number": 200, "config_path": "."})
    assert (
        firewall.next_rule_number() == 300
    ), "Next rule number updated after adding rule"


def test_rules():
    """
    .
    """
    firewall = Firewall("firewall", "in", "network", ".")
    firewall.add_rule(
        {"number": "10", "action": "reject", "protocol": "tcp", "config_path": "."}
    )
    firewall.add_rule({"action": "reject", "protocol": "udp", "config_path": "."})
    assert [rule.number for rule in firewall.rules] == [
        "10",
        20,
    ], "Rule numbers returned"


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

    monkeypatch.setattr(Rule, "validate", fake_validate)

    firewall = Firewall(
        "firewall",
        "in",
        "network",
        ".",
        rules=[Rule(1, "firewall", "."), Rule(2, "firewall", ".")],
    )
    assert firewall.validate(), "Firewall is valid"
    assert fake_validate.counter == 2, "Validation called for each rule"

    monkeypatch.setattr(Rule, "validate", lambda self: False)
    assert not firewall.validate(), "Rule validation fails"


def test_validation_failures(monkeypatch):
    """
    .
    """

    def fake_validation_errors(self) -> List[str]:
        """
        .
        """
        return ["error"] if self.number == 10 else ["failure", "issue"]

    monkeypatch.setattr(Rule, "validation_errors", fake_validation_errors)
    monkeypatch.setattr(Firewall, "validation_errors", lambda self: ["an error"])

    firewall = Firewall(
        "firewall",
        "in",
        "network",
        ".",
        rules=[Rule(10, "firewall", "."), Rule(20, "firewall", ".")],
    )
    assert firewall.validation_failures() == [
        "an error",
        "error",
        "failure",
        "issue",
    ], "Validation failures are correct"


def test_commands(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def rule_commands(self) -> List[str]:
        commands = []
        if rule_commands.counter == 1:
            commands = ["rule1-command1", "rule1-command2"]
        else:
            commands = ["rule2-command1", "rule2-command2", "rule2-command3"]

        return commands

    monkeypatch.setattr(Rule, "commands", rule_commands)

    firewall_properties = {
        "default-action": "drop",
        "description": "A in-firewall description",
        "rules": [Rule(10, "firewall1", "."), Rule(20, "firewall1", ".")],
    }
    firewall = Firewall("firewall1", "in", "network1", ".", **firewall_properties)
    ordered_commands, command_list = firewall.commands()
    assert command_list == [
        "firewall name firewall1 default-action drop",
        "firewall name firewall1 description 'A in-firewall description'",
        "rule1-command1",
        "rule1-command2",
        "rule2-command1",
        "rule2-command2",
        "rule2-command3",
    ], "Commands generated correctly"

    assert ordered_commands == [
        [
            "firewall name firewall1 default-action drop",
            "firewall name firewall1 description 'A in-firewall description'",
        ],
        ["rule1-command1", "rule1-command2"],
        ["rule2-command1", "rule2-command2", "rule2-command3"],
    ], "Ordered commands correct"

    firewall = Firewall("firewall1", "in", "network1", ".", description="Description")
    ordered_commands, command_list = firewall.commands()
    assert command_list == [
        "firewall name firewall1 default-action accept",
        "firewall name firewall1 description 'Description'",
    ], "Description command is quoted"
