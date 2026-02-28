# OAuth Token Proxy (FastAPI)

Minimal backend that proxies OAuth token exchange so the frontend never sends
`client_secret`. The frontend calls this server at `/api/oauth/token`, and this
server forwards the request to the upstream OAuth server using server-side
credentials.

## Setup

1) Create `server/.env` from `server/.env.example` and fill values:
```
OAUTH_BASE_URL=
OAUTH_CLIENT_ID=
OAUTH_CLIENT_SECRET=
OAUTH_REDIRECT_URL=
OAUTH_TOKEN_PATH=/api/oauth/token
CORS_ORIGIN=http://localhost:5173
```

2) Create venv and install deps:
```
cd server
python -m venv .venv
.venv\Scripts\activate
pip install -r requirements.txt
```

3) Run dev server:
```
uv run main.py
```

## Endpoints

- `POST /api/oauth/token`
  - Body: JSON with `code`, optional `state`, optional `code_verifier`
  - Server injects `client_id`, `client_secret`, `redirect_url` from `.env`

- `GET /health`
  - Returns `{ "status": "ok" }`
