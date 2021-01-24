"""
Contains various printing utilities for web stuff
"""
from datetime import datetime, timezone
from typing import Optional


def get_color_for_status(status: Optional[str]) -> str:
    """
    Returns HTML color code for a given status
    if None, consider it new, as in queued or created
    """
    color = "silver"
    if status == "pending":
        color = "#209cee"
    elif status in ["failure", "nonexistent"]:
        color = "#ff3860"
    elif status == "success":
        color = "#23d160"
    elif status in ["created", "requested", "queued"]:
        color = "#ffff00"

    return color


def format_timestamp(timestamp) -> str:
    """
    .
    """
    timestamp_object = datetime.fromtimestamp(timestamp).astimezone(timezone.utc)
    return timestamp_object.strftime("%Y-%m-%d %H:%M:%S.%f")[:-3] + "Z"


def elapsed_duration(start_timestamp: int, end_timestamp: int = None) -> str:
    """
    Given a start timestamp and end (defaults to now)
    get a human-readable amount of time between them
    """
    end_timestamp = end_timestamp or int(datetime.utcnow().strftime("%s"))
    return readable_duration(end_timestamp - start_timestamp)


def format_revision(revision1: str, revision2: Optional[str]):
    """
    Formats one or more revisions
    """
    return f"{revision1}..{revision2}" if revision2 else revision1


def readable_duration(seconds) -> str:
    """
    Given a number of seconds, return a human readable version
    of the interval
    """
    days = int(seconds / 86400)
    seconds -= days * 86400
    hours = int(seconds / 3600)
    seconds -= hours * 3600
    minutes = int(seconds / 60)
    seconds -= minutes * 60

    time_string = ""
    if days:
        time_string += "{0} day{1}, ".format(days, "s" if days != 1 else "")
    if hours or days:
        time_string += "{0} hour{1}, ".format(hours, "s" if hours != 1 else "")
    if minutes or hours or days:
        time_string += "{0} minute{1}, ".format(minutes, "s" if minutes != 1 else "")

    return time_string + "{0} second{1}".format(seconds, "s" if seconds != 1 else "")
