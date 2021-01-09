"""
Test configurations - loading, getting differences, etc
"""

from ubiquiti_config_generator import root_parser


def test_load_sample_config():
    """
    .
    """
    node = root_parser.RootNode.create_from_configs("sample_router_config")
