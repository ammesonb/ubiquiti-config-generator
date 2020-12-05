"""
Contains generic validation functions
"""
from typing import List


class Validatable:
    """
    A validatable node
    """

    def __init__(self, validator_map: dict, attributes: List[str] = None):
        self._validate_attributes = attributes or []
        self._validator_map = validator_map

    def validate(self) -> bool:
        """
        Validate this object
        """
        valid = True
        for attribute in self._validate_attributes:
            valid = (
                valid
                and attribute in self._validator_map
                and self._validator_map[attribute](getattr(self, attribute))
            )
            if not valid:
                break

        return valid

    def _add_validate_attribute(self, attribute: str) -> None:
        """
        Add a validatable attribute
        """
        self._validate_attributes.append(attribute)

    def _add_keyword_attributes(self, kwargs: dict) -> None:
        """
        Adds all keyword arguments
        """
        for option, value in kwargs.items():
            self._add_validate_attribute(option)
            setattr(self, option, value)
