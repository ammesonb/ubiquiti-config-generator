"""
Check config difference works as expected
"""

from ubiquiti_config_generator import deploy_helper


def test_atomic_operations():
    """
    .
    """
    difference = deploy_helper.ConfigDifference()

    difference.add({"a": 1})
    difference.remove({"b": 2})
    difference.change({"c": 3})
    difference.preserve({"d": 4})

    assert difference.added == {"a": 1}, "Added set"
    assert difference.removed == {"b": 2}, "Removed set"
    assert difference.changed == {"c": 3}, "Changed set"
    assert difference.preserved == {"d": 4}, "Preserved set"

    difference.add({"a": 9})
    difference.remove({"b": 8})
    difference.change({"c": 7})
    difference.preserve({"d": 6})

    assert difference.added == {"a": 9}, "Added updated"
    assert difference.removed == {"b": 8}, "Removed updated"
    assert difference.changed == {"c": 7}, "Changed updated"
    assert difference.preserved == {"d": 6}, "Preserved updated"


def test_compare():
    """
    .
    """
    diff = deploy_helper.ConfigDifference()

    commands = {"a": 1, "b": 2, "d": 4, "e": 5}

    previous = {"a": 9, "b": 8, "c": 7, "e": 5}

    for command, value in commands.items():
        diff.compare_commands(
            {command: value},
            {command: previous[command]} if command in previous else None,
        )

    for command, value in previous.items():
        diff.compare_commands(
            {command: commands[command]} if command in commands else None,
            {command: value},
        )

    assert diff.added == {"d": 4}, "Added set"
    assert diff.removed == {"c": 7}, "Removed set"
    assert diff.changed == {"a": 1, "b": 2}, "Changed set"
    assert diff.preserved == {"e": 5}, "Preserved set"
