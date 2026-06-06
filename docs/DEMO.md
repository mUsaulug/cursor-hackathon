# Demo Runbook

Reproducible steps to start and verify the full stack before the final presentation.

A **live demo + presentation** is mandatory — code or screenshots alone are not eligible for prizes.

## Prize prerequisites (all four required)

The bulletin requires all of the following to qualify for a prize:

1. **Live demo** of the working system in front of the jury.
2. **Reproducible results** (same input → same output; have a fallback if live inference is slow).
3. **Runnable source code** submitted **with its incremental commit history**.
4. **KVKK deletion/anonymization document** — see [`KVKK-COMPLIANCE.md`](KVKK-COMPLIANCE.md).

## Scoring (100 pts) — build toward this

| Pts | Criterion | What it rewards |
|----:|-----------|-----------------|
| 30 | Technical workability | Error-free build, adherence to the (`masterfabric-go`) architecture, live-demo performance. **Tie-breaker.** |
| 25 | Accuracy & reliability | Computer-vision detection success rate and stability |
| 20 | Public benefit | Real potential to solve an urban/public problem |
| 10 | AI adaptation | Cursor IDE, agentic ruleset, AI-tool documentation (Cursor CLI/SDK = bonus) |
| 10 | KVKK & ethics | Full anonymization + privacy compliance |
| 5 | Presentation & docs | README quality, in-code documentation |

## Start order

Open three terminal tabs:

```bash
# Tab 1 — Backend
cd backend
go run main.go
# Expected: "backend listening on :8080"

# Tab 2 — Web
cd web
npm run dev
# Expected: "Ready on http://localhost:3000"

# Tab 3 — Mobile
cd mobile
npx expo start
# Press w for browser, or scan QR with Expo Go
```

## Expected ports

| Service | URL |
|---------|-----|
| Backend API | http://localhost:8080 |
| Web app | http://localhost:3000 |
| Expo DevTools | http://localhost:8081 (web) or QR for device |

## Health check

```bash
curl http://localhost:8080/health
# Expected response: {"status":"ok"}
```

If the response is not `{"status":"ok"}`, do not proceed to demo.

## Demo readiness checklist

- [ ] `curl http://localhost:8080/health` returns `{"status":"ok"}`
- [ ] Web app loads at `http://localhost:3000` without console errors
- [ ] Mobile app starts (Expo DevTools open, QR visible)
- [ ] No `.env` files committed (`git status` shows clean)
- [ ] No raw images or PII in the diff
- [ ] Core feature flow works end-to-end (fill in when challenge is confirmed)

## When the official challenge is announced

Update this file with:

1. The specific demo flow (user journey, step by step)
2. Any required environment variables beyond defaults
3. Sample data or mock payloads to demonstrate the feature
4. Expected AI model output format (if applicable)
5. Fallback if live AI inference is slow (pre-recorded or cached response)

## Final demo checklist (pre-presentation)

- [ ] All three services running and healthy
- [ ] Core user journey demonstrated in < 3 minutes
- [ ] Privacy compliance confirmed (no raw faces/plates visible)
- [ ] Deployment URL ready (Vercel + Render) if judges want live access
- [ ] Slides / README screenshot updated to match final UI
- [ ] **Commit history is incremental and meaningful** (not one bulk commit)
- [ ] **`docs/KVKK-COMPLIANCE.md` completed and signed** (raw data deleted + documented)
- [ ] README documents Cursor/AI usage (and CLI/SDK if used) honestly
- [ ] Backend follows the delivered `masterfabric-go` architecture
- [ ] Reproducible result confirmed (and a cached fallback ready if inference is slow)
