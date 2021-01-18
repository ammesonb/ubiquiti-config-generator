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

from ubiquiti_config_generator.github import api
from ubiquiti_config_generator.github.api import GREEN_CHECK, WARNING
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
        api.get_jwt(
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

    assert api.get_access_token("a-jwt") == "v1.a-token", "Access token retrieved"
    assert installations.counter == 1, "Installations retrieved"
    assert token.counter == 1, "Token retrieved"


def test_validate_message(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(hmac.HMAC, "hexdigest", lambda self: "abc123")
    assert api.validate_message(
        {"git": {"webhook-secret": "secret"}}, "content".encode(), "sha256=abc123"
    ), "Message validates"


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
                "summary": api.RED_CROSS + " an exception",
                "text": "foo\n\nexception traceback",
                "title": "Configuration Validator",
            },
        }, "Expected extra data provided"

    monkeypatch.setattr(time, "time", lambda: 1610749859)
    monkeypatch.setattr(api, "update_check", update_check)
    monkeypatch.setattr(traceback, "format_exc", lambda: "exception traceback")
    api.update_check_with_exception("an exception", "abc123", "/url", ValueError("foo"))
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
    assert not api.update_check(
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
    assert api.update_check(
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
    assert not api.add_comment(
        "abc123", "/pull", "a comment"
    ), "Adding comment fails if PR not pulled"
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
    assert not api.add_comment("abc123", "/pull", "a comment"), "Posting comment fails"
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
    assert api.add_comment("abc123", "/pull", "a comment"), "Posting comment works"
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

    api.clone_repository("abc123", "github.com/user/repository.git", "a_folder", False)
    assert remove_tree.counter == 0, "No folder to delete"
    printed = capsys.readouterr()
    assert printed.out == "Cloning repository.git into a_folder\n", "Output printed"

    api.clone_repository("abc123", "github.com/user/repository.git", "a_folder", True)
    assert remove_tree.counter == 1, "Path deleted"

    monkeypatch.setattr(path, "exists", lambda path: False)
    api.clone_repository("abc123", "github.com/user/repository.git", "a_folder", True)
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

    api.checkout("a-repo", "abc123")
    assert chdir.counter == 2, "Directory changed, and changed back"
    assert run.counter == 1, "Checkout called"

    printed = capsys.readouterr()
    assert printed.out == "Checking out abc123 in a-repo\n", "Output printed"


def test_summarize_deploy_config():
    """
    .
    """
    summary = api.summarize_deploy_config_choices(
        {
            "apply-difference-only": True,
            "auto-rollback-on-failure": True,
            "reboot-after-minutes": 5,
            "save-after-commit": False,
        }
    )
    assert summary == "\n".join(
        [
            "## Deployment overview",
            f"- {GREEN_CHECK} Applying *DIFFERENCE* only",
            f"- {GREEN_CHECK} Will rollback on configuration error",
            f"- {GREEN_CHECK} Will restart after 5 minutes without confirm",
            f"- {GREEN_CHECK} Will **NOT** save automatically",
        ]
    ), "Safe summary correct"

    summary = api.summarize_deploy_config_choices(
        {
            "apply-difference-only": False,
            "auto-rollback-on-failure": False,
            "reboot-after-minutes": 0,
            "save-after-commit": True,
        }
    )
    assert summary == "\n".join(
        [
            "## Deployment overview",
            f"- {GREEN_CHECK} Applying *ENTIRE* configuration",
            f"- {WARNING} Will **NOT** rollback on configuration error",
            f"- {WARNING} Will **NOT** automatically restart",
            f"- {WARNING} **WILL** save configuration immediately after commit",
        ]
    ), "Less-safe summary correct"


def test_setup_config_repo(monkeypatch):
    """
    .
    """
    # pylint: disable=unused-argument
    @counter_wrapper
    def clone_repo(*args, **kwargs):
        """
        .
        """

    # pylint: disable=unused-argument
    @counter_wrapper
    def checkout(*args, **kwargs):
        """
        .
        """

    monkeypatch.setattr(api, "clone_repository", clone_repo)
    monkeypatch.setattr(api, "checkout", checkout)

    api.setup_config_repo(
        "abc123",
        "/repo",
        {"git": {"config-folder": "config", "diff-config-folder": "diff"}},
    )

    assert clone_repo.counter == 1, "Only main config cloned"
    assert checkout.counter == 0, "Checkout not called"

    api.setup_config_repo(
        "abc123",
        "/repo",
        {"git": {"config-folder": "config", "diff-config-folder": "diff"}},
        "sha",
    )

    assert clone_repo.counter == 3, "Both configs cloned"
    assert checkout.counter == 1, "Checkout called"


def test_send_github_request(monkeypatch):
    """
    .
    """
    # pylint: disable=unused-argument
    @counter_wrapper
    def get(*args, **kwargs):
        """
        .
        """

    @counter_wrapper
    def post(*args, **kwargs):
        """
        .
        """

    monkeypatch.setattr(requests, "get", get)
    monkeypatch.setattr(requests, "post", post)

    api.send_github_request("/url", "get", "abc123")
    assert get.counter == 1, "Get called"
    assert post.counter == 0, "Post not called"

    api.send_github_request("/url", "post", "abc123", {"data": "stuff"})
    assert get.counter == 1, "Get called"
    assert post.counter == 1, "Post called"
