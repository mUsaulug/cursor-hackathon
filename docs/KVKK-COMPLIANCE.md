# KVKK Data Deletion & Anonymization Record

**Prize prerequisite.** Per the official bulletin, teams must (1) anonymize faces and license plates
before any model use and (2) permanently delete all raw imagery at event end, and **document both in
writing**. This file is that document — complete and sign it before the jury review.

> Do not put any raw PII, real keys, or actual image content in this file. Record only the process and
> evidence (counts, paths, commands, timestamps).

## 1. Team & scope

| Field | Value |
|-------|-------|
| Team name | CivicLens (Urban AI Hackathon) |
| Members | _Fill team roster before jury sign-off_ |
| Date | 2026-06-06 |
| Data sources used | HF models (DETR, RDD2022 YOLO); synthetic images generated via HF MCP image-gen. Google Street View: opt-in, disabled by default (no key used in MVP). |
| Model purpose | Inanimate urban objects only (road damage, traffic signals, street furniture, waste assets). |

## 2. Purpose-limitation attestation

- [x] Models target **inanimate urban objects only**.
- [x] No identity detection, face recognition, plate reading/OCR, or person/vehicle profiling was built or run.
- [x] Faces/plates were detected **only** to anonymize them (blur utility), never to read or identify.

## 3. Anonymization

| Item | Detail |
|------|--------|
| Method | Irreversible **pixelation** (block averaging) over PII regions; whole-frame fallback when HF token absent |
| Applied before | Inference, persistence, and transmission (in-memory, before any write) |
| Tool / model used | `backend/internal/shared/imaging`, HF DETR-based PII detector when `HF_API_TOKEN` set; privacy guard blocks person/vehicle classes in output |
| Reversible? | **No** (irreversible) |
| Verification | Visual proof in `eval/redaction/street_people_blurred.png` (pedestrian + vehicle pixelated); unit tests assert region-only, averaged (non-recoverable) output. MVP runtime stores no raw image (`raw_image_stored=false`). |

## 4. Raw data deletion

| Field | Value |
|-------|-------|
| Where raw data was held | In-memory only during request handling; the one synthetic verification frame stayed in `/tmp` and was not committed |
| Deletion command(s) run | `rm -f /tmp/street_people_raw.webp /tmp/street_people_raw.png` (raw verification frame); no `data/raw/` used |
| Deleted at (timestamp) | _confirm at event end — TODO_ |
| Remaining artifacts | Anonymized derivative only (`eval/redaction/street_people_blurred.png`) |
| Git history clean? | Yes — no raw camera frames, faces, or plates committed; only anonymized/synthetic assets |

## 5. Security checklist

- [x] No raw imagery committed to git (history checked).
- [x] No `.env` files, API keys, or credentials in the repo.
- [x] Raw data was never uploaded to unencrypted cloud storage or a public repo.
- [x] API responses contain no raw image URLs or unredacted base64.

## 6. Sign-off

| Role | Name | Signature / confirmation |
|------|------|--------------------------|
| Team lead | _TODO_ | _TODO_ |
| Privacy owner | _TODO_ | _TODO_ |

_Signed at: ________________  (place / time)_
