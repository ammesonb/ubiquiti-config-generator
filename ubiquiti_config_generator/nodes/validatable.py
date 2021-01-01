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
        self._validation_errors = []

    def validate(self) -> bool:
        """
        Validate this object
        """
        valid = True
        for attribute in self._validate_attributes:
            instance_valid = attribute in self._validator_map and self._validator_map[
                attribute
            ](getattr(self, attribute))

            valid = valid and instance_valid

            if attribute not in self._validator_map:
                self.add_validation_error(
                    "{0} has attribute with no validation provided: '{1}'".format(
                        str(self), attribute
                    )
                )
            elif not instance_valid:
                self.add_validation_error(
                    "{0} attribute {1} has failed validation".format(
                        str(self), attribute
                    )
                )

        return valid

    def _add_validate_attribute(self, attribute: str) -> None:
        """
        Add a validatable attribute
        """
        if attribute not in self._validate_attributes:
            self._validate_attributes.append(attribute)

    def _add_keyword_attributes(self, kwargs: dict) -> None:
        """
        Adds all keyword arguments
        """
        for option, value in kwargs.items():
            self._add_validate_attribute(option)
            setattr(self, option, value)

    def validation_errors(self) -> List[str]:
        """
        Validation errors
        """
        return self._validation_errors

    def add_validation_error(self, error: str):
        """
        Adds a validation error
        """
        self._validation_errors.append(error)
