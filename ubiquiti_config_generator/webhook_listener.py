#!/usr/bin/env python3
"""
Contains the interactions for GitHub webhooks
"""

from fastapi import FastAPI, Request, Response, HTTPException
import uvicorn

from ubiquiti_config_generator import file_paths
from ubiquiti_config_generator.github import api, checks, push

app = FastAPI(
    title="Ubiquiti Configuration Webhook Listener",
    description="Listens to GitHub webhooks to "
    "trigger actions on Ubiquiti configurations",
    version="1.0",
)


@app.post("/")
async def on_webhook_action(request: Request) -> Response:
    """
    Runs for each webhook action
    """
    body = await request.body()
    form = await request.json()
    process_request(request.headers, body, form)


def process_request(headers: dict, body: str, form: dict) -> Response:
    """
    Perform the actual processing of a request
    """
    deploy_config = file_paths.load_yaml_from_file("deploy.yaml")

    if not api.validate_message(deploy_config, body, headers["x-hub-signature-256"]):
        print("Unauthorized request!")
        raise HTTPException(status_code=404, detail="Invalid body hash")

    access_token = api.get_access_token(api.get_jwt(deploy_config))

    print(
        "Got event {0} with action {1}".format(
            headers["x-github-event"], form.get("action", "")
        )
    )

    if headers["x-github-event"] == "check_suite":
        checks.handle_check_suite(form, access_token)
    elif headers["x-github-event"] == "check_run":
        checks.process_check_run(deploy_config, form, access_token)
    elif headers["x-github-event"] == "push":
        print(form)
    else:
        print("Skipping event - no handler registered!")


def run_listener():
    """
    Runs the listener
    """
    deploy_config = file_paths.load_yaml_from_file("deploy.yaml")
    uvicorn.run(
        "webhook_listener:app", port=deploy_config["git"]["webhook-port"], reload=True
    )


if __name__ == "__main__":
    run_listener()
