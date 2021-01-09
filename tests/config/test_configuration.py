"""
Test configurations - loading, getting differences, etc
"""

from ubiquiti_config_generator import root_parser


def test_load_sample_config():
    """
    .
    """
    node = root_parser.RootNode.create_from_configs("sample_router_config")
    valid = node.validate()
    if not valid:
        [print(failure) for failure in node.validation_failures()]

    assert node.validate(), "Created node is valid"
    assert node.is_consistent(), "Created node is consistent"
    ordered_commands, command_list = node.get_commands()

    assert command_list == [], "Commands generated correctly"
