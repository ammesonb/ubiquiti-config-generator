#!/usr/bin/env python3
"""
Contains the interactions for GitHub webhooks
"""
import hashlib
import hmac
import time

from fastapi import FastAPI, Request, Response, HTTPException
import jwt
import requests
import uvicorn

from ubiquiti_config_generator import file_paths

app = FastAPI(
    title="Ubiquiti Configuration Webhook Listener",
    description="Listens to GitHub webhooks to "
    "trigger actions on Ubiquiti configurations",
    version="1.0",
)


def get_jwt(deploy_config: dict) -> str:
    """
    Get the JSON web token, needed for basic API access
    """
    return jwt.encode(
        {
            # Request starting now
            "iat": int(time.time()),
            # Token expires in ten minutes
            "exp": int(time.time()) + 600,
            "iss": deploy_config["git"]["app-id"],
        },
        open(deploy_config["git"]["private-key-path"], "rb").read(),
        "RS256",
    )


def get_access_token(jwt_token: str) -> str:
    """
    Gets an access token, used for fully-authenticated app actions
    """
    headers = {
        "accept": "application/vnd.github.v3+json",
        "Authorization": "Bearer " + jwt_token,
    }
    installations = requests.get(
        "https://api.github.com/app/installations", headers=headers
    )

    response = requests.post(
        installations.json()[0]["access_tokens_url"], headers=headers
    )
    return response.json()["token"]


@app.post("/")
async def on_webhook_action(request: Request) -> Response:
    """
    Runs for each webhook action
    """
    deploy_config = file_paths.load_yaml_from_file("deploy.yaml")
    body = await request.body()
    headers = request.headers
    signature = hmac.new(
        deploy_config["git"]["webhook-secret"].encode(), body, hashlib.sha256
    ).hexdigest()

    if headers["x-hub-signature-256"] != "sha256=" + signature:
        print("Unauthorized request!")
        raise HTTPException(status_code=404, detail="Invalid body hash")

    access_token = get_access_token(get_jwt(deploy_config))
    print(access_token)

    form = await request.json()
    if headers["x-github-event"] == "check_suite":
        if form["action"] in ["requested", "rerequested"]:
            head_sha = form.get("check_suite", form.get("check_run")).get("head_sha")
            print(head_sha)
            print("Requesting a check")
            response = requests.post(
                form["repository"]["url"] + "/check-runs",
                json={
                    "accept": "application/vnd.github.v3+json",
                    "name": "configuration-validatog",
                    "head_sha": head_sha,
                },
                headers={
                    "accept": "application/vnd.github.v3+json",
                    "Authorization": "token " + access_token,
                },
            )
            print(response)
            print(response.json())
    elif headers["x-github-event"] == "check_run":
        pass

    # print(form)


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
