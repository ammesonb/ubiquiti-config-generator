"""
Tests the Github API functionality
"""
from datetime import datetime
import hashlib
import hmac
import requests
import time
import traceback
from typing import Union

import jwt
from ubiquiti_config_generator import github_api
from ubiquiti_config_generator.testing_utils import counter_wrapper


class Response:
    def __init__(self, json_result: Union[list, dict], status_code: int = 200):
        self.data = json_result
        self.status_code = status_code

    def json(self) -> Union[list, dict]:
        return self.data


def test_get_jwt(monkeypatch):
    """
    .
    """
    jwt_token = (
        "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpYXQiOjEwMDAsImV4cCI6MTYwMCwiaXNzIjoxf"
        "Q.tzg47AVUmkqBJuhK27Ho-1shIQ-iBbr5Z54NKDAa6i5SsG5bu9qy1NC3kfp9875yJWh1gAgInBMm"
        "C6QxnVhtWNWcBe9zXr9U6DXQD0PLC0mBxeaKynP71LuPKG21DGlzKxAvxSx9Q3YXUhOq_ydijPVXnO"
        "484SnrGQpWS6hwsDk"
    )
    assert (
        jwt.encode(
            {"iat": 1000, "exp": 1600, "iss": 1},
            open("tests/unit/resources/test-key.pem", "rb").read(),
            "RS256",
        )
        == jwt_token
    ), "Cached token correct"

    monkeypatch.setattr(time, "time", lambda: 1000.12345)
    assert (
        github_api.get_jwt(
            {
                "git": {
                    "app-id": 1,
                    "private-key-path": "tests/unit/resources/test-key.pem",
                }
            }
        )
        == jwt_token
    ), "JWT generated correctly"


def test_get_access_token(monkeypatch):
    """
    .
    """

    @counter_wrapper
    def installations(*args, **kwargs):
        return Response([{"access_tokens_url": "/tokens"}])

    @counter_wrapper
    def token(*args, **kwargs):
        return Response({"token": "v1.a-token"})

    monkeypatch.setattr(requests, "get", installations)
    monkeypatch.setattr(requests, "post", token)

    assert (
        github_api.get_access_token("a-jwt") == "v1.a-token"
    ), "Access token retrieved"
    assert installations.counter == 1, "Installations retrieved"
    assert token.counter == 1, "Token retrieved"


def test_validate_message(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(hmac.HMAC, "hexdigest", lambda self: "abc123")
    assert github_api.validate_message(
        {"git": {"webhook-secret": "secret"}}, "content".encode(), "sha256=abc123"
    ), "Message validates"


def test_handle_check_suite(monkeypatch, capsys):
    """
    .
    """
    github_api.handle_check_suite({"action": "completed"}, "abc")
    printed = capsys.readouterr()
    assert (
        printed.out == "Ignoring check_suite action completed\n"
    ), "Completed is skipped"

    monkeypatch.setattr(
        requests, "post", lambda *args, **kwargs: Response({"message": "failed"}, 403)
    )

    github_api.handle_check_suite(
        {
            "action": "requested",
            "check_suite": {"head_sha": "123abc"},
            "repository": {"url": "/"},
        },
        "abc",
    )
    printed = capsys.readouterr()
    assert printed.out == (
        "Requesting a check\n"
        "Failed to schedule check: got status 403!\n"
        "{'message': 'failed'}\n"
    ), "Failed to schedule check printed"

    monkeypatch.setattr(
        requests,
        "post",
        lambda *args, **kwargs: Response({"message": "scheduled"}, 201),
    )

    github_api.handle_check_suite(
        {
            "action": "requested",
            "check_suite": {"head_sha": "123abc"},
            "repository": {"url": "/"},
        },
        "abc",
    )
    printed = capsys.readouterr()
    assert printed.out == (
        "Requesting a check\n" "Check requested successfully\n"
    ), "Success output printed"


def test_update_check_with_exception(monkeypatch, capsys):
    """
    .
    """

    @counter_wrapper
    def update_check(access_token, check_url, status, extra_data):
        assert extra_data == {
            "completed_at": "2021-01-15T22:30:59+00:00",
            "conclusion": "failure",
            "output": {
                "summary": github_api.RED_CROSS + " an exception",
                "text": "foo\n\nexception traceback",
                "title": "Configuration Validator",
            },
        }, "Expected extra data provided"

    monkeypatch.setattr(time, "time", lambda: 1610749859)
    monkeypatch.setattr(github_api, "update_check", update_check)
    monkeypatch.setattr(traceback, "format_exc", lambda: "exception traceback")
    github_api.update_check_with_exception(
        "an exception", "abc123", "/url", ValueError("foo")
    )
    printed = capsys.readouterr()
    assert printed.out == "Caught exception foo during check!\n", "Error printed"
    assert update_check.counter == 1, "Update check called"


def test_update_check(monkeypatch, capsys):
    """
    .
    """
    monkeypatch.setattr(
        requests, "patch", lambda *args, **kwargs: Response({"message": "failed"}, 403)
    )
    assert not github_api.update_check(
        "abc123", "checks/312", "requested", {"data": "value"}
    ), "Update check fails"
    printed = capsys.readouterr()
    assert printed.out == (
        "Updating check 312 to requested\n"
        "Check update 312 failed with code 403\n"
        "{'message': 'failed'}\n"
    ), "Expected failures printed"

    monkeypatch.setattr(
        requests, "patch", lambda *args, **kwargs: Response({"message": "updated"}, 200)
    )
    assert github_api.update_check(
        "abc123", "checks/312", "requested", {"data": "value"}
    ), "Update check succeeds"
    printed = capsys.readouterr()
    assert printed.out == (
        "Updating check 312 to requested\n"
    ), "Success has no extra prints"
