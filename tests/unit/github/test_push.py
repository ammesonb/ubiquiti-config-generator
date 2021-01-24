"""
Test push Git code
"""

from ubiquiti_config_generator.github import push, api
from ubiquiti_config_generator.testing_utils import counter_wrapper
from tests.unit.github.test_api import Response


def test_is_against_primary_branch():
    """
    .
    """
    config = {"git": {"primary-branch": "main"}}
    assert not push.is_against_primary_branch(
        config, "refs/head/feature"
    ), "Feature is not primary"
    assert not push.is_against_primary_branch(
        config, "refs/head/main-test"
    ), "Test from main is not primary"
    assert push.is_against_primary_branch(config, "refs/head/main"), "Main is primary"


def test_check_for_deployment(monkeypatch, capsys):
    """
    .
    """
    # pylint: disable=unused-argument
    @counter_wrapper
    def set_commit_status(url: str, ref: str, access_token: str, status: str):
        """
        .
        """
        assert status == "pending", "Commit status is pending"

    monkeypatch.setattr(api, "set_commit_status", set_commit_status)
    monkeypatch.setattr(
        push, "is_against_primary_branch", lambda *args, **kwargs: False
    )

    form = {
        "repository": {"statuses_url": "/statuses", "deployments_url": "/deployments"},
        "before": "bef",
        "after": "af",
        "ref": "123abc",
    }

    assert (
        push.check_push_for_deployment({}, form, "abc123") is None
    ), "No deployments against non-primary branch"
    assert set_commit_status.counter == 1, "Commit status set, against all branches"

    monkeypatch.setattr(push, "is_against_primary_branch", lambda *args, **kwargs: True)

    @counter_wrapper
    def send_github_request(url: str, method: str, access_token: str, json: dict):
        """
        .
        """
        assert json["payload"]["previous_commit"] == (
            "sha" if send_github_request.counter == 1 else "bef"
        ), "Previous commit correct"
        return (
            Response({"message": "failed"}, 403)
            if send_github_request.counter == 1
            else Response({"message": "created"}, 201)
        )

    monkeypatch.setattr(api, "send_github_request", send_github_request)
    monkeypatch.setattr(api, "get_active_deployment_sha", lambda *args, **kwargs: "sha")

    assert not push.check_push_for_deployment(
        {}, form, "abc123"
    ), "Create deployment fails"
    printed = capsys.readouterr()
    assert set_commit_status.counter == 2, "Commit status set"
    assert send_github_request.counter == 1, "Attempted to create deployment"
    assert printed.out == (
        "Failed to create deployment\n" "{'message': 'failed'}\n"
    ), "Fail deploy printed"

    monkeypatch.setattr(api, "get_active_deployment_sha", lambda *args, **kwargs: None)
    assert push.check_push_for_deployment(
        {}, form, "abc123"
    ), "Create deployment succeeds"
    printed = capsys.readouterr()
    assert set_commit_status.counter == 3, "Commit status set"
    assert send_github_request.counter == 2, "Deployment created"
    assert printed.out == "", "Nothing printed"