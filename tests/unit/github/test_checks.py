"""
Tests the Github API functionality
"""
import time
from typing import Union

from ubiquiti_config_generator import root_parser
from ubiquiti_config_generator.github import checks, api, deploy_helper
from ubiquiti_config_generator.github.api import GREEN_CHECK, RED_CROSS
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


def test_get_output_validations(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(time, "time", lambda: 1610984224)
    output = checks.get_output_of_validations([])
    assert output == {
        "completed_at": "2021-01-18T15:37:04+00:00",
        "conclusion": "success",
        "output": {
            "title": "Configuration Validator",
            "summary": f"{GREEN_CHECK} Configuration successfully validated",
        },
    }, "Valid configuration output correct"

    output = checks.get_output_of_validations(["failure", "error"])
    assert output == {
        "completed_at": "2021-01-18T15:37:04+00:00",
        "conclusion": "failure",
        "output": {
            "title": "Configuration Validator",
            "summary": f"{RED_CROSS} Invalid configuration",
            "text": "- failure\n- error",
        },
    }, "Invalid validation output correct"


def test_get_pr_comment(monkeypatch):
    """
    .
    """
    # pylint: disable=unused-argument
    monkeypatch.setattr(
        api, "summarize_deploy_config_choices", lambda config: "deploy config"
    )
    all_category_differences = deploy_helper.diff_configurations(
        ["a 1", "b 2", "c 3", "d 5",], ["a 9", "b 2", "d 4", "e 6"],
    )
    monkeypatch.setattr(
        deploy_helper,
        "diff_configurations",
        lambda branch, prod: all_category_differences,
    )
    # This is mocked via the diff configurations, so can make this a no-op
    # pylint: disable=unused-argument
    monkeypatch.setattr(root_parser.RootNode, "get_commands", lambda self: ([], []))
    assert checks.get_pr_comment(
        {},
        root_parser.RootNode(None, [], None, [], None),
        root_parser.RootNode(None, [], None, [], None),
    ) == (
        "deploy config\n"
        "## Commands added:\n\n"
        "- c 3\n"
        "## Commands removed:\n\n"
        "- e 6\n"
        "## Commands changed:\n\n"
        "- a 1\n"
        "- d 5"
    ), "PR comment generated as expected for all categories"

    only_add_differences = deploy_helper.ConfigDifference()
    only_add_differences.add({"a": 5})
    # pylint: disable=unused-argument
    monkeypatch.setattr(
        deploy_helper, "diff_configurations", lambda branch, prod: only_add_differences
    )
    assert checks.get_pr_comment(
        {},
        root_parser.RootNode(None, [], None, [], None),
        root_parser.RootNode(None, [], None, [], None),
    ) == (
        "deploy config\n" "## Commands added:\n\n" "- a 5"
    ), "Only added returned for command summary"


def test_process_check_run(monkeypatch, capsys):
    """
    .
    """
    deploy_config = {
        "git": {"config-folder": "config", "diff-config-folder": "diff-config"}
    }

    form = {
        "action": "completed",
        "check_run": {"url": "/checks", "head_sha": "123abc", "pull_requests": []},
        "repository": {
            "full_name": "repository",
            "statuses_url": "github.com/user/repo/statuses{/sha}",
        },
    }

    assert checks.process_check_run(deploy_config, form, "abc123"), "No action succeeds"
    printed = capsys.readouterr()
    assert (
        printed.out == "Ignoring check_run action completed\n"
    ), "Ignoring completed output"

    form["action"] = "created"
    # pylint: disable=unused-argument
    monkeypatch.setattr(api, "update_check", lambda *args, **kwargs: False)
    assert not checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Update check fails causes process to fail"

    # pylint: disable=unused-argument
    @counter_wrapper
    def setup_repos(*args, **kwargs):
        """
        .
        """

    def fail_create(*args, **kwargs):
        """
        .
        """
        raise ValueError("Missing argument")

    @counter_wrapper
    def update_exception(*args, **kwargs):
        """
        .
        """

    monkeypatch.setattr(api, "update_check", lambda *args, **kwargs: True)
    monkeypatch.setattr(api, "setup_config_repo", setup_repos)
    monkeypatch.setattr(root_parser.RootNode, "create_from_configs", fail_create)
    monkeypatch.setattr(api, "update_check_with_exception", update_exception)

    assert not checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Fail to load configurations causes failure"
    assert setup_repos.counter == 1, "Set up repos called"
    assert update_exception.counter == 1, "Tried to update check with the exception"

    monkeypatch.setattr(
        root_parser.RootNode,
        "create_from_configs",
        lambda file_path: root_parser.RootNode(None, [], None, [], None),
    )

    def error_validate(self):
        """
        .
        """
        raise AttributeError("missing attribute")

    monkeypatch.setattr(root_parser.RootNode, "validate", error_validate)

    assert not checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Fail to validate fails"
    assert update_exception.counter == 2, "Tried to update check with exception"

    @counter_wrapper
    def get_validation_output(validation):
        """
        .
        """
        return {"conclusion": "whatever"}

    @counter_wrapper
    def update_check(*args, **kwargs):
        """
        Returns True first time, False second
        """
        return update_check.counter % 2

    monkeypatch.setattr(root_parser.RootNode, "validate", lambda self: True)
    monkeypatch.setattr(root_parser.RootNode, "validation_failures", lambda self: [])
    monkeypatch.setattr(checks, "get_output_of_validations", get_validation_output)
    monkeypatch.setattr(api, "set_commit_status", lambda *args, **kwargs: False)
    assert not checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Fail to set commit status causes failure"
    assert get_validation_output.counter == 1, "Got validation output"

    def check_commit_status_url(*args, **kwargs):
        """
        .
        """
        assert (
            kwargs["url"] == "github.com/user/repo/statuses/abc123"
        ), "Commit status URL correct"
        return True

    monkeypatch.setattr(
        api, "set_commit_status", lambda *args, **kwargs: check_commit_status_url
    )
    monkeypatch.setattr(api, "update_check", update_check)
    assert not checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Fail to set final check status causes failure"
    assert get_validation_output.counter == 2, "Got validation output"

    @counter_wrapper
    def get_pr_comment(*args, **kwargs):
        """
        .
        """
        return ""

    @counter_wrapper
    def add_comment(*args, **kwargs):
        """
        .
        """
        return add_comment.counter % 2 == 1

    monkeypatch.setattr(api, "update_check", lambda *args, **kwargs: True)
    monkeypatch.setattr(checks, "get_pr_comment", get_pr_comment)
    monkeypatch.setattr(api, "add_comment", add_comment)
    assert checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Updates successfully if no PRs"
    assert get_validation_output.counter == 3, "Got validation output"
    assert get_pr_comment.counter == 0, "PR comment not retrieved"
    assert add_comment.counter == 0, "No comments added"

    form["check_run"]["pull_requests"] = [{"url": "/"}, {"url": "/2"}]
    assert not checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Fails if a pull request comment fails"
    assert get_pr_comment.counter == 1, "PR comment retrieved once"
    assert (
        add_comment.counter == 2
    ), "Two comments added, even though the first one failed"

    monkeypatch.setattr(api, "add_comment", lambda *args, **kwargs: True)
    assert checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Succeeds if all PR comments succeed"
    assert get_pr_comment.counter == 2, "PR comment retrieved once"


def test_finalize_check_state(monkeypatch):
    """
    .
    """
    form = {
        "repository": {"statuses_url": "/statuses"},
        "check_run": {"head_sha": "abc123", "url": "/check/1"},
    }
    monkeypatch.setattr(
        checks,
        "get_output_of_validations",
        lambda *args, **kwargs: {"conclusion": "whatever"},
    )

    # pylint: disable=unused-argument
    @counter_wrapper
    def update_check(*args, **kwargs):
        """
        .
        """
        return True

    @counter_wrapper
    def finalize_commit(*args, **kwargs):
        """
        .
        """
        return True

    monkeypatch.setattr(api, "update_check", update_check)
    monkeypatch.setattr(checks, "finalize_commit_status", finalize_commit)
    assert checks.finalize_check_state([], form, "abc123"), "Check state works"
    assert update_check.counter == 1, "Update check called"
    assert finalize_commit.counter == 1, "Finalize commit called"

    monkeypatch.setattr(api, "update_check", lambda *args, **kwargs: True)
    monkeypatch.setattr(checks, "finalize_commit_status", lambda *args, **kwargs: False)
    assert not checks.finalize_check_state(
        [], form, "abc123"
    ), "Check state fails if commit status fails"

    monkeypatch.setattr(api, "update_check", lambda *args, **kwargs: False)
    monkeypatch.setattr(checks, "finalize_commit_status", lambda *args, **kwargs: True)
    assert not checks.finalize_check_state(
        [], form, "abc123"
    ), "Check state fails if check update fails"


def test_finalize_commit_status(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def make_request(url, *args, **kwargs):
        """
        .
        """
        assert url == "github.com/commits/ba21dc32"
        return True

    monkeypatch.setattr(api, "set_commit_status", make_request)
    assert checks.finalize_commit_status(
        "github.com/commits{/sha}", "ba21dc32", "abc123", "success"
    ), "Request value returned"
    assert make_request.counter == 1, "Request sent"
