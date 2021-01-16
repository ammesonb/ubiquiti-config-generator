"""
Tests the Github API functionality
"""
import hmac
import os
from os import path
import shutil
import subprocess
import time
import traceback
from typing import Union

import jwt
import requests
from ubiquiti_config_generator import github_api
from ubiquiti_config_generator.testing_utils import counter_wrapper


# pylint: disable=too-few-public-methods
class Response:
    """
    Fakes endpoint response
    """

    def __init__(self, json_result: Union[list, dict], status_code: int = 200):
        self.data = json_result
        self.status_code = status_code

    def json(self) -> Union[list, dict]:
        """
        JSON
        """
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

    # pylint: disable=unused-argument
    @counter_wrapper
    def installations(*args, **kwargs):
        return Response([{"access_tokens_url": "/tokens"}])

    # pylint: disable=unused-argument
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

    # pylint: disable=unused-argument
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

    # pylint: disable=unused-argument
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

    # pylint: disable=unused-argument
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
    # pylint: disable=unused-argument
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

    # pylint: disable=unused-argument
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


def test_add_comment(monkeypatch, capsys):
    """
    .
    """
    # pylint: disable=unused-argument
    monkeypatch.setattr(
        requests, "get", lambda *args, **kwargs: Response({"message": "bad"}, 403)
    )
    github_api.add_comment("abc123", "/pull", "a comment")
    printed = capsys.readouterr()
    assert printed.out == (
        "Failed to get pull request /pull\n" "{'message': 'bad'}\n"
    ), "Failure to get pull request prints"

    # pylint: disable=unused-argument
    monkeypatch.setattr(
        requests,
        "get",
        lambda *args, **kwargs: Response({"comments_url": "comments"}, 200),
    )
    # pylint: disable=unused-argument
    monkeypatch.setattr(
        requests, "post", lambda *args, **kwargs: Response({"message": "failure"}, 403)
    )
    github_api.add_comment("abc123", "/pull", "a comment")
    printed = capsys.readouterr()
    assert printed.out == (
        "Posting comment\n"
        "Failed to post comment to review /pull\n"
        "{'message': 'failure'}\n"
    ), "Failure to post comment to pull request prints"

    # pylint: disable=unused-argument
    monkeypatch.setattr(
        requests, "post", lambda *args, **kwargs: Response({"message": "posted"}, 201)
    )
    github_api.add_comment("abc123", "/pull", "a comment")
    printed = capsys.readouterr()
    assert printed.out == ("Posting comment\n"), "Posting comment only text printed"


def test_clone_repo(monkeypatch, capsys):
    """
    .
    """
    monkeypatch.setattr(path, "exists", lambda path: True)

    # pylint: disable=unused-argument
    @counter_wrapper
    def remove_tree(*args, **kwargs):
        pass

    # pylint: disable=unused-argument
    @counter_wrapper
    def run(*args, **kwargs):
        pass

    monkeypatch.setattr(shutil, "rmtree", remove_tree)
    monkeypatch.setattr(subprocess, "run", run)

    github_api.clone_repository(
        "abc123", "github.com/user/repository.git", "a_folder", False
    )
    assert remove_tree.counter == 0, "No folder to delete"
    printed = capsys.readouterr()
    assert printed.out == "Cloning repository.git into a_folder\n", "Output printed"

    github_api.clone_repository(
        "abc123", "github.com/user/repository.git", "a_folder", True
    )
    assert remove_tree.counter == 1, "Path deleted"

    monkeypatch.setattr(path, "exists", lambda path: False)
    github_api.clone_repository(
        "abc123", "github.com/user/repository.git", "a_folder", True
    )
    assert remove_tree.counter == 1, "Path not deleted if doesn't exist"

    assert run.counter == 3, "Run called three times"


def test_checkout(monkeypatch, capsys):
    """
    .
    """

    def getcwd():
        return "somewhere"

    @counter_wrapper
    def chdir(directory: str):
        """
        .
        """
        if chdir.counter == 2:
            assert directory == "somewhere", "Returned to start"

    # pylint: disable=unused-argument
    @counter_wrapper
    def run(*args, **kwargs):
        """
        .
        """

    monkeypatch.setattr(os, "getcwd", getcwd)
    monkeypatch.setattr(os, "chdir", chdir)
    monkeypatch.setattr(subprocess, "run", run)

    github_api.checkout("a-repo", "abc123")
    assert chdir.counter == 2, "Directory changed, and changed back"
    assert run.counter == 1, "Checkout called"

    printed = capsys.readouterr()
    assert printed.out == "Checking out abc123 in a-repo\n", "Output printed"
