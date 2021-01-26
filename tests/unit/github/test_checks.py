"""
Tests the Github API functionality
"""
import time
from typing import Union

from ubiquiti_config_generator import root_parser
from ubiquiti_config_generator.github import checks, api, deploy_helper
from ubiquiti_config_generator.github.api import GREEN_CHECK, RED_CROSS
from ubiquiti_config_generator.messages import db
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

    @counter_wrapper
    def create_check(check, cursor=None):
        """
        .
        """
        assert check.status == "pending", "Check status correct"

    @counter_wrapper
    def update_check(log, cursor=None):
        """
        .
        """
        assert log.status == "failure", "Log should fail"
        assert log.message == "Failed to create a check run", "Log has expected message"

    @counter_wrapper
    def add_check_log(log, cursor=None):
        """
        .
        """
        assert log.status == "log", "Is a log"
        assert log.message == "Check run scheduled", "Log has expected message"

    @counter_wrapper
    def update_commit_status(
        url: str, sha: str, access_token: str, status: str, description: str
    ):
        """
        .
        """
        assert status == "failure", "Commit should have failed"

    monkeypatch.setattr(db, "create_check", create_check)
    monkeypatch.setattr(db, "update_check_status", update_check)
    monkeypatch.setattr(db, "add_check_log", add_check_log)
    monkeypatch.setattr(api, "set_commit_status", update_commit_status)
    checks.handle_check_suite(
        {
            "action": "requested",
            "check_suite": {"head_sha": "123abc", "app": {"external_url": "/app"}},
            "repository": {"url": "/", "statuses_url": "/statuses"},
        },
        "abc",
    )
    printed = capsys.readouterr()
    assert printed.out == (
        "Requesting a check\n"
        "Failed to schedule check: got status 403!\n"
        "{'message': 'failed'}\n"
    ), "Failed to schedule check printed"
    assert update_commit_status.counter == 1, "Commit status updated"
    assert create_check.counter == 1, "Check created"
    assert update_check.counter == 1, "Check status updated"

    # pylint: disable=unused-argument
    monkeypatch.setattr(
        api,
        "send_github_request",
        lambda *args, **kwargs: Response({"message": "scheduled"}, 201),
    )

    checks.handle_check_suite(
        {
            "action": "requested",
            "check_suite": {"head_sha": "123abc", "app": {"external_url": "/app"}},
            "repository": {"url": "/"},
        },
        "abc",
    )
    printed = capsys.readouterr()
    assert printed.out == (
        "Requesting a check\n" "Check requested successfully\n"
    ), "Success output printed"
    assert create_check.counter == 2, "Check created"
    assert add_check_log.counter == 1, "Added check log"


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
            "deployments_url": "deployments",
            "clone_url": "clone",
        },
    }

    # pylint: disable=unused-argument
    @counter_wrapper
    def check_update_success(log, cursor=None):
        """
        .
        """
        assert log.status == "success", "Check should succeed"

    @counter_wrapper
    def check_update_fails(log, cursor=None):
        """
        .
        """
        assert log.status == "failure", "Check status fails"

    @counter_wrapper
    def add_check_log(log, cursor=None):
        """
        .
        """

    monkeypatch.setattr(db, "update_check_status", check_update_success)
    monkeypatch.setattr(db, "add_check_log", add_check_log)

    assert checks.process_check_run(deploy_config, form, "abc123"), "No action succeeds"
    assert check_update_success.counter == 1, "Check status updated"
    printed = capsys.readouterr()
    assert (
        printed.out == "Ignoring check_run action completed\n"
    ), "Ignoring completed output"

    form["action"] = "created"
    monkeypatch.setattr(db, "update_check_status", check_update_fails)
    # pylint: disable=unused-argument
    monkeypatch.setattr(api, "update_check", lambda *args, **kwargs: False)
    assert not checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Update check fails causes process to fail"
    assert check_update_fails.counter == 1, "Check updated to fail"

    # pylint: disable=unused-argument
    def fail_create(*args, **kwargs):
        """
        .
        """
        raise ValueError("Missing argument")

    @counter_wrapper
    def get_deployment(*args, **kwargs):
        """
        .
        """
        raise ValueError("Bad deployment")

    @counter_wrapper
    def finalize_check(*args, **kwargs):
        """
        .
        """
        return False

    monkeypatch.setattr(api, "update_check", lambda *args, **kwargs: True)
    monkeypatch.setattr(api, "get_active_deployment_sha", get_deployment)
    monkeypatch.setattr(checks, "finalize_check_state", finalize_check)
    assert not checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Fail to load deployment causes failure"
    assert finalize_check.counter == 1, "Finalized checks"
    assert check_update_fails.counter == 2, "Check updated to fail"
    assert add_check_log.counter == 1, "One log added"

    @counter_wrapper
    def setup_repos(*args, **kwargs):
        """
        .
        """

    @counter_wrapper
    def update_exception(*args, **kwargs):
        """
        .
        """

    @counter_wrapper
    def finalize_commit_status(*args, **kwargs):
        """
        .
        """

    monkeypatch.setattr(api, "get_active_deployment_sha", lambda *args, **kwargs: True)
    monkeypatch.setattr(api, "setup_config_repo", setup_repos)
    monkeypatch.setattr(root_parser.RootNode, "create_from_configs", fail_create)
    monkeypatch.setattr(api, "update_check_with_exception", update_exception)
    monkeypatch.setattr(checks, "finalize_commit_status", finalize_commit_status)

    assert not checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Fail to load configurations causes failure"
    assert setup_repos.counter == 1, "Set up repos called"
    assert update_exception.counter == 1, "Tried to update check with the exception"
    assert finalize_commit_status.counter == 1, "Commit status finalized"
    assert check_update_fails.counter == 3, "Check updated to fail"
    assert add_check_log.counter == 4, "Three more logs added"

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
    assert finalize_commit_status.counter == 2, "Commit status finalized"
    assert check_update_fails.counter == 4, "Check updated to fail"
    assert add_check_log.counter == 8, "Four more logs added"

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
    assert finalize_check.counter == 2, "Check state finalized"
    assert check_update_fails.counter == 5, "Check updated to fail"
    assert add_check_log.counter == 13, "Five more logs added"

    monkeypatch.setattr(checks, "finalize_check_state", lambda *args, **kwargs: False)
    monkeypatch.setattr(api, "update_check", update_check)
    assert not checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Set check state causes failure"
    assert check_update_fails.counter == 6, "Check updated to fail"
    assert add_check_log.counter == 18, "Five more logs added, again"

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

    monkeypatch.setattr(checks, "finalize_check_state", lambda *args, **kwargs: True)
    monkeypatch.setattr(api, "update_check", lambda *args, **kwargs: True)
    monkeypatch.setattr(checks, "get_pr_comment", get_pr_comment)
    monkeypatch.setattr(api, "add_comment", add_comment)
    monkeypatch.setattr(db, "update_check_status", check_update_success)
    assert checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Updates successfully if no PRs"
    assert get_pr_comment.counter == 0, "PR comment not retrieved"
    assert add_comment.counter == 0, "No comments added"
    assert check_update_success.counter == 2, "Check update succeed called"
    assert add_check_log.counter == 24, "Six more logs added"

    form["check_run"]["pull_requests"] = [{"url": "/"}, {"url": "/2"}]
    assert not checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Fails if a pull request comment fails"
    assert get_pr_comment.counter == 1, "PR comment retrieved once"
    assert (
        add_comment.counter == 2
    ), "Two comments added, even though the first one failed"
    assert (
        check_update_success.counter == 3
    ), "Check update succeed called, even if PR update fails"
    assert add_check_log.counter == 30, "Six more logs added, again"

    monkeypatch.setattr(
        root_parser.RootNode, "validation_failures", lambda self: ["failure", "error"]
    )
    monkeypatch.setattr(db, "update_check_status", check_update_fails)

    monkeypatch.setattr(api, "add_comment", lambda *args, **kwargs: True)
    assert checks.process_check_run(
        deploy_config, form, "abc123"
    ), "Succeeds if all PR comments succeed"
    assert get_pr_comment.counter == 2, "PR comment retrieved once"
    assert (
        check_update_fails.counter == 7
    ), "Check update is failure if validations fail"
    assert (
        add_check_log.counter == 38
    ), "Add 6 standard logs, plus one for each of the two validation issues"


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
