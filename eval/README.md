# CivicLens Core — Evaluation Harness

Reproducible smoke checks for the vision API (design doc §9.8).

## Automated (authoritative)

The end-to-end checks run as a Go test against the fully wired API:

```bash
cd backend
go test ./internal/app/ -run TestE2E -v
```

This asserts the §9.8 checkpoints: analyze works, JSON schema is correct,
`raw_image_stored=false`, person/car are filtered, priority is assigned, and
low-confidence detections fall to `needs_review`.

## Manual smoke (against a running server)

Start the backend (`cd backend && go run .`) then use the cases in
[`smoke_test_cases.json`](smoke_test_cases.json). Expected normalized outputs
for the precomputed scenes are in [`expected_outputs.json`](expected_outputs.json).

## sample_images/

KVKK-safe synthetic street scenes (no faces, no plates) generated via the
Hugging Face MCP image-generation tools. They are used by the web dashboard
demo and the blur verification step. Raw field imagery is never committed
(see `docs/PRIVACY.md`).
