"""
Generates the base HTML of the page
"""
from datetime import datetime
from typing import Optional

# This is how they want it done, and use dynamically added tags to global scope
# While this is neat....doesn't seem like the best way to do things
# pylint: disable=wildcard-import,unused-wildcard-import
from pyhtml import *


def get_color_for_status(status: Optional[str]) -> str:
    """
    Returns HTML color code for a given status
    if None, consider it new, as in queued or created
    """
    color = None
    if status == "pending":
        color = "#209cee"
    elif status == "failure":
        color = "#ff3860"
    elif status == "success":
        color = "#23d160"
    elif status in ["created", "requested", "queued", None]:
        color = "#e0e317"

    return color


def format_timestamp(timestamp) -> str:
    """
    .
    """
    timestamp_object = datetime.fromtimestamp(timestamp)
    return timestamp_object.strftime("%Y-%m-%d %H:%M:%SZ")


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


def generate_page(context: dict) -> str:
    """
    Creates the HTML for the page, based off the context provided
    """
    # Since it turns out, all of these are defined using the magic dynamic definitions
    # as mentioned above for pyhtml
    # pylint: disable=undefined-variable
    page = html(
        head(
            title("Configuration Validator"),
            link(href="/main.css", rel="stylesheet"),
            link(
                href=(
                    "https://fonts.googleapis.com/css?"
                    "family=Lora:400,700|Tangerine:700"
                ),
                rel="stylesheet",
            ),
        ),
        body(
            div(id="main")(
                h1(
                    f"{context['type']}: ".title(),
                    span(style=f"color: {get_color_for_status(context['status'])}")(
                        context["status"]
                    ),
                ),
                div(
                    h2(style="margin-bottom: 0")(
                        "Revision: "
                        + format_revision(
                            context["revision1"], context.get("revision2", None)
                        )
                    ),
                    h2(
                        style=(
                            "margin-top: 0.5%; "
                            "float: left; "
                            "vertical-align: middle; "
                            "line-height: 50px"
                        )
                    )(
                        "Elapsed: "
                        + elapsed_duration(
                            context["started"], context.get("ended", None)
                        )
                    ),
                    h3(
                        style=(
                            "margin-top: 0.5%; "
                            "float: right; "
                            "vertical-align: middle; "
                            "line-height: 50px"
                        )
                    )(f"Started at: {format_timestamp(context['started'])}"),
                ),
                hr(style="clear: both"),
                ul([log.html() for log in context.get("logs", [])]),
            ),
        ),
    )

    return page.render()
