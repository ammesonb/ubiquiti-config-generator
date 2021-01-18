"""
Tests the webhook listener
"""
from fastapi import HTTPException
import pytest
import uvicorn

from ubiquiti_config_generator import webhook_listener, file_paths
from ubiquiti_config_generator.github import api, checks
from ubiquiti_config_generator.testing_utils import counter_wrapper


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

    monkeypatch.setattr(api, "validate_message", lambda *args, **kwargs: True)
    monkeypatch.setattr(checks, "handle_check_suite", handle_check_suite)
    monkeypatch.setattr(checks, "process_check_run", handle_check_run)
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
        {"x-hub-signature-256": "abc123", "x-github-event": "unknown"},
        "body",
        {"action": "reverted"},
    )
    printed = capsys.readouterr()
    assert printed.out == (
        "Got event unknown with action reverted\n"
        "Skipping event - no handler registered!\n"
    ), "Unknown event text printed"
    assert get_access_token.counter == 3, "Access token retrieved"
