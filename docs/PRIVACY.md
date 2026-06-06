# Privacy & KVKK Compliance

This hackathon project handles urban imagery and sensor data. **Faces and license plates must never be stored or exposed.** Follow these rules throughout development and demo. Violations are grounds for disqualification.

## Purpose limitation (hard rule)

- Computer-vision models may only target **inanimate urban objects** (e.g. signs, bins, road damage).
- **Forbidden:** identity detection, face recognition, license-plate reading/OCR, and any person or
  vehicle profiling or tracking.
- Faces and license plates are detected **solely to anonymize them** — never to read, match, or identify.

## Google Street View imagery

The external data source (Google Street View API) returns real street scenes that contain people and
vehicles. Treat every fetched frame as raw PII:

- Anonymize faces/plates **before** training, inference, persistence, or display.
- Keep the `GOOGLE_STREET_VIEW_API_KEY` in a local `.env` only — never commit it.
- Do not cache raw Street View frames to disk or git; keep only anonymized derivatives.

## What must never be committed

- Raw camera frames, video, or screenshots from the field
- Unprocessed uploads in `data/raw/`, `uploads/raw/`, or `private/`
- `.env` files or API keys
- Datasets containing identifiable persons or vehicle plates

Use `.env.example` with placeholder values only. Keep real secrets in local `.env` (gitignored).

## Anonymization before persistence or transmission

Apply anonymization **in memory, before** any write to disk, database, network payload, **or model
training/inference**:

1. **Detect** faces and license plates (on-device or server-side model) — only to locate regions to redact.
2. **Redact** via blur, pixelation, or crop — **irreversible**.
3. **Verify** no raw regions remain in the output buffer.
4. **Only then** persist, send, train on, or run inference over the processed result.

Never store originals “for later processing.” Delete raw buffers immediately after redaction.

## Deletion flow

| Stage | Action |
|-------|--------|
| Capture | Hold frame in memory only |
| Processing | Redact faces/plates in memory |
| Post-process | Discard raw buffer; keep anonymized output only |
| Storage | Store anonymized artifacts with TTL (e.g. 24h for demo) |
| Shutdown / demo end | Delete all stored artifacts and temp files |
| Repo | No raw media in git — ever |

Document your team’s TTL and cleanup command in the backend README when you add storage.

**At event end, deleting all raw imagery and documenting it is a prize prerequisite.** Record the
deletion/anonymization in [`KVKK-COMPLIANCE.md`](KVKK-COMPLIANCE.md) and sign it before the jury review.

## Pre-commit checklist

- [ ] No files under `data/raw/`, `uploads/raw/`, or `private/`
- [ ] No `.env` or credentials in the diff
- [ ] No sample images with visible faces or plates
- [ ] API responses do not include raw image URLs or base64 of unredacted frames

## Demo & deployment

- Use **synthetic or pre-redacted** assets for UI mockups
- Production/staging APIs must not log request bodies containing imagery
- Render/Vercel env vars: set via dashboard, never in the repository

## Reporting incidents

If raw PII slips into git: rotate any exposed keys, remove the data from history before sharing the repo, and re-run anonymization on any downstream copies.
