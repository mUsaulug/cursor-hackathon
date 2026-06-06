# Technical Decisions

Short log of significant decisions made during the hackathon. Update as you go.

## Decision log

| Time | Decision | Reason | Impact | Owner |
|------|----------|--------|--------|-------|
| Pre-hackathon | Independent folders (web/, mobile/, backend/) instead of monorepo tooling | Zero config overhead; Turborepo/Nx adds complexity for a 6-hour sprint | No shared packages; each app runs independently | Team |
| Pre-hackathon | Keep problem domain generic until official checklist | Official challenge spec overrides all assumptions | Architecture stays flexible; no hardcoded entity names | Team |
| Pre-hackathon | Privacy-first defaults (KVKK) | Legal requirement; faces and plates must not be stored or transmitted raw | All image pipelines: anonymize before persist | Team |
| Pre-hackathon | Prioritize working demo over architecture | 6-hour constraint; judges evaluate working software | Minimal abstractions; no production-grade auth or CI | Team |
| Pre-hackathon | claude-sonnet-4-5 for routine coding, claude-opus-4 for architecture | Cost and speed trade-off; Opus is slower but deeper for multi-file planning | Switch model based on task scope | Team |
| Bulletin received | Encode confirmed constraints (stack, Hugging Face, Google Street View API, Vercel/Render, commit discipline) into AGENTS/rules/README | Official bulletin is now the source of truth | Workspace matches mandatory requirements; assumptions narrowed | Team |
| Bulletin received | Backend must adopt `masterfabric-go` architecture; keep current `main.go` only as a temporary health stub | Bulletin forbids custom backend architecture; real structure delivered at 11:00 | No effort spent on custom backend design; ready to swap in delivered architecture | Team |
| Bulletin received | Add `docs/KVKK-COMPLIANCE.md` deletion/anonymization record | Signed KVKK document is a prize prerequisite | Compliance evidence ready to complete before jury | Team |
| Bulletin received | Plan to commit incrementally with meaningful messages from the start | Jury scores commit history; bulk upload risks disqualification | Continuous version control discipline | Team |
| Build kickoff | Build `vision` bounded context with Go stdlib in design-doc §4 hexagonal layout (`internal/domain|application|infrastructure|shared`) since `masterfabric-go` was not delivered to the workspace | Keep momentum; layers map 1:1 to masterfabric-go so a later swap is cheap | Backend progresses now; restructure if the real repo arrives | Team |
| Build kickoff | Exclude Redis + Kafka for the MVP | Design doc §3 lists them as "won't use", §16 calls Kafka a risk, and `backend.mdc` mandates stdlib-only | In-memory analysis store; ports defined so adapters can be added later | Team |
| Build kickoff | Precomputed-first inference; live HF DETR + OpenRouter optional behind env vars | No API keys provided; demo must run offline and reproducibly | Reliable demo path with no live dependency; live path degrades gracefully | Team |
| Build kickoff | Rule YAML files are the source of truth, read by a minimal stdlib YAML reader (no `gopkg.in/yaml.v3`) | `backend.mdc` forbids non-stdlib deps; rules must stay human-editable | `docs/rules/*.yaml` drive normalizer/priority/confidence; schema capped at 2 nesting levels | Team |
| Build kickoff | Use the Hugging Face MCP to generate REAL precomputed fixtures and KVKK-safe synthetic sample images (not mock) | MCP gives real inference (`SAM3`) and reproducible image-gen without shipping an API key | Stronger "precomputed inference, not mock" claim; PII-avoidance by design | Team |

## How to use this log

Add a row when you:
- Choose one technology over another
- Make a trade-off between quality and speed
- Deviate from the original plan
- Decide NOT to implement something (and why)

Keep entries short — one sentence per cell is enough. The goal is a 30-second explanation if a judge asks "why did you build it this way?"
