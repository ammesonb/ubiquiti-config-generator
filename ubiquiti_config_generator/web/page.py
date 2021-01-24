"""
Generates the base HTML of the page
"""

# This is how they want it done, and use dynamically added tags to global scope
# While this is neat....doesn't seem like the best way to do things
# pylint: disable=wildcard-import,unused-wildcard-import
from pyhtml import *
from ubiquiti_config_generator.web import print_utils


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
                    span(
                        style="color: "
                        + print_utils.get_color_for_status(context["status"])
                    )(context["status"]),
                ),
                div(
                    h2(style="margin-bottom: 0")(
                        "Revision: "
                        + print_utils.format_revision(
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
                        + print_utils.elapsed_duration(
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
                    )(
                        "Started at: "
                        + print_utils.format_timestamp(context["started"])
                    ),
                ),
                hr(style="clear: both"),
                ul([log.html() for log in context.get("logs", [])]),
            ),
        ),
    )

    return page.render()
