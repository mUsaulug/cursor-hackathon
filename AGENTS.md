# Agent Instructions — Urban AI Hackathon

Monorepo for a 6-hour AI-driven urban solutions hackathon.
**Prioritize a working demo, privacy compliance, and simplicity.**

## Source of truth

The **official technical checklist / rule bulletin** is the source of truth and overrides everything else.
The bulletin has arrived — its confirmed constraints are encoded below. Anything still unknown (the
specific challenge problem, the delivered Expo/Go architecture, the dataset schema) stays generic and is
marked `[ASSUMPTION]` until handed over at event start (11:00).

## Confirmed by the official bulletin

- **Tech stack (mandatory):** Web → Next.js · Mobile → Expo · Backend → Go (Golang)
- **Backend architecture (mandatory):** the **`masterfabric-go`** repo architecture MUST be used. Building
  your own backend architecture is **not accepted**. The current `backend/main.go` is a temporary health
  stub — replace it with the `masterfabric-go` structure once delivered. Do not invest in custom backend
  architecture. See `backend/AGENTS.md`.
- **AI models & datasets:** use the **Hugging Face** platform.
- **External data source:** **Google Street View API** (first 10,000 requests are free — enough for the
  event). API key lives only in `.env`; never commit it. Street View frames contain faces/plates → run
  them through anonymization before any persistence (see `docs/PRIVACY.md`).
- **Hosting:** Web → Vercel · Backend → Render.com.
- **Commit discipline (mandatory):** commit incrementally with meaningful messages. A single end-of-event
  bulk upload is grounds for **disqualification / point loss** — the jury evaluates the commit history.
- **Cursor IDE + Agentic ruleset (mandatory):** all work happens in Cursor; `.cursor/rules/` provides the
  ruleset. Document AI/Cursor usage (features, prompt techniques, CLI/SDK) in `README.md` — Cursor CLI &
  SDK usage earns **bonus** points.

## Still generic until handed over (do not hardcode)

- The specific challenge problem / municipality / object classes
- The delivered Expo & Go architecture details (given at 11:00)
- Dataset entity names and field schemas
- Computer-vision model choice (pick from Hugging Face when the task is known)

## Project layout

```
web/       → Next.js (Vercel)
mobile/    → Expo
backend/   → Go HTTP API (Render)
docs/      → Shared documentation (incl. PRIVACY.md)
```

Each folder is independent. No shared packages, no monorepo tooling.

## Privacy / KVKK (non-negotiable)

Read `docs/PRIVACY.md` before touching imagery or user data.

- **Purpose limitation:** models may only target **inanimate urban objects** (signs, bins, road damage,
  etc.). Identity detection, face recognition, plate reading/OCR, and person/vehicle profiling or tracking
  are **strictly forbidden** (disqualification).
- Faces and license plates are detected **only to anonymize them** — never to read or identify. Anonymize
  **irreversibly (blur/pixelate) before model training or inference**, persistence, and transmission.
- **Never** commit raw camera images, faces, license plates, or `.env` files
- **Never** add files under `data/raw/`, `uploads/raw/`, or `private/`
- Use `.env.example` placeholders only; real values stay local
- Delete raw buffers immediately after anonymization
- At event end, **all raw imagery must be deleted and documented** — fill in
  `docs/KVKK-COMPLIANCE.md` (a prize prerequisite).

When implementing image pipelines: redact first, store second.

## Development principles

1. **Minimal scope** — smallest change that works; no production-grade abstractions
2. **No unnecessary packages** — justify every new dependency
3. **Reproducible local demo** — document commands in root `README.md`
4. **No secrets in code** — use environment variables

## Run commands

```bash
cd web && npm run dev          # http://localhost:3000
cd mobile && npx expo start   # Expo dev server
cd backend && go run main.go  # http://localhost:8080
```

## Backend conventions

- Expose `GET /health` for Render health checks
- Enable CORS for `localhost:3000` and Expo dev origins
- Default port: `8080` (override with `PORT` env)

## Frontend conventions

- Web: `NEXT_PUBLIC_API_URL` → backend (default `http://localhost:8080`)
- Mobile: `EXPO_PUBLIC_API_URL` → backend; use LAN IP on physical devices, not `localhost`

## What not to do

- Do not globally gitignore all `.png`/`.jpg` — only raw data directories

## Cursor model strategy

- Routine edits (single file, boilerplate): `claude-sonnet`
- Architecture / multi-app planning: `claude-opus` → switch back to Sonnet for implementation
- For features touching 2+ files: use Plan Mode (`Shift+Tab` in Agent) before coding

## Cursor feature guide

This is a reference — not a checklist to run through on every message. Use it when
the context makes a suggestion genuinely useful, not as a default footer.

### Modes (`Shift+Tab` or `Cmd+.` to cycle)

- **Agent** — edits files, runs terminal; default for implementation
- **Ask** — read-only, no changes; good for exploring before touching anything
- **Plan** (`/plan`) — generates a reviewable Markdown plan before writing code; worth it when scope is unclear or work spans multiple apps
- **Debug** — instruments with runtime logs; use for hard-to-reproduce failures
- **Inline Edit** (`Cmd+K`) — surgical single-location rewrite

### Context commands (mention when they'd change the outcome)

- `@Files` — when you already know which file is relevant
- `@Codebase` — when you don't know where something lives
- `@Web` — latest external docs or APIs
- `@Branch` / `@Commit` — reviewing a diff or writing a commit message
- `@Terminal` — when the user pastes a terminal error

### Useful slash commands

- `/compress` — chat grown long and unfocused
- `/new-chat` — cleaner fresh start with explicit `@Files` context
- `/model claude-opus-4` / `/model claude-sonnet-4-5` — switch mid-session

## Commit discipline (mandatory)

- Commit in small, meaningful steps — never one bulk upload at the end.
- One logical change per commit; imperative messages (e.g. `feat(web): add sign-detection panel`,
  `fix(backend): handle empty payload`).
- Commit working states; do not commit `.env`, raw imagery, or build outputs.
- The commit history is part of the score — keep it readable.

## Before claiming done

- [ ] `go run main.go` returns `{"status":"ok"}` on `/health`
- [ ] `npm run dev` serves the web app
- [ ] `npx expo start` launches the mobile app
- [ ] No secrets or raw imagery in the diff
- [ ] Low-confidence AI outputs marked as `needs_review`, not presented as facts
- [ ] Changes committed incrementally with a clear message
- [ ] See `docs/DEMO.md` for full pre-presentation checklist (scoring + prize prerequisites)
