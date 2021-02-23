"""
Contains various printing utilities for web stuff
"""
from datetime import datetime, timezone
from typing import Optional

NORMAL = "silver"
IN_PROGRESS = "#209cee"
FAILED = "#ff3860"
SUCCESS = "#23d160"
WARNING = "#ffff00"


def get_color_for_status(status: Optional[str]) -> str:
    """
    Returns HTML color code for a given status
    if None, consider it new, as in queued or created
    """
    color = NORMAL
    if status in ["pending", "in_progress"]:
        color = IN_PROGRESS
    elif status in ["failure", "nonexistent"]:
        color = FAILED
    elif status == "success":
        color = SUCCESS
    elif status in ["created", "requested", "queued"]:
        color = WARNING

    return color


def to_utc(timestamp: datetime):
    """
    Converts a datetime object to UTC
    """
    return timestamp.astimezone(timezone.utc)


def current_unix_timestamp() -> float:
    """
    Gets the current epoch time, with microsecond precision in a float
    """
    return float(datetime.utcnow().strftime("%s.%f"))


def format_timestamp(timestamp: float) -> str:
    """
    .
    """
    timestamp_object = to_utc(datetime.fromtimestamp(timestamp))
    return timestamp_object.strftime("%Y-%m-%d %H:%M:%S.%f")[:-3] + "Z"


def elapsed_duration(start_timestamp: float, end_timestamp: float = None) -> str:
    """
    Given a start timestamp and end (defaults to now)
    get a human-readable amount of time between them
    """
    end_timestamp = end_timestamp or current_unix_timestamp()
    return readable_duration(end_timestamp - start_timestamp)


def format_revision(revision1: str, revision2: Optional[str] = None):
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
