# CivicLens Core — Model & Data Card

CivicLens Core is a **deterministic decision layer**, not a trained model. It
consumes detections from Hugging Face perception models and turns them into
KVKK-safe, human-reviewable, prioritized municipal actions. This card documents
the perception models, datasets, and the decision rules.

## Perception models (Hugging Face)

| Model | HF ID | Role | Mode |
|-------|-------|------|------|
| DETR ResNet-50 | [`facebook/detr-resnet-50`](https://hf.co/facebook/detr-resnet-50) | Live COCO baseline — traffic infrastructure, street furniture | `live_hf` (opt-in via token) |
| YOLOv12s Road Damage | [`rezzzq/yolo12s-road-damage-rdd2022`](https://hf.co/rezzzq/yolo12s-road-damage-rdd2022) | Road damage (potholes/cracks) — demo hero | `road_damage` (precomputed) |
| Precomputed bundle | CivicLens fixtures | Reliable offline demo path (real model-shaped outputs) | `precomputed` (default) |

> **Critical limitation:** DETR is trained on COCO and **cannot** detect road
> damage (no such class). Pothole/crack detection comes from the RDD2022 YOLO
> model. The dashboard states the active mode for every result (transparency).

## Datasets (grounding)

| Dataset | Role | Reference |
|---------|------|-----------|
| COCO 2017 | DETR baseline classes (traffic light, bench, ...) | DETR paper [arXiv:2005.12872](https://arxiv.org/abs/2005.12872) |
| RDD2022 | Road damage classes D00/D10/D20/D40 | [hf.co/papers/2209.08538](https://hf.co/papers/2209.08538) (Arya et al., 2022) |

RDD2022 class mapping used by the ontology: `D00` longitudinal crack, `D10`
transverse crack, `D20` alligator crack, `D40` pothole — all normalized to
`road_damage`.

## Decision rules (deterministic, no LLM)

The label→action chain is rule-driven and lives in `docs/rules/`:

- `ontology.yaml` — raw label → normalized urban object type + KVKK label policy
- `confidence_thresholds.yaml` — per-type auto-accept / needs-review thresholds
- `priority_policy.yaml` — normalized type → maintenance priority

Chain: privacy guard → normalizer → confidence evaluator → review router →
priority engine → report builder. No step calls an LLM. OpenRouter is optional
and only rewrites the summary prose; it never changes a KVKK or priority outcome.

## KVKK / privacy

- Detection targets are **inanimate urban objects only**.
- `person`, `bicycle`, `motorcycle` are **blocked** (removed + counted).
- `car`, `truck`, `bus` are **hidden** by default (tracking/plate risk).
- **Precomputed / sample paths** use **avoidance-by-design**
  (`pii_strategy=avoidance_by_design`) — synthetic scenes with no real PII.
- **Citizen/staff uploads and Street View** use **blur-before-inference**
  (`pii_strategy=blur_applied`): HF PII detector when `HF_API_TOKEN` is set,
  otherwise whole-frame pixelation fallback. Raw images are never stored
  (`raw_image_stored=false`).

## Intended use & out-of-scope

- **Intended:** municipal maintenance triage of infrastructure (road damage,
  traffic signals, street furniture, waste assets) with human-in-the-loop review.
- **Out of scope (forbidden):** identity detection, face recognition, license
  plate reading/OCR, person or vehicle profiling/tracking.
