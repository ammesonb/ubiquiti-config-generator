"""
Tests the page creation code
"""
from ubiquiti_config_generator.web import page
from ubiquiti_config_generator.messages.log import Log


def test_deployment_page():
    """
    Tests the entire page creation, for deployments
    This is checks but with an extra thing, so should be sufficient
    Could mock out methods but this is actually a good integration test,
    so won't do that since DB dependencies are upstream anyways
    """

    html = page.generate_page(
        {
            "type": "deployment",
            "revision1": "abcdef",
            "revision2": "fedcba",
            "status": "success",
            "started": 1611516927.543210,
            "ended": 1611517099.987654,
            "logs": [
                Log("abcdef", "A message", 1611516928.132435, "fedcba", "created"),
                Log("abcdef", "Progress", 1611516935.978675, "fedcba", "log"),
                Log("abcdef", "Done", 1611517099.654321, "fedcba", "success"),
            ],
        }
    )

    assert html == "\n".join(
        [
            "<!DOCTYPE html>",
            "<html>",
            "  <head>",
            "    <title>",
            "      Configuration Validator",
            "    </title>",
            '    <link href="/main.css" rel="stylesheet"/>',
            '    <link href="https://fonts.googleapis.com/css?family=Lora:400,700|Tangerine:700" rel="stylesheet"/>',
            "  </head>",
            "  <body>",
            '    <div id="main">',
            "      <h1>",
            "        Deployment: ",
            '        <span style="color: 23d160">',
            "          success",
            "        </span>",
            "      </h1>",
            "      <div>",
            '        <h2 style="margin-bottom: 0">',
            "          Revision: abcdef..fedcba",
            "        </h2>",
            '        <h2 style="margin-top: 0.5%; float: left; vertical-align: middle; line-height: 50px">',
            "          Elapsed: 2 minutes, 52.44444394111633 seconds",
            "        </h2>",
            '        <h3 style="margin-top: 0.5%; float: right; vertical-align: middle; line-height: 50px">',
            "          Started at: 2021-01-24 19:35:27.543Z",
            "        </h3>",
            "      </div>",
            '      <hr style="clear: both; margin-bottom: 2%"/>',
            '      <table style="font-family: Courier New, monospace; width: 100%">',
            "        <colgroup>",
            '          <col style="width: 20%">',
            '          <col style="width: 10%">',
            '          <col style="width: 70%">',
            "        </colgroup>",
            "",
            '        <tr style="font-weight: bold; text-align: left;">',
            "          <th>",
            "            Timestamp",
            "          </th>",
            "          <th>",
            "            Level",
            "          </th>",
            "          <th>",
            "            Message",
            "          </th>",
            "        </tr>",
            "        <tr>",
            '          <td style="padding: 0.5%">',
            "            2021-01-24 19:35:28.132Z",
            "          </td>",
            '          <td style="padding: 0.5%">',
            '            [<span style="color: #ffff00">created</span>]',
            "          </td>",
            '          <td style="padding: 0.5%">',
            "            A message",
            "          </td>",
            "        </tr>",
            "        <tr>",
            '          <td style="padding: 0.5%">',
            "            2021-01-24 19:35:35.978Z",
            "          </td>",
            '          <td style="padding: 0.5%">',
            '            [<span style="color: silver">log</span>]',
            "          </td>",
            '          <td style="padding: 0.5%">',
            "            Progress",
            "          </td>",
            "        </tr>",
            "        <tr>",
            '          <td style="padding: 0.5%">',
            "            2021-01-24 19:38:19.654Z",
            "          </td>",
            '          <td style="padding: 0.5%">',
            '            [<span style="color: 23d160">success</span>]',
            "          </td>",
            '          <td style="padding: 0.5%">',
            "            Done",
            "          </td>",
            "        </tr>",
            "      </table>",
            "    </div>",
            "  </body>",
            "</html>",
        ]
    ), "HTML generated as expected"
