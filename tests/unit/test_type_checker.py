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


def test_is_mac():
    """
    .
    """
    assert not type_checker.is_mac("abc"), "Random string not mac"
    assert not type_checker.is_mac(123), "Number is not mac"
    assert not type_checker.is_mac(["aa:bb:cc:dd:ee:ff"]), "Array is not mac"
    assert not type_checker.is_mac({"aa:bb:cc:dd:ee:ff": True}), "Dictionary is not mac"
    assert not type_checker.is_mac("aa:bb:cc:dd:ee"), "Missing octet is not mac"
    assert not type_checker.is_mac("aa:bb:cc:dd:ee:ee:ff"), "Extra octet is not mac"
    assert not type_checker.is_mac(
        "aabbccddeeff"
    ), "No separator for valid MAC is not mac"
    assert not type_checker.is_mac(
        "aa bb cc dd ee ff"
    ), "Space for valid MAC is not mac"
    assert type_checker.is_mac("aa-bb-cc-dd-ee-ff"), "Dashes is MAC"
    assert type_checker.is_mac("aa:bb:cc:dd:ee:ff"), "Colon is MAC"
    assert type_checker.is_mac("aa:bb-cc:dd:ee-ff"), "Mixed dash and colon is MAC"


def test_is_translated_port():
    """
    .
    """
    assert not type_checker.is_translated_port("abc"), "String not port"
    assert not type_checker.is_translated_port(80), "Number is not port"
    assert not type_checker.is_translated_port([80]), "Array is not port"
    assert not type_checker.is_translated_port(
        {"abc": 80}
    ), "String key is not port translation"
    assert not type_checker.is_translated_port(
        {"80": "abc"}
    ), "String value is not port translation"
    assert not type_checker.is_translated_port(
        {80: 8080, 443: 4430}
    ), "Multiple keys is not port translation"
    assert type_checker.is_translated_port(
        {"80": "8080"}
    ), "String numbers are port translation"
    assert type_checker.is_translated_port(
        {"80": 8080}
    ), "String key with int value is port translation"
    assert type_checker.is_translated_port(
        {80: "8080"}
    ), "Int key with str value is port translation"
    assert type_checker.is_translated_port(
        {80: 8080}
    ), "Int key with int value is port translation"


def test_is_address_and_or_port():
    """
    .
    """
    assert not type_checker.is_address_and_or_port("abc"), "String not valid"
    assert not type_checker.is_address_and_or_port(80), "Number is not valid"
    assert not type_checker.is_address_and_or_port([80]), "Array is not valid"
    assert not type_checker.is_address_and_or_port({"abc": 123}), "Dict is not valid"
    assert not type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: 123}
    ), "Non-array address invalid"
    assert not type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: [123]}
    ), "Non-string address invalid"
    assert not type_checker.is_address_and_or_port(
        {type_checker.PORT: "123"}
    ), "Non-array port invalid"
    assert not type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: [123], type_checker.PORT: [123]}
    ), "Number port with number address invalid"
    assert not type_checker.is_address_and_or_port(
        {"other": "foo", type_checker.ADDRESS: ["123"], type_checker.PORT: [123]}
    ), "Extra field at beginning invalid"
    assert not type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: ["123"], type_checker.PORT: [123], "other": 123}
    ), "Extra field at end invalid"
    assert not type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: ["123"], type_checker.PORT: [123], "other": 123}
    ), "Extra field at end invalid"

    assert not type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: ["123", 80], type_checker.PORT: [123]}
    ), "Mixed address type invalid"
    assert not type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: ["123", 123], type_checker.PORT: [123, "abc"]}
    ), "Mixed address and port type invalid"

    assert type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: ["123"]}
    ), "Single address is valid"
    assert type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: ["123", "234"]}
    ), "Multiple address is valid"

    assert type_checker.is_address_and_or_port(
        {type_checker.PORT: [80]}
    ), "Single port is valid"
    assert type_checker.is_address_and_or_port(
        {type_checker.PORT: [80, 443]}
    ), "Multiple port is valid"

    assert type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: ["123"], type_checker.PORT: [80]}
    ), "Single address/port combination is valid"
    assert type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: ["123", "234"], type_checker.PORT: [80, 443, "ssh"]}
    ), "Multiple address/port combination is valid"


def test_is_firewall_direction():
    """
    .
    """
    assert not type_checker.is_firewall_direction("abc"), "String not valid"
    assert not type_checker.is_firewall_direction(80), "Number is not valid"
    assert not type_checker.is_firewall_direction(["in"]), "Array is not valid"
    assert not type_checker.is_firewall_direction(
        {"direction": "out"}
    ), "Dict is not valid"
    assert type_checker.is_firewall_direction("in"), "In is direction"
    assert type_checker.is_firewall_direction("out"), "Out is direction"
    assert type_checker.is_firewall_direction("local"), "Local is direction"
