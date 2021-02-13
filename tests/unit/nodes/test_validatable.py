"""
Tests validatable object
"""

from ubiquiti_config_generator.nodes.validatable import Validatable


# pylint: disable=protected-access


def test_attributes_added():
    """
    .
    """
    obj = Validatable({"an-attribute": True}, ["an-attribute"])
    assert obj._validate_attributes == ["an-attribute"], "Attributes set"
    assert obj._validator_map == {"an-attribute": True}, "Map set"

    obj._add_validate_attribute("other-attribute")
    assert obj._validate_attributes == [
        "an-attribute",
        "other-attribute",
    ], "Attributes added"
    obj._add_keyword_attributes({"third-attr": 3, "fourth-attr": 4})
    assert obj._validate_attributes == [
        "an-attribute",
        "other-attribute",
        "third-attr",
        "fourth-attr",
    ], "Attributes updated"
    assert getattr(obj, "third-attr") == 3, "Third attribute set"
    assert getattr(obj, "fourth-attr") == 4, "Fourth attribute set"


def test_validate():
    """
    .
    """

    class MockValidatable(Validatable):
        """
        Mock validatable
        """

        def __str__(self):
            """
            .
            """
            return "Test node"

    obj = MockValidatable(
        {
            "valid-attr": lambda value: True,
            "invalid-attr": lambda value: False,
            "even-attr": lambda value: value % 2 == 0,
        }
    )

    assert obj.validate(), "No attributes is valid"
    obj._add_keyword_attributes({"valid-attr": "something"})
    assert obj.validate(), "Valid attr is valid"
    obj._add_keyword_attributes({"even-attr": 1})
    assert not obj.validate(), "Odd attr breaks validation"
    assert obj.validation_errors() == [
        "Test node attribute even-attr has failed validation"
    ], "Validation error added for failure"
    setattr(obj, "even-attr", 2)
    assert obj.validate(), "Even attr fixes validation"

    obj._add_validate_attribute("nonexistent-attr")
    assert not obj.validate(), "Nonexistent attribute fails"

    assert obj.validation_errors() == [
        "Test node attribute even-attr has failed validation",
        "Test node has attribute with no validation provided: 'nonexistent-attr'",
    ], "Validation error added for nonexistent attribute"

    obj._add_keyword_attributes({"invalid-attr": "value"})
    assert not obj.validate(), "Invalid attr breaks validation"

    assert obj.validation_errors() == [
        "Test node attribute even-attr has failed validation",
        "Test node has attribute with no validation provided: 'nonexistent-attr'",
        "Test node has attribute with no validation provided: 'nonexistent-attr'",
        "Test node attribute invalid-attr has failed validation",
    ], "Validation error added for invalid attribute"


def test_equals():
    """
    .
    """
    valid = Validatable({"stuff": lambda: True})
    valid._add_keyword_attributes({"stuff": 123})
    valid2 = Validatable({"stuff": lambda: True})
    valid2._add_keyword_attributes({"stuff": 234})

    assert valid != 123, "Not equal to int"
    assert valid != "123", "Not equal to str"
    assert valid != [123], "Not equal to list"
    assert valid != {"stuff": 123}, "Not equal to dict"
    assert valid != valid2, "Not equal to other valid with different value"

    valid2.stuff = 123
    assert valid == valid2, "Is equal, with correct value"
    assert [valid] == [valid2], "Is equal in list as well"
