from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from pathlib import Path
from typing import Optional
import os
from dotenv import load_dotenv
import httpx
from starlette.responses import JSONResponse, PlainTextResponse

load_dotenv(dotenv_path=Path(__file__).resolve().parent / '.env')

def require_env(name: str) -> str:
    value = os.getenv(name)
    if not value:
        raise RuntimeError(f"{name} is required")
    return value

OAUTH_BASE_URL = require_env("OAUTH_BASE_URL").rstrip("/")
OAUTH_CLIENT_ID = require_env("OAUTH_CLIENT_ID")
OAUTH_CLIENT_SECRET = require_env("OAUTH_CLIENT_SECRET")
OAUTH_REDIRECT_URL = require_env("OAUTH_REDIRECT_URL")
OAUTH_TOKEN_PATH = os.getenv("OAUTH_TOKEN_PATH", "/api/oauth/token")
TOKEN_URL = f"{OAUTH_BASE_URL}{OAUTH_TOKEN_PATH}"

CORS_ORIGIN = os.getenv("CORS_ORIGIN", "http://localhost:5173")

app = FastAPI(title="OAuth Token Proxy")

app.add_middleware(
    CORSMiddleware,
    allow_origins=[CORS_ORIGIN],
    allow_credentials=True,
    allow_methods=["POST", "GET", "OPTIONS"],
    allow_headers=["*"],
)

class TokenRequest(BaseModel):
    grant_type: str = "authorization_code"
    code: str
    redirect_url: Optional[str] = None
    client_id: Optional[str] = None
    state: Optional[str] = None
    code_verifier: Optional[str] = None

@app.get("/health")
def health():
    return {"status": "ok"}

@app.post("/api/oauth/token")
async def exchange_token(body: TokenRequest):
    payload = {
        "grant_type": body.grant_type,
        "code": body.code,
        "client_id": OAUTH_CLIENT_ID or body.client_id,
        "client_secret": OAUTH_CLIENT_SECRET,
        "redirect_url": OAUTH_REDIRECT_URL or body.redirect_url,
    }
    if body.state:
        payload["state"] = body.state
    if body.code_verifier:
        payload["code_verifier"] = body.code_verifier

    if not payload["client_id"]:
        raise HTTPException(status_code=500, detail="OAUTH_CLIENT_ID is missing")
    if not payload["redirect_url"]:
        raise HTTPException(status_code=500, detail="OAUTH_REDIRECT_URL is missing")

    async with httpx.AsyncClient(timeout=10.0) as client:
        resp = await client.post(TOKEN_URL, json=payload)

    content_type = resp.headers.get("content-type", "")
    if "application/json" in content_type.lower():
        return JSONResponse(content=resp.json(), status_code=resp.status_code)
    return PlainTextResponse(resp.text, status_code=resp.status_code)

if __name__ == "__main__":
    import uvicorn

    host = os.getenv("SERVER_HOST", "127.0.0.1")
    port = int(os.getenv("SERVER_PORT", "8443"))
    reload = os.getenv("SERVER_RELOAD", "true").lower() in ("1", "true", "yes", "on")

    uvicorn.run("main:app", host=host, port=port, reload=reload)
