"""
A firewall rule
"""


class Rule:
    """
    Represents a firewall rule
    """

    def __init__(self, number: int, **kwargs):
        self.number = number

        for key, arg in kwargs.items():
            pass
