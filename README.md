# CivicLens Core — Urban AI (Cursor Hackathon Istanbul)

**CivicLens Core** turns street-level AI detections into KVKK-safe,
human-reviewable, maintenance-prioritized municipal actions. The model only says
*what* it sees; the Go core decides *what it means for the municipality* —
deterministically, with privacy and human review built in.

```
HF detection  ->  CivicLens Core (Go)              ->  Municipal action
{traffic light}   privacy guard -> normalize ->         {traffic_signal,
 score 0.91        confidence -> review -> priority       priority: medium,
                                                          review: auto_accepted,
                                                          kvkk_safe: true}
```

Three independent apps in one repo:

| App | Stack | Folder | Deploy |
|-----|-------|--------|--------|
| Web dashboard | Next.js | `web/` | Vercel |
| Mobile field view | Expo | `mobile/` | Expo EAS / dev build |
| Vision API | Go (hexagonal) | `backend/` | Render |

The backend follows the design doc's hexagonal layout
(`internal/domain | application | infrastructure | shared`) using the Go standard
library only. `masterfabric-go` was not delivered to the workspace; the layers
map 1:1 so it can be swapped in later (see `docs/DECISIONS.md`).

**Mandatory constraints (official bulletin):**

- Backend **must** follow the `masterfabric-go` architecture (delivered at event start) — no custom backend architecture.
- AI models/datasets via **Hugging Face**; external imagery via **Google Street View API** (first 10k requests free).
- **Commit incrementally** with meaningful messages — the jury evaluates the commit history. No single bulk upload.
- All development in **Cursor IDE** with an agentic ruleset (`.cursor/rules/`); document AI usage below.

**Privacy / KVKK:** Models target inanimate urban objects only — no identity/face recognition, plate reading, or profiling. Faces and license plates must be anonymized before any use and never stored or exposed. See [`docs/PRIVACY.md`](docs/PRIVACY.md).

## Quick start

### Prerequisites

- Node.js 18+
- Go 1.22+
- npm (web) and Expo CLI via `npx` (mobile)

### 1. Backend (Go)

```bash
cd backend
go run .
```

→ `http://localhost:8080` — verify with `curl http://localhost:8080/health`

Try the vision API (no key needed — precomputed-first):

```bash
curl -X POST "http://localhost:8080/api/v1/vision/analyze?source_ref=sample_road_damage&mode=road_damage"
curl http://localhost:8080/api/v1/vision/demo-results
```

Run the backend tests (unit + E2E):

```bash
cd backend && go test ./...
```

### 2. Web (Next.js)

```bash
cd web
cp .env.example .env.local   # optional — defaults to localhost backend
npm install                  # first time only
npm run dev
```

→ `http://localhost:3000`

### 3. Mobile (Expo)

```bash
cd mobile
cp .env.example .env           # optional — set LAN IP for physical device
npm install                    # first time only
npx expo start
```

→ Scan QR with Expo Go, or press `w` for web.

**Physical device tip:** Set `EXPO_PUBLIC_API_URL` to your machine's LAN IP (e.g. `http://192.168.1.10:8080`), not `localhost`.

## Project structure

```
.
├── README.md                  # This file
├── AGENTS.md                  # Cursor / AI agent instructions (root)
├── .cursorignore              # Full block: secrets, raw data, node_modules
├── .cursorindexingignore      # Index-only block: build outputs, lock files
├── .cursor/
│   ├── rules/
│   │   ├── hackathon.mdc      # Always-apply: project scope, privacy, demo
│   │   ├── web.mdc            # Auto-attached: Next.js rules
│   │   ├── mobile.mdc         # Auto-attached: Expo rules
│   │   └── backend.mdc        # Auto-attached: Go rules
│   └── mcp.json               # GitHub MCP server config
├── docs/
│   ├── 2026-06-06-civiclens-core-design.md  # Architecture design doc
│   ├── MODEL_CARD.md          # Models, datasets, decision rules
│   ├── rules/                 # ontology / priority / confidence YAML (SSOT)
│   ├── PRIVACY.md             # KVKK anonymization & deletion rules
│   ├── KVKK-COMPLIANCE.md     # Signed deletion/anonymization record
│   ├── DEMO.md                # Reproducible demo runbook
│   └── DECISIONS.md           # Technical decision log
├── eval/                      # Smoke harness + synthetic samples + blur proof
├── web/                       # Next.js frontend
├── mobile/                    # Expo mobile app
└── backend/                   # Go HTTP API
```

## Environment variables

| App | File | Variable | Default |
|-----|------|----------|---------|
| Web | `web/.env.local` | `NEXT_PUBLIC_API_URL` | `http://localhost:8080` |
| Mobile | `mobile/.env` | `EXPO_PUBLIC_API_URL` | `http://localhost:8080` |
| Backend | `backend/.env` | `PORT` | `8080` |

External services (all optional — the demo runs precomputed-first without any key):

- `HF_API_TOKEN` — enables live HF DETR inference (`mode=live_hf`); absent → precomputed fallback
- `HF_INFERENCE_BASE_URL` — override the HF inference base (default `https://router.huggingface.co/hf-inference/models`)
- `OPENROUTER_API_KEY` / `OPENROUTER_MODEL` — optional LLM that rewrites only the report summary prose
- `GOOGLE_STREET_VIEW_API_KEY` — enables the opt-in Street View source (blur-before-inference)

Never commit `.env` files or real keys. Use `.env.example` placeholders only.

## Vision API (backend)

Base URL: `NEXT_PUBLIC_API_URL` / `EXPO_PUBLIC_API_URL` (default `http://localhost:8080`).

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/health`, `/health/live`, `/health/ready` | Health checks |
| POST | `/api/v1/vision/analyze` | Analyze an uploaded image (multipart `image`) or a sample (`?source_ref=&mode=`) |
| GET | `/api/v1/vision/analyze/{id}` | Fetch a stored analysis |
| GET | `/api/v1/vision/demo-results` | Two reproducible precomputed scenes |
| GET | `/api/v1/vision/model-info` | Active modes/models + limitations (transparency) |
| GET | `/api/v1/vision/privacy-report` | KVKK status of the latest analysis |
| GET | `/api/v1/vision/summary` | Maintenance report for the latest analysis |
| GET | `/api/v1/vision/reviews` | Needs-review detections (human-in-the-loop queue) |
| PATCH | `/api/v1/vision/reviews/{detectionId}` | Human accept/reject override |
| POST | `/api/v1/vision/report` | Maintenance report (OpenRouter prose or local fallback) |

Decision rules live in [`docs/rules/`](docs/rules) and the model card in
[`docs/MODEL_CARD.md`](docs/MODEL_CARD.md).

Never commit `.env` files. Use `.env.example` placeholders only.

## Privacy rules (summary)

- No raw camera images in git
- No faces or license plates stored — anonymize before persistence or transmission
- Raw data folders (`data/raw/`, `uploads/raw/`, `private/`) are gitignored
- Delete temp buffers after redaction

Full checklist: [`docs/PRIVACY.md`](docs/PRIVACY.md)

## Deployment

- **Web:** Connect `web/` to Vercel; set `NEXT_PUBLIC_API_URL` to your Render URL
- **Backend:** Render Web Service from `backend/`; health check `/health`
- **Mobile:** Expo build when ready; point `EXPO_PUBLIC_API_URL` at production API

## AI-Driven development (Cursor)

> The jury scores AI adaptation (10 pts) and documentation (5 pts). Keep this section **honest and
> specific** as you build — record what you actually used, not aspirational claims. Do not fabricate
> tools, metrics, or results.

See [`AGENTS.md`](AGENTS.md) for agent rules: minimal diffs, no secrets, strict privacy.

### How we used Cursor

- **Agentic ruleset:** `.cursor/rules/` (always-apply + per-app scoped rules) constrained every session
  — e.g. `backend.mdc` "stdlib only" drove a custom minimal YAML reader instead of adding a dependency.
- **Plan Mode:** the whole CivicLens Core build started from a reviewed plan in `.cursor/plans/`; the
  agent reconciled the design doc against the kickoff prompt (caught Redis/Kafka and section-count
  discrepancies) before writing code.
- **Parallel subagents:** the web dashboard and the mobile field view were built by two background
  subagents running on `composer-2.5-fast` while the parent agent (Claude) built the Go blur/OpenRouter
  layer — independent workstreams in parallel.
- **Hugging Face MCP:** used to (1) verify `facebook/detr-resnet-50` and `rezzzq/yolo12s-road-damage-rdd2022`
  metadata and the inference response schema, (2) generate KVKK-safe synthetic street images
  (`eval/sample_images`) with the Z-Image / FLUX spaces, and (3) produce the blur-verification scene.
- **Verification loop:** every phase ran `go build/vet/test` + runtime `curl` smoke + `ReadLints` before
  an incremental commit.

### Cursor CLI & SDK

- **Cursor CLI / SDK:** not used in this build. (The HF MCP tool surface — not the Cursor SDK — drove the
  AI integrations.) Recorded honestly per the bulletin.

### Commit discipline (mandatory)

Commit in small, meaningful steps with imperative messages — the jury reviews the commit history and a
single bulk upload risks disqualification. One logical change per commit; never commit `.env`, raw
imagery, or build outputs. See [`AGENTS.md`](AGENTS.md#commit-discipline-mandatory).

### Model strategy

| Task | Model |
|------|-------|
| Cross-cutting domain/use-case design + integration | parent agent (Claude) |
| Well-scoped UI subagents (web, mobile) | `composer-2.5-fast` |
| Single-file edits, boilerplate, refactors | `claude-sonnet-4-5` |
| Tab autocomplete | Cursor built-in (leave as-is) |

### Plan Mode (Shift+Tab in Agent)

For any feature touching 2+ files: open Agent (`Cmd+I`) → press `Shift+Tab` → describe the feature → review the generated plan → click Build. Plans save to `.cursor/plans/` and can be committed for team review.

### Rules

`.cursor/rules/` contains four scoped rule files:
- `hackathon.mdc` — always-apply project constraints, privacy, demo commands
- `web.mdc` — auto-loaded when editing `web/` files
- `mobile.mdc` — auto-loaded when editing `mobile/` files
- `backend.mdc` — auto-loaded when editing `backend/` Go files

### MCP

`.cursor/mcp.json` configures the GitHub MCP server. Set `GITHUB_TOKEN` in your local `.env` to enable agent-driven branch and PR management from within Cursor.
