"""
Configuration nodes
"""
from ubiquiti_config_generator.nodes.external_addresses import ExternalAddresses
from ubiquiti_config_generator.nodes.firewall import Firewall
from ubiquiti_config_generator.nodes.global_settings import GlobalSettings
from ubiquiti_config_generator.nodes.host import Host
from ubiquiti_config_generator.nodes.network import Network
from ubiquiti_config_generator.nodes.port_group import PortGroup
from ubiquiti_config_generator.nodes.rule import Rule

__all__ = [
    "ExternalAddresses",
    "Firewall",
    "GlobalSettings",
    "Host",
    "Network",
    "PortGroup",
    "Rule",
]
