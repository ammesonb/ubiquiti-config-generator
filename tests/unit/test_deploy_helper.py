"""
Deploy helper functionality testing
"""

from ubiquiti_config_generator import deploy_helper


def test_compare_commands():
    """
    .
    """
    current = [
        "command 1",
        "firewall 2",
        'description "a command description"',
        "some stuff",
        "foo bar",
    ]

    previous = [
        "some thing",
        'description "a command description"',
        "foo baz",
        "firewall 2",
        "obsolete interface",
    ]

    difference = deploy_helper.diff_configurations(current, previous)
    assert difference.added == {"command": "1"}, "Added correct"
    assert difference.removed == {"obsolete": "interface"}, "Removed correct"
    assert difference.changed == {"some": "stuff", "foo": "bar"}, "Changed correct"
    assert difference.preserved == {
        "description": "a command description",
        "firewall": "2",
    }, "Preserved correct"
