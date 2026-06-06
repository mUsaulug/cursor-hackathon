# KVKK Blur Verification

Visual proof that the CivicLens blur utility (`backend/internal/shared/imaging`,
exposed via `backend/cmd/anonymize`) irreversibly anonymizes PII regions
(faces, people, license plates) **before** any inference or persistence
(design doc 5.3; `docs/PRIVACY.md`).

- `street_people_blurred.png` — a synthetic street scene (no real identity) in
  which the pedestrian and the vehicle have been irreversibly pixelated.
- The raw (pre-blur) frame is intentionally **not** committed: KVKK requires raw
  imagery to be discarded after anonymization. Only the anonymized derivative is
  kept.

## Reproduce

```bash
cd backend
go run ./cmd/anonymize \
  -in /path/to/frame.png \
  -out blurred.png \
  -regions "x,y,w,h;x,y,w,h" \
  -block 22
```

The blur is averaging-based pixelation and cannot be reversed. WebP input is not
supported (Go stdlib); convert to PNG/JPEG first.
