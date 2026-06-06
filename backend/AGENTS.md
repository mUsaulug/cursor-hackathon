# Agent Instructions — Backend (Go)

Minimal Go HTTP API for the hackathon demo. Deploy target: Render.

## Architecture (mandatory — read first)

The official bulletin requires the **`masterfabric-go`** repo architecture for the backend. **Building a
custom architecture is not accepted.** The current `main.go` is only a temporary `/health` stub so the
stack runs end-to-end. When the `masterfabric-go` structure is delivered at event start, **replace** this
stub with it (preserve the `/health` endpoint and CORS for the demo). Do not invest in a bespoke design.

## Conventions

- Standard library preferred; check `go.mod` before adding any dependency
- Always keep `GET /health` returning `{"status":"ok"}` — Render health check depends on it
- Default port `8080`; override with `PORT` env var
- CORS handled in `corsMiddleware` — extend `isAllowedOrigin`, do not replace the function
- No timeoutless outbound HTTP calls

## Privacy (non-negotiable)

Read `docs/PRIVACY.md` before touching any image or user-data endpoint.

- Never log or persist raw camera frames
- Anonymize faces and plates in memory before any write or API response
- Do not return raw image URLs or base64 of unredacted frames

## Run

```bash
cd backend
go run main.go
# → http://localhost:8080
curl http://localhost:8080/health
# → {"status":"ok"}
```

## Before claiming done

- [ ] `go run main.go` succeeds with no errors
- [ ] `/health` returns `{"status":"ok"}`
- [ ] No secrets or raw imagery in the diff
- [ ] New endpoints have CORS and correct HTTP method checks
