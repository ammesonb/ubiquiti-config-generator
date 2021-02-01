"""
Tests the webhook listener
"""
from fastapi import HTTPException
import pytest
import uvicorn

from ubiquiti_config_generator import webhook_listener, file_paths
from ubiquiti_config_generator.github import api, checks, push, deployment
from ubiquiti_config_generator.messages import db
from ubiquiti_config_generator.messages.check import Check
from ubiquiti_config_generator.messages.deployment import Deployment
from ubiquiti_config_generator.web import page
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_authenticate(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(
        file_paths,
        "load_yaml_from_file",
        lambda path: {"logging": {"user": "username", "pass": "password"}},
    )

    # pylint: disable=too-few-public-methods
    class Credentials:
        """
        Mocks credentials
        """

        username = "username"
        password = "passwd"

    credentials = Credentials()

    with pytest.raises(HTTPException):
        webhook_listener.authenticate(credentials)

    credentials.password = "password"
    assert (
        webhook_listener.authenticate(credentials) == "username"
    ), "User authenticates"


def test_run_listener(monkeypatch):
    """
    .
    """
    # pylint: disable=unused-argument
    @counter_wrapper
    def load_yaml(file_path: str):
        """
        .
        """
        assert "deploy.yaml" in file_path, "File loaded is deploy.yaml"
        return {"git": {"webhook-port": 443}}

    @counter_wrapper
    def run(*args, **kwargs):
        """
        .
        """

    monkeypatch.setattr(file_paths, "load_yaml_from_file", load_yaml)
    monkeypatch.setattr(uvicorn, "run", run)

    webhook_listener.run_listener()

    assert load_yaml.counter == 1, "Deploy yaml loaded"
    assert run.counter == 1, "Run listener called"


def test_process_request(monkeypatch, capsys):
    """
    .
    """
    # pylint: disable=unused-argument
    @counter_wrapper
    def get_access_token(*args, **kwargs):
        """
        .
        """
        return "abc123"

    @counter_wrapper
    def load_yaml(file_path: str):
        """
        .
        """
        assert "deploy.yaml" in file_path, "File loaded is deploy.yaml"
        return {"git": {"webhook-port": 443}}

    monkeypatch.setattr(api, "get_jwt", lambda *args, **kwargs: "abc123")
    monkeypatch.setattr(api, "get_access_token", get_access_token)

    monkeypatch.setattr(file_paths, "load_yaml_from_file", load_yaml)
    monkeypatch.setattr(api, "validate_message", lambda *args, **kwargs: False)

    with pytest.raises(HTTPException):
        webhook_listener.process_request({"x-hub-signature-256": "abc123"}, "body", {})

    printed = capsys.readouterr()
    assert printed.out == "Unauthorized request!\n", "Unauthorized printed"

    @counter_wrapper
    def handle_check_suite(*args, **kwargs):
        """
        .
        """

    @counter_wrapper
    def handle_check_run(*args, **kwargs):
        """
        .
        """

    @counter_wrapper
    def handle_push(*args, **kwargs):
        """
        .
        """

    @counter_wrapper
    def handle_deployment(*args, **kwargs):
        """
        .
        """

    monkeypatch.setattr(api, "validate_message", lambda *args, **kwargs: True)
    monkeypatch.setattr(checks, "handle_check_suite", handle_check_suite)
    monkeypatch.setattr(checks, "process_check_run", handle_check_run)
    monkeypatch.setattr(push, "check_push_for_deployment", handle_push)
    monkeypatch.setattr(deployment, "handle_deployment", handle_deployment)
    webhook_listener.process_request(
        {"x-hub-signature-256": "abc123", "x-github-event": "check_suite"},
        "body",
        {"action": "requested"},
    )
    printed = capsys.readouterr()
    assert (
        printed.out == "Got event check_suite with action requested\n"
    ), "Check suite text printed"
    assert get_access_token.counter == 1, "Access token retrieved"
    assert handle_check_suite.counter == 1, "Check suite process called"

    webhook_listener.process_request(
        {"x-hub-signature-256": "abc123", "x-github-event": "check_run"},
        "body",
        {"action": "created"},
    )
    printed = capsys.readouterr()
    assert (
        printed.out == "Got event check_run with action created\n"
    ), "Check run text printed"
    assert get_access_token.counter == 2, "Access token retrieved"
    assert handle_check_run.counter == 1, "Check run process called"

    webhook_listener.process_request(
        {"x-hub-signature-256": "abc123", "x-github-event": "push"},
        "body",
        {"action": ""},
    )
    printed = capsys.readouterr()
    assert printed.out == "Got event push with action \n", "Push text printed"
    assert handle_push.counter == 1, "Push process called"

    webhook_listener.process_request(
        {"x-hub-signature-256": "abc123", "x-github-event": "deployment"},
        "body",
        {"action": "created"},
    )
    printed = capsys.readouterr()
    assert (
        printed.out == "Got event deployment with action created\n"
    ), "Deployment text printed"
    assert handle_deployment.counter == 1, "Deployment process called"

    webhook_listener.process_request(
        {"x-hub-signature-256": "abc123", "x-github-event": "unknown"},
        "body",
        {"action": "reverted"},
    )
    printed = capsys.readouterr()
    assert printed.out == (
        "Got event unknown with action reverted\n"
        "Skipping event - no handler registered!\n"
    ), "Unknown event text printed"


def test_render_check(monkeypatch):
    """
    .
    """

    def get_check(revision: str):
        return Check(revision, "success", 123, 234, ["log"])

    def fake_render(context: dict):
        """
        .
        """
        assert context == {
            "type": "check",
            "revision1": "abc",
            "status": "success",
            "started": 123,
            "ended": 234,
            "logs": ["log"],
        }, "Correct context passed"

        return "check page"

    monkeypatch.setattr(db, "get_check", get_check)
    monkeypatch.setattr(page, "generate_page", fake_render)

    assert webhook_listener.render_check("abc") == "check page", "Correct return value"


def test_render_deployment(monkeypatch):
    """
    .
    """

    def get_deployment(revision1: str, revision2: str):
        return Deployment(revision1, revision2, "failure", 321, 432, ["log"])

    def fake_render(context: dict):
        """
        .
        """
        assert context == {
            "type": "deployment",
            "revision1": "abc",
            "revision2": "cba",
            "status": "failure",
            "started": 321,
            "ended": 432,
            "logs": ["log"],
        }, "Correct context passed"

        return "deployment page"

    monkeypatch.setattr(db, "get_deployment", get_deployment)
    monkeypatch.setattr(page, "generate_page", fake_render)

    assert (
        webhook_listener.render_deployment("abc", "cba") == "deployment page"
    ), "Correct return value"
