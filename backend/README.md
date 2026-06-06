# Backend — Go API

Minimal HTTP service for the hackathon demo. Deploy target: **Render**.

> **Architecture mandate:** per the official bulletin the backend must follow the **`masterfabric-go`**
> repo architecture (delivered at event start). This `main.go` is a temporary `/health` stub — replace it
> with the delivered structure, keeping `/health` and CORS intact. Do not build a custom architecture.

## Run locally

```bash
cd backend
go run main.go
```

Server starts on `http://localhost:8080` (override with `PORT`).

## Health check

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

Render uses this endpoint for health checks.

## CORS

Enabled for local demo origins:

- `http://localhost:3000` — Next.js web
- `http://localhost:8081` — Expo web
- `http://192.168.*` / `http://10.*` — physical device on LAN

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |

Copy from `.env.example` when you add more vars. Never commit `.env`.

## Render deployment

1. Create a **Web Service** pointing at `backend/`
2. Build command: `go build -o server .`
3. Start command: `./server`
4. Health check path: `/health`

## Privacy

Do not log or persist raw imagery. See `docs/PRIVACY.md`. Anonymize faces and plates before any storage or API response that includes visual data.
