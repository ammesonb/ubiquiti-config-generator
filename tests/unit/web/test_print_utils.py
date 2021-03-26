"""
Tests printing/visualization utilities
"""
from datetime import datetime, timezone

from ubiquiti_config_generator.web import print_utils
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_format_revision():
    """
    .
    """
    assert (
        print_utils.format_revision("rev1") == "rev1"
    ), "Single revision formats correctly"
    assert (
        print_utils.format_revision("rev1", "rev2") == "rev1..rev2"
    ), "Two revisions formats correctly"


def test_readable_time():
    """
    .
    """
    assert print_utils.readable_duration(0) == "0.00 seconds", "No seconds"
    assert print_utils.readable_duration(1) == "1.00 second", "One second"
    assert print_utils.readable_duration(60) == "1 minute, 0.00 seconds", "One minute"
    assert (
        print_utils.readable_duration(61) == "1 minute, 1.00 second"
    ), "One minute, one second"
    assert (
        print_utils.readable_duration(3600) == "1 hour, 0 minutes, 0.00 seconds"
    ), "One hour"
    assert (
        print_utils.readable_duration(3601) == "1 hour, 0 minutes, 1.00 second"
    ), "One hour, one second"
    assert (
        print_utils.readable_duration(86400)
        == "1 day, 0 hours, 0 minutes, 0.00 seconds"
    ), "One day"
    assert (
        print_utils.readable_duration(86401) == "1 day, 0 hours, 0 minutes, 1.00 second"
    ), "One day, one second"
    assert (
        print_utils.readable_duration(90061) == "1 day, 1 hour, 1 minute, 1.00 second"
    ), "One of everything"
    assert (
        print_utils.readable_duration(180122)
        == "2 days, 2 hours, 2 minutes, 2.00 seconds"
    ), "Two of everything"


def test_get_color():
    """
    .
    """
    assert (
        print_utils.get_color_for_status("success") == print_utils.SUCCESS
    ), "Success color right"
    assert (
        print_utils.get_color_for_status("pending") == print_utils.IN_PROGRESS
    ), "Pending color right"
    assert (
        print_utils.get_color_for_status("failure") == print_utils.FAILED
    ), "Failed color right"
    assert (
        print_utils.get_color_for_status("nonexistent") == print_utils.FAILED
    ), "Nonexistent color right"
    assert (
        print_utils.get_color_for_status("queued") == print_utils.WARNING
    ), "Queued color right"
    assert (
        print_utils.get_color_for_status("created") == print_utils.WARNING
    ), "Created color right"


def test_format_timestamp(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(
        print_utils,
        "to_utc",
        lambda datetime_object: datetime(2021, 1, 24, 14, 17, 20, 134520, timezone.utc),
    )
    assert (
        print_utils.format_timestamp(123456789) == "2021-01-24 14:17:20.134Z"
    ), "Timestamp formatted correctly"


def test_elapsed_duration(monkeypatch):
    """
    .
    """

    @counter_wrapper
    def duration(interval: float):
        """
        .
        """
        return interval

    monkeypatch.setattr(print_utils, "current_unix_timestamp", lambda: 123)
    monkeypatch.setattr(print_utils, "readable_duration", duration)
    assert (
        print_utils.elapsed_duration(120) == 3
    ), "Interval returned using current time"
    assert (
        print_utils.elapsed_duration(120.123, 122.123) == 2
    ), "Interval returned using specific end time"
