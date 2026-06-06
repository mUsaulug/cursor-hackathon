# Cursor Hackathon Istanbul — Urban AI

AI-driven urban solutions hackathon workspace. Three independent apps in one repo.

| App | Stack | Folder | Deploy |
|-----|-------|--------|--------|
| Web | Next.js | `web/` | Vercel |
| Mobile | Expo | `mobile/` | Expo EAS / dev build |
| API | Go | `backend/` | Render |

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
go run main.go
```

→ `http://localhost:8080` — verify with `curl http://localhost:8080/health`

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
│   ├── PRIVACY.md             # KVKK anonymization & deletion rules
│   ├── DEMO.md                # Reproducible demo runbook
│   └── DECISIONS.md           # Technical decision log
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

External services (add to the relevant app's local `.env` when wired up, never commit real values):

- `GOOGLE_STREET_VIEW_API_KEY` — Google Street View API (first 10k requests free)
- `HUGGINGFACE_TOKEN` — Hugging Face model/dataset access

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

### How we use Cursor (fill in during the event)

- **Agentic ruleset:** `.cursor/rules/` (always-apply + per-app scoped rules) drives every agent session.
- **Modes used:** _(e.g. Plan Mode for multi-file features, Agent for implementation, Debug for ...)_ — TODO
- **Prompt techniques:** _(e.g. plan-then-build, file-scoped `@Files`, smoke-test-after-change loop)_ — TODO
- **MCP:** GitHub MCP (`.cursor/mcp.json`) for branch/PR/commit assistance. — TODO: note how it helped

### Cursor CLI & SDK (bonus points)

The bulletin awards **extra AI-adaptation credit** for using Cursor CLI and/or the Cursor SDK in the
workflow. If used, document it here:

- **Cursor CLI:** _(what automation/commands — e.g. scripted agent runs, batch edits)_ — TODO
- **Cursor SDK:** _(any scripts/CI built on `@cursor/sdk` or `cursor-sdk`)_ — TODO

### Commit discipline (mandatory)

Commit in small, meaningful steps with imperative messages — the jury reviews the commit history and a
single bulk upload risks disqualification. One logical change per commit; never commit `.env`, raw
imagery, or build outputs. See [`AGENTS.md`](AGENTS.md#commit-discipline-mandatory).

### Model strategy

| Task | Model |
|------|-------|
| Single-file edits, boilerplate, refactors | `claude-sonnet-4-5` (default) |
| Planning features that span 2+ apps | `claude-opus-4` |
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
