# KVKK Data Deletion & Anonymization Record

**Prize prerequisite.** Per the official bulletin, teams must (1) anonymize faces and license plates
before any model use and (2) permanently delete all raw imagery at event end, and **document both in
writing**. This file is that document — complete and sign it before the jury review.

> Do not put any raw PII, real keys, or actual image content in this file. Record only the process and
> evidence (counts, paths, commands, timestamps).

## 1. Team & scope

| Field | Value |
|-------|-------|
| Team name | _TODO_ |
| Members | _TODO_ |
| Date | _TODO (2026-06-06)_ |
| Data sources used | _e.g. Google Street View API / provided dataset / ... — TODO_ |
| Model purpose | _Inanimate urban objects only (e.g. sign detection) — TODO_ |

## 2. Purpose-limitation attestation

- [ ] Models target **inanimate urban objects only**.
- [ ] No identity detection, face recognition, plate reading/OCR, or person/vehicle profiling was built or run.
- [ ] Faces/plates were detected **only** to anonymize them.

## 3. Anonymization

| Item | Detail |
|------|--------|
| Method | _blur / pixelate / crop — TODO_ |
| Applied before | _persistence / training / inference / transmission (all) — TODO_ |
| Tool / model used | _TODO (e.g. Hugging Face model name)_ |
| Reversible? | **No** (irreversible) |
| Verification | _how you confirmed no raw faces/plates remain — TODO_ |

## 4. Raw data deletion

| Field | Value |
|-------|-------|
| Where raw data was held | _local memory / temp dir path (never git) — TODO_ |
| Deletion command(s) run | _e.g. `rm -rf data/raw uploads/raw` — TODO_ |
| Deleted at (timestamp) | _TODO_ |
| Remaining artifacts | _anonymized derivatives only — TODO_ |
| Git history clean? | _confirm no raw media ever committed — TODO_ |

## 5. Security checklist

- [ ] No raw imagery committed to git (history checked).
- [ ] No `.env` files, API keys, or credentials in the repo.
- [ ] Raw data was never uploaded to unencrypted cloud storage or a public repo.
- [ ] API responses contain no raw image URLs or unredacted base64.

## 6. Sign-off

| Role | Name | Signature / confirmation |
|------|------|--------------------------|
| Team lead | _TODO_ | _TODO_ |
| Privacy owner | _TODO_ | _TODO_ |

_Signed at: ________________  (place / time)_
