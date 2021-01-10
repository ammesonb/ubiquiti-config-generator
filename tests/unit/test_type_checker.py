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
        [type_checker.ENABLE]
    ), "Array with enable isn't boolean"
    assert not type_checker.is_string_boolean(
        {type_checker.ENABLE: type_checker.DISABLE}
    ), "Dictionary with enable isn't boolean"
    assert type_checker.is_string_boolean(type_checker.ENABLE), "Enable is boolean"
    assert type_checker.is_string_boolean(type_checker.DISABLE), "Disable is boolean"


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


def test_is_name():
    """
    .
    """
    assert not type_checker.is_name(80), "Number is not valid"
    assert not type_checker.is_name(["in"]), "Array is not valid"
    assert not type_checker.is_name({"abc": "def"}), "Dictionary is not valid"
    assert type_checker.is_name("abc"), "String is valid"
    assert type_checker.is_name("abc123"), "String with numbers is valid"
    assert type_checker.is_name("123host"), "String starting with numbers is valid"
    assert type_checker.is_name("valid-host"), "Name with dashes is valid"
    assert type_checker.is_name("valid_host"), "Host with underscores is valid"
    assert type_checker.is_name(
        "valid_host-new"
    ), "Host with underscores and dash is valid"
    assert type_checker.is_name(
        "valid_host-new-1"
    ), "Host with underscores, dash, and number is valid"

    assert not type_checker.is_name("Invalid#"), "Hash is invalid"
    assert not type_checker.is_name("Invalid!"), "Exclamation mark invalid"
    assert not type_checker.is_name("Invalid@"), "At symbol invalid"
    assert not type_checker.is_name("Invalid^"), "Caret invalid"
    assert not type_checker.is_name("Invalid&"), "Ampersand invalid"
    assert not type_checker.is_name("Invalid*"), "Asterisk invalid"
    assert not type_checker.is_name("Invalid("), "Paren invalid"
    assert not type_checker.is_name("Invalid<"), "Less than invalid"
    assert not type_checker.is_name("Invalid."), "Period invalid"
    assert not type_checker.is_name("Invalid,"), "Comma invalid"
    assert not type_checker.is_name("Invalid?"), "Question mark invalid"
    assert not type_checker.is_name("Invalid/"), "Slash invalid"
    assert not type_checker.is_name("Invalid\\"), "Backslash invalid"
    assert not type_checker.is_name("Invalid;"), "Semicolon invalid"
    assert not type_checker.is_name("Invalid:"), "Colon invalid"
    assert not type_checker.is_name("Invalid$"), "Dollar invalid"
    assert not type_checker.is_name("Invalid%"), "Percent invalid"
    assert not type_checker.is_name("Invalid+"), "Plus invalid"
    assert not type_checker.is_name("Invalid="), "Equals invalid"


def test_is_description():
    """
    .
    """
    assert not type_checker.is_description(80), "Number is not valid"
    assert not type_checker.is_description(["in"]), "Array is not valid"
    assert not type_checker.is_description({"abc": "def"}), "Dictionary is not valid"
    assert type_checker.is_description("abc"), "String is valid"
    assert type_checker.is_description("abc123"), "String with numbers is valid"
    assert type_checker.is_description(
        "123host"
    ), "String starting with numbers is valid"
    assert type_checker.is_description("valid-host"), "Description with dashes is valid"
    assert type_checker.is_description("valid_host"), "Underscores is valid"
    assert type_checker.is_description(
        "valid_host!@#$%^&*()_+-="
    ), "Various punctuation is valid"
    assert not type_checker.is_description("it's a host"), "Single quote invalid"
    assert not type_checker.is_description('it\'s a "host"'), "Double quote invalid"


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
    # assert not type_checker.is_address_and_or_port(
    # {type_checker.PORT: "123"}
    # ), "Non-array port invalid"
    assert not type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: [123], type_checker.PORT: [123]}
    ), "Number port with number address invalid"
    assert not type_checker.is_address_and_or_port(
        {"other": "foo", type_checker.ADDRESS: "123", type_checker.PORT: 123}
    ), "Extra field at beginning invalid"
    assert not type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: "123", type_checker.PORT: 123, "other": 123}
    ), "Extra field at end invalid"

    assert type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: "123"}
    ), "Single address is valid"

    assert type_checker.is_address_and_or_port(
        {type_checker.PORT: 80}
    ), "Single port is valid"

    assert type_checker.is_address_and_or_port(
        {type_checker.ADDRESS: "123", type_checker.PORT: 80}
    ), "Single address/port combination is valid"


def test_is_source_destination():
    """
    .
    """
    assert not type_checker.is_source_destination("abc"), "String not valid"
    assert not type_checker.is_source_destination(80), "Number is not valid"
    assert not type_checker.is_source_destination(["allow"]), "Array is not valid"
    assert not type_checker.is_source_destination({"allow": "bad"}), "Dict is not valid"

    assert not type_checker.is_source_destination(
        {"allow": True}
    ), "Missing source/destination invalid"
    assert not type_checker.is_source_destination(
        {"allow": True, "source": []}
    ), "Non dictionary source invalid"
    assert not type_checker.is_source_destination(
        {"allow": True, "destination": "Abc"}
    ), "Non dictionary destination invalid"
    assert not type_checker.is_source_destination(
        {"allow": True, "source": {}}
    ), "Empty source invalid"
    assert not type_checker.is_source_destination(
        {"allow": True, "destination": {}}
    ), "Empty destination invalid"
    assert not type_checker.is_source_destination(
        {"allow": True, "source": {}, "destination": {}}
    ), "Empty source and destination invalid"
    assert not type_checker.is_source_destination(
        {"a_key": "stuff"}
    ), "Invalid key is invalid"
    assert not type_checker.is_source_destination(
        {
            "a_key": "stuff",
            "allow": True,
            "source": {"address": "123"},
            "destination": {"address": "321"},
        }
    ), "Extra key is invalid"
    assert not type_checker.is_source_destination(
        {
            "a_key": "stuff",
            "allow": True,
            "log": "abc",
            "source": {"address": "123"},
            "destination": {"address": "321"},
        }
    ), "Non bool log invalid"

    assert not type_checker.is_source_destination(
        {"allow": True, "source": {"address": 123, "port": "80"}}
    ), "Numeric address invalid"
    assert not type_checker.is_source_destination(
        {"allow": True, "rule": "abc", "source": {"address": "abc"}}
    ), "String rule is invalid"
    assert type_checker.is_source_destination(
        {"allow": True, "source": {"address": "abc"}}
    ), "Source address valid"
    assert type_checker.is_source_destination(
        {"allow": True, "source": {"port": "abc"}}
    ), "Source port valid"
    assert type_checker.is_source_destination(
        {"allow": True, "source": {"address": "123.123.123.123", "port": "80"}}
    ), "Source address and port is valid"
    assert type_checker.is_source_destination(
        {"allow": True, "destination": {"address": "abc"}}
    ), "Destination address valid"
    assert type_checker.is_source_destination(
        {"allow": True, "destination": {"port": "abc"}}
    ), "Destination port valid"
    assert type_checker.is_source_destination(
        {
            "allow": True,
            "rule": "20",
            "destination": {"address": "321.321.321.321", "port": "443"},
        }
    ), "Destination address and port is valid"
    assert type_checker.is_source_destination(
        {
            "allow": True,
            "rule": 20,
            "log": True,
            "protocol": "all",
            "description": "A connection",
            "source": {"address": "123.123.123.123", "port": "80"},
            "destination": {"address": "hosts", "port": "ports"},
        }
    ), "Source and destination address and port is valid"


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


def test_is_protocol():
    """
    .
    """
    assert not type_checker.is_protocol("abc"), "String not valid"
    assert not type_checker.is_protocol(80), "Number is not valid"
    assert not type_checker.is_protocol(["tcp"]), "Array is not valid"
    assert not type_checker.is_protocol({"tcp"}), "Dictionary not valid"
    assert type_checker.is_protocol("tcp"), "TCP valid"
    assert type_checker.is_protocol("udp"), "UDP valid"
    assert type_checker.is_protocol("tcp_udp"), "TCP/UDP valid"
    assert type_checker.is_protocol("ip"), "IP valid"


def test_is_state():
    """
    .
    """
    assert not type_checker.is_state("new"), "String not valid"
    assert not type_checker.is_state(80), "Number is not valid"
    assert not type_checker.is_state(["related"]), "Array is not valid"
    assert not type_checker.is_state(
        {"related": False}
    ), "Dictionary with bool not valid"
    assert not type_checker.is_state(
        {"related": "abcdef"}
    ), "Dictionary with string not valid"
    assert not type_checker.is_state(
        {"log": "enable"}
    ), "Dictionary with wrong key not valid"
    assert not type_checker.is_state(
        {"related": "enable", "log": "enable"}
    ), "Dictionary with extra key not valid"
    assert type_checker.is_state({"related": "enable"}), "Related valid"
    assert type_checker.is_state(
        {"related": "enable", "new": "disable"}
    ), "Mixed keys valid"
    assert type_checker.is_state(
        {
            "related": "enable",
            "established": "enable",
            "new": "disable",
            "invalid": "disable",
        }
    ), "All keys valid"


def test_is_nat_type():
    """
    .
    """
    assert not type_checker.is_nat_type("abc"), "String not valid"
    assert not type_checker.is_nat_type(80), "Number is not valid"
    assert not type_checker.is_nat_type(["in"]), "Array is not valid"
    assert not type_checker.is_nat_type({"source": True}), "Dictionary is not valid"

    assert type_checker.is_nat_type("source"), "Source is valid"
    assert type_checker.is_nat_type("destination"), "Destination is valid"
    assert type_checker.is_nat_type("masquerade"), "Masquerade is valid"
