# CivicLens Full Repo Review — 2026-06-06

Sistematik tam-repo analizi: backend, web, mobile, docs, infra, deploy.  
Metodoloji: faz bazlı kod incelemesi + `verification-before-completion` (komut çıktıları) + subagent review.

---

## Executive Summary

1. **Backend güçlü:** Hexagonal mimari, Wave 2 E2E testleri (`go test ./...` geçiyor), deterministik vision pipeline, audit middleware, in-memory store'da ham görüntü yok.
2. **KVKK kritik risk:** `HF_API_TOKEN` yokken gerçek vatandaş/saha fotoğrafları noop anonymizer ile işleniyor; `POST /api/v1/vision/analyze` multipart yolu anonymizer'ı tamamen atlıyor ve canlı HF'ye ham byte gönderebilir.
3. **Demo akışı kısmen kopuk:** Web operator review → task oluşturma çalışır; web saha personeli start/evidence/close yapmıyor; mobile kısmen yapıyor; manager close hiçbir client'ta yok. OperatorView API sözleşmesi uyumsuz.
4. **Deploy tutarsızlığı:** Vercel `civiclens-backend-1rkq.onrender.com` (200 OK); EAS `civiclens-backend.onrender.com` (404). Prod demo farklı backend'lere işaret edebilir.
5. **Ödül önkoşulları eksik:** `KVKK-COMPLIANCE.md` imzasız; `DEMO.md` 3 dk script ve checklist'ler boş; `MODEL_CARD.md` Wave 2 blur stratejisiyle çelişiyor.

---

## Faz 0 — Gereksinim Matrisi

| Gereksinim | Kaynak | Durum | Kanıt |
|------------|--------|-------|-------|
| Next.js web | Bulletin | ✅ | `web/` — build geçti |
| Expo mobile | Bulletin | ✅ | `mobile/` — `tsc --noEmit` geçti |
| Go backend | Bulletin | ✅ | `backend/` — 27 route, health OK |
| masterfabric-go mimarisi | Bulletin | ⚠️ Sapma | Hexagonal interim; `DECISIONS.md` belgelenmiş |
| Hugging Face modelleri | Bulletin | ✅ | DETR + RDD2022 precomputed + opt-in live |
| Google Street View | Bulletin | ❌ Wire yok | `streetview/` var, `app.go`'da bağlı değil |
| KVKK anonymization | Bulletin | ⚠️ Koşullu | HF token yokken noop; vision upload bypass |
| Incremental commits | Bulletin | ✅ | 30+ anlamlı commit (Wave 1→2) |
| Cursor dokümantasyonu | Bulletin | ✅ | `README.md` AI bölümü dolu; CLI/SDK kullanılmamış |
| KVKK imzalı belge | Ödül prereq | ❌ | `KVKK-COMPLIANCE.md` TODO alanları |
| Canlı demo script | `DEMO.md` | ❌ | "fill in when challenge confirmed" |
| Postgres/Redis/Kafka | Wave 2 tasarım | ⚠️ Deferred | `docker-compose.yml` + `init.sql` hazır; backend in-memory |
| 4 rol (citizen/field/operator/manager) | Wave 2 | ⚠️ Kısmi | Web 4 rol; mobile 2 rol |
| Blur-before-inference (Wave 2) | `DECISIONS.md` | ⚠️ HF'ye bağlı | `app.go:89-96` noop fallback |

### Repo envanteri

- ~103 kaynak dosya (`.go`, `.ts`, `.tsx`)
- Backend: 20 `*_test.go`, 27 HTTP route
- Web: 0 test, ESLint 1 hata, build OK
- Mobile: 0 test, TypeScript OK
- Kod içi `TODO`/`FIXME`: yok (sadece docs'ta)

### Bilinen riskler (önceden işaretli — doğrulandı)

| Risk | Doğrulama |
|------|-----------|
| masterfabric-go yok | `DECISIONS.md` satır 18, `backend/AGENTS.md` |
| In-memory persistence | `store/*_inmemory.go`, restart'ta veri kaybı |
| Street View unwired | `grep streetview *.go` — sadece adapter + test |
| Deploy URL mismatch | curl: `-1rkq` → 200, `civiclens-backend` → 404 |
| Web test yok | glob `*.test.*` → 0 |
| KVKK doc eksik | `KVKK-COMPLIANCE.md` §1,4,5,6 TODO |

---

## Critical (Bloklayıcı)

### C1 — Gerçek fotoğraflarda noop anonymizer (HF token yokken)

**Dosya:** `backend/internal/app/app.go:89-96`, `anonymizer/noop_detector.go:8-10`

Wave 2 kararı aktif blur-before-inference olmasına rağmen, `HF_API_TOKEN` absent iken tüm citizen/staff upload'ları `NoopDetector` kullanıyor. No-op açıkça "real citizen/staff photos için kullanılmamalı" diyor. Bölgeler bulunamazsa görüntü **değişmeden** encode edilir (`anonymizer.go:68-74`).

**Etki:** Varsayılan local/demo ortamında `POST /api/v1/reports` ve evidence upload KVKK ihlali riski. Jüri 10 puanlık KVKK kriterinde ve diskalifiye kuralında doğrudan risk.

**Öneri:** HF token yokken fail-closed (400 reject) veya her zaman whole-frame blur (detector'sız).

---

### C2 — Vision analyze API anonymizer bypass + ham byte HF'ye

**Dosya:** `handler/vision/handler.go:67-79`, `handler/vision/mapper.go:48-63`, `huggingface/detr_adapter.go:27-31`

`POST /api/v1/vision/analyze` multipart upload doğrudan `AnalyzeImageUseCase`'e gider; intake anonymizer zincirinden geçmez. HF token varken ham byte'lar Hugging Face'e gönderilir.

**Etki:** "Anonymize before inference, persistence, or transmission" kuralına aykırı yan kapı.

**Öneri:** Multipart upload'u anonymizer'dan geçir veya HF token varken upload'u reddet.

---

### C3 — Deploy URL tutarsızlığı (web vs mobile prod)

| Dosya | URL | Smoke test |
|-------|-----|------------|
| `web/vercel.json` | `https://civiclens-backend-1rkq.onrender.com` | `{"status":"ok"}` HTTP 200 |
| `mobile/eas.json` | `https://civiclens-backend.onrender.com` | `Not Found` HTTP 404 |
| `render.yaml` | service name `civiclens-backend` | — |

**Etki:** Vercel web ve EAS mobile preview farklı (veya ölü) backend'lere bağlanır; canlı demo tutarsız veri gösterir.

**Öneri:** `eas.json`'u `-1rkq` URL'siyle hizala veya tek canonical Render URL belirle.

---

### C4 — OperatorView API sözleşmesi uyumsuzluğu

**Dosya:** `web/components/OperatorView.tsx:66-67,148-155` vs `handler/task/handler.go:47-61`

Web `ReviewResponse { report, task? }` bekliyor. Backend accept'te bare `Task` (201), reject'te bare `Report` (200) döndürüyor.

**Etki:** "Son İnceleme Sonucu" bölümü boş/hatalı kartlar gösterir; operator demo akışı kırık görünür.

**Öneri:** Backend'i wrapper döndürecek şekilde güncelle veya frontend parsing'i düzelt.

---

### C5 — KVKK compliance belgesi tamamlanmamış (ödül prerequisite)

**Dosya:** `docs/KVKK-COMPLIANCE.md`

Eksik: team name, members, deletion timestamp, §5 security checklist (4 madde), §6 imzalar.

**Etki:** Bulletin'e göre ödül için zorunlu belge eksik.

---

## Important (Demo öncesi)

### I1 — Web field staff workflow eksik

`web/components/FieldStaffView.tsx` sadece `GET /tasks`. Backend'de `start`, `evidence`, `close` var; mobile'da start+evidence var.

Web `field_staff` rolü report gönderebilir ama task lifecycle yürütemez.

---

### I2 — Manager close-task UI yok (her iki platform)

`POST /api/v1/tasks/{id}/close` backend'de manager capability ile var; web ve mobile'da UI yok. E2E demo yolculuğu manager adımında kesiliyor.

---

### I3 — Mobile ReportTab görüntüsüz POST (her zaman fail)

**Dosya:** `mobile/App.tsx:52-66`

"Bildirim" tab'ı `image` olmadan `POST /api/v1/reports` yapıyor. Backend image zorunlu (`handler/report/handler.go`).

---

### I4 — Demo sample görselleri 404

**Dosya:** `web/components/AiAnalysisDashboard.tsx:31-37`

`/samples/street_traffic_01.webp`, `/samples/road_pothole_01.webp` referans ediliyor; `web/public/samples/` yok. API `source_ref` çalışabilir ama preview `<img>` kırık.

---

### I5 — Mobile `needs_review` gösterilmiyor

**Dosya:** `mobile/components/CitizenView.tsx:202-218`

`review_status` tipi var ama UI'da yok. Düşük güven tespitleri fact gibi sunuluyor — bulletin kuralına aykırı.

Web'de `DetectionCard` + review aksiyonları doğru yapılmış.

---

### I6 — RBAC: spoofable X-Role + açık GET endpoint'leri

**Dosya:** `identity/model.go`, tüm handler'lar

- Auth yok; client `X-Role` header'ı serbestçe set edebilir (demo OK, prod/KVKK accountability değil).
- `GET /reports`, `GET /tasks`, `GET /vision/*`, `PATCH /vision/reviews/{id}` RBAC'siz.
- Identity model'de read capability tanımı yok.

---

### I7 — Privacy metadata yanlış raporlanıyor

**Dosya:** `application/vision/privacy_guard.go:25-36`

Pipeline her zaman `Anonymized: false`, `PIIStrategyAvoidanceByDesign` döndürüyor. Intake'te blur uygulanmış olsa bile `/vision/privacy-report` yanlış bilgi verir.

---

### I8 — MODEL_CARD vs Wave 2 strateji drift

**Dosya:** `docs/MODEL_CARD.md:48-51` vs `docs/DECISIONS.md:24`

MODEL_CARD hâlâ MVP stratejisi olarak `avoidance_by_design` diyor. Wave 2 kararı `blur_applied` (gerçek fotoğraflar). KVKK doc blur/pixelation diyor — tutarsız.

---

### I9 — `.env.example` drift

| Dosya | Eksik |
|-------|-------|
| `backend/.env.example` | Sadece `PORT`; `HF_API_TOKEN`, `OPENROUTER_*`, `ALLOWED_ORIGIN` yok |
| Root `.env.example` | `GOOGLE_STREET_VIEW_API_KEY` yok |
| `render.yaml` | Tüm secret'lar comment'li (bilinçli ama prod HF yok → C1 ile birleşir) |

---

### I10 — Duplicate task oluşturma mümkün

**Dosya:** `application/tasking/service.go:45-74`

`CreateFromReport` report status kontrolü yapmıyor. Zaten `task_created` veya `rejected` report'tan tekrar task oluşabilir.

---

### I11 — Google Street View bulletin gereksinimi karşılanmıyor

Adapter (`infrastructure/streetview/`) testli ama `app.go`'da wire edilmemiş. Bulletin external data source olarak Street View API zorunlu kılıyor.

---

### I12 — masterfabric-go bulletin uyumu

Interim hexagonal layout iyi belgelenmiş (`DECISIONS.md`) ama bulletin "custom architecture not accepted" diyor. Jüri sorusuna hazır migration narrative gerekli.

---

### I13 — DEMO.md eksik

- 3 dakikalık jüri scripti yok
- Demo readiness + final checklist tümü `[ ]`
- `go run main.go` vs README `go run .` tutarsızlığı

---

### I14 — Verification service sessiz hata yutma

**Dosya:** `application/verification/service.go:114-119`

`_ = s.tasks.Save(ctx, t)` — task state divergence riski.

---

## Minor

| ID | Bulgu | Konum |
|----|-------|-------|
| M1 | Web ESLint hatası: `setState` in effect | `web/components/CitizenView.tsx:64` |
| M2 | Stub routes `/tasks`, `/reports` → redirect | `web/app/tasks/page.tsx` |
| M3 | Web citizen `source_type: citizen_mobile` (web değil) | `web/app/page.tsx:26` |
| M4 | Hardcoded `assigned_to: saha_ekip_1` | `OperatorView.tsx`, `FieldStaffView.tsx` |
| M5 | `priority_factors.yaml` embed edilmemiş, kullanılmıyor | `docs/rules/`, `config/rules.go` |
| M6 | `DuplicateCount` priority'ye etki etmiyor | `report_inmemory.go`, `priority_engine.go` |
| M7 | Kullanılmayan report status sabitleri | `domain/report/model.go:38-41` |
| M8 | `ReviewItem` tipi web'de unused | `web/app/types.ts` |
| M9 | `web/README.md` create-next-app boilerplate | `web/README.md` |
| M10 | `docker-compose.yml` stale comment `go run ./cmd/civiclens` | satır 3 — böyle binary yok |
| M11 | Mobile `RECORD_AUDIO` permission kullanılmıyor | `mobile/app.json` |
| M12 | Mobile duplicate report UI (Analiz vs Bildirim) | `App.tsx` |
| M13 | Mobile evidence placeholder (app icon) | `FieldStaffView.tsx:217-224` |
| M14 | Audit trail actor/resource ID yok | `middleware/audit.go:53-60` |
| M15 | Analytics operator erişimi capability map dışı | `analytics/handler.go:28` |

---

## Nice-to-have

| ID | Bulgu |
|----|-------|
| N1 | Web unit/E2E testleri (Playwright smoke) |
| N2 | Vision `model-info`, `privacy-report`, `reviews` list web AI paneline bağlanabilir |
| N3 | URL-driven role/tab routing (bookmark) |
| N4 | Polling / optimistic updates |
| N5 | `priority_factors.yaml` multi-factor scoring implementasyonu |
| N6 | Cursor CLI/SDK bonus puanı için entegrasyon |
| N7 | Postgres adapter (W2-5) — product sonrası |

---

## Strengths (İyi Yapılanlar)

1. **Temiz hexagonal mimari** — domain/application/infrastructure ayrımı; handler'lar ince.
2. **Deterministik decision chain** — privacy → normalize → confidence → review → priority; LLM sadece prose.
3. **Güçlü backend E2E** — citizen → operator → field → manager lifecycle testli (`internal/app/*_e2e_test.go`).
4. **KVKK-first intake sırası** — anonymize → analyze → route (`create_report_usecase.go`).
5. **Ham görüntü persist edilmiyor** — store'larda sadece `image_ref` metadata.
6. **Precomputed-first demo** — HF token olmadan reproducible sonuç.
7. **Audit middleware** — tüm mutating request'ler loglanıyor.
8. **Embedded YAML rules** — ontology, priority, confidence, routing SSOT.
9. **Web AI dashboard** — bbox overlay, `needs_review` badge, KVKK panel (vision tab).
10. **Incremental commit history** — Wave 1→2 anlamlı mesajlarla.
11. **Eval harness** — `eval/smoke_test_cases.json` + `eval/redaction/` blur kanıtı.
12. **Root README kapsamlı** — API tablosu, privacy, Cursor kullanımı.

---

## Wave 2 E2E Demo Yolculuğu Haritası

| Adım | API | Web UI | Mobile UI | Durum |
|------|-----|--------|-----------|-------|
| 1. Citizen report | `POST /reports` | CitizenView ✅ | CitizenView ✅ | OK |
| 2. Anonymize + vision | backend intake | — | — | ⚠️ HF yokken noop |
| 3. Operator queue | `GET /reports` | OperatorView ✅ | — | OK |
| 4. Accept/reject | `POST /reports/{id}/review` | OperatorView ⚠️ | — | Response parse bug |
| 5. Task list | `GET /tasks` | FieldStaffView (read) | FieldStaffView ✅ | Web kısıtlı |
| 6. Start task | `POST /tasks/{id}/start` | — | FieldStaffView ✅ | Web eksik |
| 7. Upload evidence | `POST /tasks/{id}/evidence` | — | FieldStaffView ⚠️ | Placeholder img |
| 8. AI verify | backend auto | — | — | OK (backend) |
| 9. Manager close | `POST /tasks/{id}/close` | — | — | **UI yok** |
| 10. Analytics | `GET /analytics/summary` | ManagerView ✅ | — | OK |
| 11. Audit log | `GET /audit` | — | — | **UI yok** |
| 12. Vision demo | `POST /vision/analyze` | AiAnalysisDashboard ⚠️ | demo-results ✅ | Sample 404 web |

---

## API Endpoint Kullanım Matrisi (Client vs Backend)

### Kullanılan (web ve/veya mobile)

`POST /reports`, `GET /reports`, `POST /reports/{id}/review`, `GET /tasks`, `POST /tasks/{id}/start`, `POST /tasks/{id}/evidence`, `GET /analytics/summary`, `POST /vision/analyze`, `GET /vision/summary`, `PATCH /vision/reviews/{id}`, `GET /vision/demo-results`

### Backend'de var, client'ta yok

`GET /reports/{id}`, `POST /tasks`, `GET /tasks/{id}`, `POST /tasks/{id}/assign`, `GET /tasks/{id}/evidence`, `POST /tasks/{id}/close`, `GET /audit`, `GET /vision/analyze/{id}`, `GET /vision/model-info`, `GET /vision/privacy-report`, `GET /vision/reviews`, `POST /vision/report`

---

## Backend Test Coverage Gap

| Paket | Test |
|-------|------|
| `internal/app` | ✅ 5 E2E dosyası |
| `application/vision` | ✅ pipeline, router, usecase |
| `application/intake` | ✅ usecase |
| `application/verification` | ⚠️ sadece pure function |
| `application/tasking` | ❌ |
| `application/analytics` | ❌ |
| Handlers (task, verification, vision, analytics) | ❌ (report handler + e2e) |
| `huggingface`, `openrouter`, `middleware` | ❌ |
| KVKK: upload without HF token blurred | ❌ |

---

## Jüri Scoring Checklist (`DEMO.md` 100 puan)

| Pts | Kriter | Değerlendirme | Gap |
|----:|--------|---------------|-----|
| 30 | Teknik çalışırlık | Backend test OK, web build OK, lint 1 hata | OperatorView bug, deploy URL, E2E UI kopuk |
| 25 | Doğruluk/güvenilirlik | Precomputed reproducible; live HF opt-in | Vision upload bypass; mobile needs_review |
| 20 | Kamu yararı | Güçlü municipal ops narrative | — |
| 10 | AI/Cursor adaptasyonu | README dolu, agentic rules | CLI/SDK bonus yok |
| 10 | KVKK | Pipeline tasarımı iyi | C1, C2, C5; MODEL_CARD drift |
| 5 | Sunum/docs | README, design docs, sunum HTML | DEMO script, KVKK imza |

---

## Recommended Fix Order

| Öncelik | Fix | Effort |
|---------|-----|--------|
| 1 | KVKK fail-closed veya whole-frame blur (C1) | S |
| 2 | Vision upload anonymizer (C2) | M |
| 3 | Deploy URL hizala (C3) | S |
| 4 | OperatorView response fix (C4) | S |
| 5 | KVKK-COMPLIANCE.md doldur + imzala (C5) | S |
| 6 | `public/samples/` veya preview kaldır (I4) | S |
| 7 | Mobile ReportTab düzelt/kaldır (I3) | S |
| 8 | Mobile needs_review UX (I5) | S |
| 9 | Manager close UI web (I2) | M |
| 10 | Web field staff start/evidence (I1) | M |
| 11 | MODEL_CARD + backend .env.example güncelle (I8, I9) | S |
| 12 | DEMO.md 3 dk script yaz (I13) | S |
| 13 | Street View wire veya bulletin deviation doc (I11) | M-L |

---

## Verification Log

Komutlar bu review oturumunda çalıştırıldı (2026-06-06).

| Komut | Sonuç |
|-------|-------|
| `cd backend && go test ./... -count=1` | **PASS** — tüm paketler ok |
| `cd backend && go vet ./...` | **PASS** — çıktı yok (exit 0) |
| `cd web && npm run lint` | **FAIL** — 1 error `CitizenView.tsx:64` react-hooks/set-state-in-effect |
| `cd web && npm run build` | **PASS** — Next.js 16.2.7, 6 static routes |
| `cd mobile && npx tsc --noEmit` | **PASS** — exit 0 |
| `curl https://civiclens-backend-1rkq.onrender.com/health` | **200** `{"status":"ok"}` |
| `curl https://civiclens-backend.onrender.com/health` | **404** Not Found |
| `git log --oneline -30` | 30 anlamlı commit |
| `rg TODO\|FIXME --glob '*.{go,ts,tsx}'` | Kodda yok |
| `git check-ignore .env` | `.gitignore:21` — ignore OK |

---

## Faz Tamamlama Özeti

| Faz | Durum | Ana çıktı |
|-----|-------|-----------|
| 0 Baseline | ✅ | Gereksinim matrisi (yukarı) |
| 1 Backend | ✅ | 5 Critical/Important backend bulgu |
| 2 Web | ✅ | API mapping, lint/build, UX gaps |
| 3 Mobile | ✅ | Parity matrix, ReportTab bug |
| 4 Cross-cutting | ✅ | E2E yolculuk haritası |
| 5 KVKK | ✅ | C1, C2, C5, MODEL_CARD drift |
| 6 Deploy | ✅ | URL mismatch kanıtlandı |
| 7 Docs | ✅ | Scoring checklist |
| 8 Sentez | ✅ | Bu dosya |

---

*Review yapan: Cursor Agent (plan: Full Repo Review).*

---

## Fix Uygulaması — 2026-06-06 (Review Findings Fix Plan)

### Çözülen Critical maddeler

| ID | Fix |
|----|-----|
| C1 | `WholeFrameDetector` + HF detector when token set; never noop on uploads |
| C2 | Vision multipart upload → anonymizer in `prepare_image.go` |
| C3 | `mobile/eas.json` URL → `civiclens-backend-1rkq.onrender.com` |
| C4 | `OperatorView` parses bare Task/Report from review API |
| C5 | `KVKK-COMPLIANCE.md` team name + §5 checklist (imza hâlâ kullanıcıya bırakıldı) |

### Çözülen Important maddeler

| ID | Fix |
|----|-----|
| I1 | Web `FieldStaffView` — start + evidence upload |
| I2 | Web `ManagerView` — close task + audit log snippet |
| I3 | Mobile broken `ReportTab` kaldırıldı |
| I4 | Demo placeholder SVG (API-driven, no `/samples/` 404) |
| I5 | Mobile `needs_review` badge |
| I7 | `AnonymizationMeta` → honest privacy report |
| I8 | `MODEL_CARD.md` Wave 2 blur stratejisi |
| I9 | `backend/.env.example`, root `.env.example`, `render.yaml` HF template |
| I10 | `CreateFromReport` status guard + test |
| I11 | Street View wire in `app.go` + `prepareStreetView` |
| I13 | `DEMO.md` 3 dk script |
| I14 | Verification service Save error handling |

### Post-fix verification log

| Komut | Sonuç |
|-------|-------|
| `cd backend && go test ./... -count=1` | **PASS** |
| `cd backend && go vet ./...` | **PASS** |
| `cd web && npm run lint` | **PASS** |
| `cd web && npm run build` | **PASS** |
| `cd mobile && npx tsc --noEmit` | **PASS** |
| `curl …/civiclens-backend-1rkq.onrender.com/health` | **200** (deploy güncellemesi ayrı push gerektirir) |

### Kullanıcı aksiyonu

- `backend/.env` içine `HF_API_TOKEN=hf_…` koy (gitignore'da; commit etme)
- Render dashboard'a aynı token'ı ekle
- `KVKK-COMPLIANCE.md` §6 imzalarını jüri öncesi doldur
