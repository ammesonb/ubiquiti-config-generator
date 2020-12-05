"""
Configuration nodes
"""
from ubiquiti_config_generator.nodes.global_settings import GlobalSettings
from ubiquiti_config_generator.nodes.external_addresses import ExternalAddresses
from ubiquiti_config_generator.nodes.network import Network
from ubiquiti_config_generator.nodes.port_group import PortGroup

__all__ = ["GlobalSettings", "ExternalAddresses", "Network", "PortGroup"]
