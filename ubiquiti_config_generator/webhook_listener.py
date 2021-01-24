#!/usr/bin/env python3
"""
Contains the interactions for GitHub webhooks
"""
import secrets

from fastapi import FastAPI, Depends, Request, Response, HTTPException, status
from fastapi.responses import FileResponse, HTMLResponse
from fastapi.security import HTTPBasic, HTTPBasicCredentials
import uvicorn

from ubiquiti_config_generator import file_paths
from ubiquiti_config_generator.github import api, checks, push
from ubiquiti_config_generator.web import page

app = FastAPI(
    title="Ubiquiti Configuration Webhook Listener",
    description="Listens to GitHub webhooks to "
    "trigger actions on Ubiquiti configurations",
    version="1.0",
)


security = HTTPBasic()


def authenticate(credentials: HTTPBasicCredentials = Depends(security)):
    """
    Checks the current user for authentication
    """
    logging_config = file_paths.load_yaml_from_file("deploy.yaml")["logging"]
    correct_user = secrets.compare_digest(credentials.username, logging_config["user"])
    correct_pass = secrets.compare_digest(credentials.password, logging_config["pass"])

    if not (correct_user and correct_pass):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Incorrect username or password",
            headers={"WWW-Authenticate": "Basic"},
        )
    return credentials.username


@app.post("/")
async def on_webhook_action(request: Request) -> Response:
    """
    Runs for each webhook action
    """
    body = await request.body()
    form = await request.json()
    process_request(request.headers, body, form)


@app.get("/background.png")
async def background():
    """
    Returns the background image
    """
    return FileResponse("web/background.png")


@app.get("/main.css")
async def css():
    """
    Returns the CSS file
    """
    return FileResponse("web/main.css")


# pylint: disable=unused-argument
@app.get("/check/{rev}", response_class=HTMLResponse)
async def check_status(rev: str, username: str = Depends(authenticate)):
    """
    Returns the check status logs
    """
    return page.generate_page(
        {"type": "check", "status": "pending", "started": 1611501101, "revision1": rev}
    )


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
        push.check_push_for_deployment(deploy_config, form, access_token)
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
