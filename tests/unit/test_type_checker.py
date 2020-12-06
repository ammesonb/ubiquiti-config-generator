"""
Test type checking validations
"""
from ubiquiti_config_generator import type_checker


def test_is_string_boolean():
    """
    .
    """
    assert not type_checker.is_string_boolean("true"), "String true isn't boolean"
    assert not type_checker.is_string_boolean("false"), "String false isn't boolean"
    assert not type_checker.is_string_boolean("0"), "String 0 isn't boolean"
    assert not type_checker.is_string_boolean("1"), "String 1 isn't boolean"
    assert not type_checker.is_string_boolean(0), "Int 0 isn't boolean"
    assert not type_checker.is_string_boolean(1), "Int 1 isn't boolean"
    assert not type_checker.is_string_boolean("other"), "Other string isn't boolean"
    assert not type_checker.is_string_boolean(
        [type_checker.ENABLED]
    ), "Array with enabled isn't boolean"
    assert not type_checker.is_string_boolean(
        {type_checker.ENABLED: type_checker.DISABLED}
    ), "Dictionary with enabed isn't boolean"
    assert type_checker.is_string_boolean(type_checker.ENABLED), "Enabled is boolean"
    assert type_checker.is_string_boolean(type_checker.DISABLED), "Disabled is boolean"


def test_is_ip_address():
    """
    .
    """
    assert not type_checker.is_ip_address("abc"), "String isn't IP"
    assert not type_checker.is_ip_address(123), "Number isn't IP"
    assert not type_checker.is_ip_address(["1.1.1.1"]), "Array of IP is not IP"
    assert not type_checker.is_ip_address({"1.1.1.1": "2.2.2.2"}), "Dict isn't an IP"
    assert not type_checker.is_ip_address("10.10.10"), "Missing octet"
    assert not type_checker.is_ip_address("10.10.10.10.10"), "Too many octets"
    assert not type_checker.is_ip_address("0.0.0.256"), "Outside bounds at end"
    assert not type_checker.is_ip_address(
        "256.255.255.255"
    ), "Outside bounds at beginning"
    assert type_checker.is_ip_address("255.255.255.255"), "Broadcast is an IP"
    assert type_checker.is_ip_address("192.168.0.1"), "Valid IP"


def test_subnet_mask():
    """
    .
    """
    assert not type_checker.is_subnet_mask("abc"), "String isn't subnet"
    assert not type_checker.is_subnet_mask([12]), "Array isn't subnet"
    assert not type_checker.is_subnet_mask({12: 21}), "Dict isn't subnet"
    assert not type_checker.is_subnet_mask("123"), "Mask too large"
    assert not type_checker.is_subnet_mask("-3"), "Negative mask not valid"
    assert type_checker.is_subnet_mask("0"), "Valid mask"
    assert type_checker.is_subnet_mask("24"), "Valid mask"
    assert type_checker.is_subnet_mask("32"), "Valid mask"


def test_is_cidr(monkeypatch):
    """
    .
    """
    assert not type_checker.is_cidr("1.1.1.1"), "Address not CIDR"
    assert not type_checker.is_cidr("1.1.1.1:80"), "Address with port not CIDR"

    monkeypatch.setattr(type_checker, "is_ip_address", lambda value: False)
    monkeypatch.setattr(type_checker, "is_subnet_mask", lambda value: False)
    assert not type_checker.is_cidr("1.1.1.1/24"), "IP and subnet fail not CIDR"

    monkeypatch.setattr(type_checker, "is_subnet_mask", lambda value: True)
    assert not type_checker.is_cidr("1.1.1.1/24"), "IP fail not CIDR"

    monkeypatch.setattr(type_checker, "is_ip_address", lambda value: True)
    assert type_checker.is_cidr("1.1.1.1/24"), "Is CIDR"

    monkeypatch.setattr(type_checker, "is_subnet_mask", lambda value: False)
    assert not type_checker.is_cidr("1.1.1.1/24"), "Subnet fail not CIDR"


def test_is_string():
    """
    .
    """
    assert not type_checker.is_string(1), "Number not string"
    assert not type_checker.is_string(["a"]), "Array not string"
    assert not type_checker.is_string({"a": "b"}), "Dict not string"
    assert type_checker.is_string("123"), "Is string"


def test_is_number():
    """
    .
    """
    assert not type_checker.is_number([1]), "Array not number"
    assert not type_checker.is_number({1: 2}), "Dict not number"
    assert not type_checker.is_number("abc"), "String is not number"
    assert type_checker.is_number("123"), "String is number"
    assert type_checker.is_number(1), "Is number"
    assert type_checker.is_number(-1), "Negative is number"


def test_is_duplex():
    """
    .
    """
    assert not type_checker.is_duplex("123"), "String is not duplex"
    assert not type_checker.is_duplex(123), "Number is not duplex"
    assert not type_checker.is_duplex([type_checker.AUTO]), "Array is not duplex"
    assert not type_checker.is_duplex(
        {type_checker.AUTO: True}
    ), "Dictionary is not duplex"
    assert type_checker.is_duplex(type_checker.AUTO), "Auto is duplex"
    assert type_checker.is_duplex(type_checker.FULL), "Full is duplex"
    assert type_checker.is_duplex(type_checker.HALF), "Half is duplex"


def test_is_speed():
    """
    .
    """
    assert not type_checker.is_speed("123"), "String is not speed"
    assert not type_checker.is_speed("10"), "String 10 is not speed"
    assert not type_checker.is_speed("100"), "String 100 is not speed"
    assert not type_checker.is_speed("1000"), "String 1000 is not speed"
    assert not type_checker.is_speed(123), "Number is not speed"
    assert not type_checker.is_speed(50), "50 is not speed"
    assert not type_checker.is_speed(128), "128 is not speed"
    assert not type_checker.is_speed(250), "250 is not speed"
    assert not type_checker.is_speed(256), "256 is not speed"
    assert not type_checker.is_speed(500), "500 is not speed"
    assert not type_checker.is_speed([type_checker.AUTO]), "Array is not speed"
    assert not type_checker.is_speed(
        {type_checker.AUTO: True}
    ), "Dictionary is not speed"
    assert type_checker.is_speed(type_checker.AUTO), "Auto is speed"
    assert type_checker.is_speed(10), "10 is speed"
    assert type_checker.is_speed(100), "100 is speed"
    assert type_checker.is_speed(1000), "1000 is speed"


def test_is_action():
    """
    .
    """
    assert not type_checker.is_action("123"), "String is not action"
    assert not type_checker.is_action(123), "Number is not action"
    assert not type_checker.is_action([type_checker.ACCEPT]), "Array is not action"
    assert not type_checker.is_action(
        {type_checker.ACCEPT: True}
    ), "Dictionary is not action"
    assert type_checker.is_action(type_checker.ACCEPT), "Accept is action"
    assert type_checker.is_action(type_checker.REJECT), "Reject is action"
    assert type_checker.is_action(type_checker.DROP), "Drop is action"
