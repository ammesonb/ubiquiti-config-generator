"""
Tests the Github API functionality
"""
from typing import Union

from ubiquiti_config_generator.github import checks, api


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


def test_handle_check_suite(monkeypatch, capsys):
    """
    .
    """
    checks.handle_check_suite({"action": "completed"}, "abc")
    printed = capsys.readouterr()
    assert (
        printed.out == "Ignoring check_suite action completed\n"
    ), "Completed is skipped"

    # pylint: disable=unused-argument
    monkeypatch.setattr(
        api,
        "send_github_request",
        lambda *args, **kwargs: Response({"message": "failed"}, 403),
    )

    checks.handle_check_suite(
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
        api,
        "send_github_request",
        lambda *args, **kwargs: Response({"message": "scheduled"}, 201),
    )

    checks.handle_check_suite(
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
