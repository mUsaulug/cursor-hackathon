# CivicLens — Wave 2: Belediye Operasyon Platformu (Ürün Tasarımı)

> **Tarih:** 2026-06-06
> **Durum:** Tasarım — onay bekliyor (ikinci dalga)
> **Önceki dalga:** `docs/2026-06-06-civiclens-core-design.md` (CivicLens Core — tek görüntü analiz çekirdeği, tamamlandı)
> **Bu belgenin amacı:** Çekirdeği gerçek veriyle beslenen, satılabilir bir belediye operasyon platformuna çevirmek.

---

## 0. Tek cümle

> **CivicLens**, vatandaş ve belediye saha personelinin yüklediği kentsel problem görüntülerini analiz eden, KVKK-güvenli hale getiren, sınıflandırıp önceliklendiren, ilgili müdürlüğe **görev** olarak yönlendiren ve görev tamamlandığında **kanıt fotoğrafıyla doğrulayan** AI destekli belediye operasyon platformudur.

Birinci dalga "ne görüyorum"u çözdü. İkinci dalga "bu bildirim nasıl işe ve çözüme dönüşür"ü çözer.

---

## 1. Neden bu yön doğru (öneri analizi)

Kullanıcı önerisinin temel tezi doğru ve onaylıyorum: **gerçek veri = belediye kamera arşivi değil; gerçek veri = vatandaş + saha personeli üretimi.** Sistem kendi üretim akışından veri toplar. Bu hem gerçekçi hem satılabilir.

Sistemi bilen biri olarak önerinin üzerine koyduğum/değiştirdiğim kararlar:

| Öneri | Kararım | Gerekçe |
|-------|---------|---------|
| "Blur gerekmez" (Wave 1) | **Blur artık ZORUNLU** | Vatandaş fotoğrafı gerçek PII içerir (yüz/plaka). Avoidance-by-design sadece sentetik/altyapı demosu içindi. |
| Kafka opsiyonel/risk (Wave 1 §16) | **Kafka bu dalgada omurga** | Tek analiz değil, çok-adımlı görev yaşam döngüsü var; event-driven doğru model. |
| Redis dışlanmıştı (Wave 1 §3) | **Redis devrede** | Duplicate (geohash) + dashboard sayaç + idempotency için gerçek ihtiyaç. |
| RBAC atlanmıştı (Wave 1) | **RBAC çekirdek** | 4 rol var; masterfabric-go RBAC tam burada anlam kazanır. |
| `analysis` = bildirim | **`report` ≠ `analysis`** | Bir report'un before/after çok analizi olur. Pipeline aynen reuse. |
| Metin → kategori | **Metin yalnızca sinyal** | Karar deterministik kalmalı (Wave 1 ilkesi); metin güveni artırır, karar vermez. |
| Recheck (Street View/Mapillary) | **MVP'de personel/vatandaş tekrarı; Street View destekleyici** | Street View güncellik garantisi yok; ana recheck saha kanıtı. |
| Completion "AI doğrular" | **before/after deterministik diff + insan onayı** | LLM "çözüldü" demez; kural + human-in-the-loop. |

---

## 2. Wave 1'den ne devralınıyor (reuse analizi)

İkinci dalga sıfırdan başlamaz; mevcut hexagonal çekirdek doğrudan genişler.

| Mevcut bileşen | Wave 2'de rolü |
|----------------|----------------|
| `internal/domain/vision` (`AnalysisResult`, `Detection`, `BoundingBox`, `PrivacyReport`) | Aynen kalır; `AnalysisResult`'a `report_id` alanı eklenir |
| Karar zinciri (`application/vision` pipeline) | **Hiç değişmeden** her görselde çalışır (intake + completion) |
| `docs/rules/*.yaml` + stdlib loader | Pattern genişler: `department_routing.yaml`, `priority_factors.yaml` eklenir |
| `infrastructure/huggingface` DETR + `demo` precomputed | Reuse; yeni: **anonymization detector** (yüz/plaka bbox) adapter |
| `infrastructure/openrouter` | Reuse; yeni: metin sinyali için opsiyonel sınıflandırma (yine sadece sinyal) |
| `shared/imaging` (PixelateRegions) + `cmd/anonymize` | **Blur pipeline'ın hazır temeli** — ingest'e bağlanır |
| `infrastructure/streetview` | Recheck için destekleyici kaynak |
| `infrastructure/store` in-memory | Wave 2'de `report/task/evidence` için **Postgres** ile değişir; port aynı kalır |

> Sonuç: çekirdek korunur, üzerine **report / task / evidence / identity** bounded context'leri eklenir.

---

## 3. Roller (RBAC)

| Rol | Yapabildikleri |
|-----|----------------|
| `citizen` | Bildirim oluştur (foto + konum + açıklama), kendi bildirimlerini izle |
| `field_staff` | Bildirim oluştur, atanan görevleri gör, kanıt fotoğrafı yükle, tamamla |
| `operator` (operasyon görevlisi) | Bildirimleri incele, görev oluştur/ata, review onayı, duplicate birleştir |
| `manager` (yönetici) | Dashboard/rapor, SLA, müdürlük performansı (yazma yok, salt okuma + kapatma onayı) |

MVP: basit rol-token (header). Ürün: masterfabric-go JWT/RBAC. Her yazma işlemi **audit trail**'e düşer (kim, ne zaman, hangi karar) — KVKK ve hesap verebilirlik.

---

## 4. Yeni bounded context'ler

```
internal/
  domain/
    vision/        (mevcut — analiz)
    report/        (yeni — bildirim + yaşam döngüsü)
    task/          (yeni — görev + atama + SLA)
    evidence/      (yeni — tamamlama kanıtı + doğrulama)
    identity/      (yeni — kullanıcı/rol)
  application/
    intake/        (bildirim alma orkestrasyonu: anonymize -> analyze -> classify -> dedup -> route)
    tasking/       (görev oluşturma/atama/durum)
    verification/  (before/after completion diff)
    analytics/     (dashboard agregasyonları)
  infrastructure/
    anonymizer/    (yeni — yüz/plaka detector + blur, ingest'te)
    geocode/       (yeni — reverse geocoding, opsiyonel)
    persistence/postgres/   (report/task/evidence repo)
    events/kafka/  (yeni — yaşam döngüsü event bus)
    cache/redis/   (yeni — dedup geohash + dashboard cache + idempotency)
```

---

## 5. Çekirdek akışlar (durum makineleriyle)

### 5.1 Report yaşam döngüsü
```
created -> anonymized -> ai_analyzed -> dedup_checked -> waiting_for_review
   -> (operator) -> task_created | rejected | merged_into(existing)
```

### 5.2 Task yaşam döngüsü
```
created -> assigned -> started -> evidence_uploaded -> ai_verified
   -> (manager) -> completed | reopened
```

### 5.3 Intake akışı (en kritik)
```
[citizen/staff] foto+konum+açıklama
  -> Idempotent intake (client report_id)
  -> KVKK Anonymizer (yüz/plaka tespit -> IRREVERSIBLE blur -> raw discard)   [ZORUNLU]
  -> Vision pipeline (mevcut karar zinciri)
  -> Metin sinyali (açıklama -> kategori güven artışı, karar değil)
  -> Duplicate detection (geohash + tip + zaman penceresi)
  -> Department routing (YAML)
  -> Priority (çok faktörlü: tip + şikayet sayısı + konum riski)
  -> waiting_for_review
```

### 5.4 Completion verification
```
field_staff "after" foto yükler
  -> anonymize -> vision analyze
  -> diff(before, after): aynı tip/konumda problem hâlâ var mı?
       likely_resolved | still_present | needs_human
  -> manager kapatır / yeniden açar
```

### 5.5 Recheck (opsiyonel, MVP sonrası)
Süre sonra: saha personeli tekrar foto / yeni vatandaş bildirimi (birincil) + Street View (destekleyici, güncellik garantisiz).

---

## 6. Veri modeli (Wave 1 şemasına bağlı)

```jsonc
// Report — bildirim (kaynak veri)
{
  "report_id": "rep_001",
  "source_type": "citizen_mobile | staff_mobile | web",
  "reporter_role": "citizen | field_staff",
  "description": "Yolda büyük çukur var",
  "location": { "lat": 41.0082, "lng": 28.9784 },
  "address_context": "Kadıköy, ... (reverse geocode, opsiyonel)",
  "image_ref": "img_001",            // sadece ANONİMLEŞTİRİLMİŞ türev
  "analysis_id": "ana_001",          // vision pipeline çıktısı
  "problem_type": "road_damage",
  "priority": "high",
  "review_status": "needs_review",
  "assigned_department": "Fen İşleri Müdürlüğü",
  "duplicate_of": null,
  "status": "waiting_for_review",
  "created_at": "2026-06-06T10:30:00+03:00"
}

// AnalysisResult — MEVCUT şema + report_id (tek eklenti)
{ "analysis_id": "ana_001", "report_id": "rep_001", "model_mode": "live_hf",
  "detections": [ ... ], "kvkk_safe": true, "privacy": { "pii_strategy": "blur_applied", "anonymized": true } }

// Task
{ "task_id": "task_001", "report_id": "rep_001", "assigned_department": "Fen İşleri Müdürlüğü",
  "assigned_to": "field_team_01", "priority": "high", "status": "assigned", "sla": "48h" }

// CompletionEvidence
{ "evidence_id": "ev_001", "task_id": "task_001", "before_analysis_id": "ana_001",
  "after_analysis_id": "ana_900", "image_ref": "after_img_001",
  "ai_verification": "likely_resolved | still_present | needs_human", "manager_approval": "pending" }
```

> Kritik: `image_ref` her zaman **anonimleştirilmiş** türevi gösterir. Ham foto hiçbir zaman kalıcılaşmaz (`raw_image_stored=false` korunur).

---

## 7. Kritik mimari kararlar

1. **Modüler monolit** (ayrı mikroservis değil). masterfabric-go hexagonal yapısında bounded context'ler. Erken ürün + hackathon için doğru; sınırlar net olduğu için sonra bölünebilir.
2. **Persistence: PostgreSQL (+ PostGIS opsiyonel).** Report/task yaşam döngüsü in-memory ile yönetilemez. Port `vision/AnalysisStorePort` pattern'i ile repo arayüzleri; MVP'de in-memory adapter korunur, ürün Postgres.
3. **Kafka: yaşam döngüsü omurgası.** Eventler: `report.created/anonymized/ai_analyzed/dedup_checked`, `task.created/assigned/started/evidence_uploaded/ai_verified/completed/reopened`. Kural (Wave 1'den): **ilk HTTP yanıtı Kafka beklemez**, publish arka planda; hata → log + devam.
4. **Redis:** `report:hash:{sha}` (aynı görsel tekrar analizini engelle), `geo:recent:{geohash}` (dedup), `dashboard:summary:{district}` (sayaç cache), idempotency anahtarı. Hata → skip, isteği bloklamaz.
5. **KVKK blur ZORUNLU (en kritik değişiklik):** ingest'te otomatik yüz/plaka tespiti → `imaging.PixelateRegions` ile irreversible blur → ham buffer discard. Yeni `anonymizer` adapter (HF face/plate detector; tespit YALNIZCA anonimleştirme için, tanıma yok). `pii_strategy="blur_applied"`.
6. **Duplicate detection:** önce basit ve deterministik — `geohash(precision ~7) + problem_type + zaman penceresi`. Eşleşme → mevcut report'a `+1 şikayet`, priority artışı. Embedding benzerliği sonraki faz.
7. **Çok faktörlü priority:** `priority_factors.yaml` — tip ağırlığı + şikayet sayısı + konum riski (okul/anayol/hastane yakını; Overpass ile bağlam, opsiyonel). Deterministik skor → critical/high/medium/low.
8. **Department routing:** `department_routing.yaml` (normalized_object_type → müdürlük). Wave 1 rule-loader pattern'i.
9. **Completion verification deterministik:** before priority-yüksek `road_damage` → after aynı tip/konumda tespit yoksa/confidence düştüyse `likely_resolved`; aksi `still_present`; belirsiz `needs_human`. LLM değil.
10. **Reverse geocoding:** Nominatim (rate-limit, attribution) veya Google Geocoding; opsiyonel, sadece bağlam (mahalle/cadde). Konum kararını etkilemez.
11. **Idempotency & offline:** mobil offline olabilir → client tarafı `report_id` üretir, intake idempotent (Redis anahtarı). Saha personeli senaryosu için şart.
12. **OpenRouter:** yalnızca rapor prozası + (opsiyonel) metin sinyali özetleme; karar yolunda değil (Wave 1 ilkesi korunur).

---

## 8. API yüzeyi (yeni + mevcut)

```
# Intake
POST /api/v1/reports                      # foto+konum+açıklama -> Report (anonymize+analyze+dedup+route)
GET  /api/v1/reports                       # filtreli liste (rol bazlı)
GET  /api/v1/reports/{id}
POST /api/v1/reports/{id}/review           # operator: accept/reject/merge

# Tasking
POST /api/v1/tasks                         # report -> task (operator)
GET  /api/v1/tasks?assigned_to=...         # saha personeli görev listesi
POST /api/v1/tasks/{id}/assign
POST /api/v1/tasks/{id}/start
POST /api/v1/tasks/{id}/evidence           # after foto -> anonymize+analyze+verify
POST /api/v1/tasks/{id}/close              # manager: completed | reopened

# Analytics
GET  /api/v1/analytics/summary?district=   # mahalle/müdürlük/SLA/çözüm süresi

# Mevcut (Wave 1) — reuse
POST /api/v1/vision/analyze ... (intake ve evidence içeride bunu çağırır)
```

---

## 9. KVKK ürün gereksinimi (sertleşiyor)

- Gerçek vatandaş/personel fotoğrafı = PII. **Ingest'te otomatik yüz/plaka blur zorunlu**, inference/persist/transmit ÖNCESİ, in-memory, irreversible.
- Ham görüntü asla kalıcılaşmaz; sadece anonimleştirilmiş türev saklanır.
- Blur başarısızsa (detector hata) → güvenli taraf: görseli kabul etme/işleme veya tam-kare güçlü blur + `needs_human`. "Sessizce ham görsel saklama" YASAK.
- Audit trail + KVKK-COMPLIANCE güncellenir (artık `blur_applied`, avoidance-by-design değil).
- Konum verisi de kişisel veridir → erişim rol bazlı, ham GPS log'lanmaz (yalnızca işleme amaçlı).

---

## 10. Fazlı yol haritası (ikinci dalga)

| Faz | Kapsam | Çıktı |
|-----|--------|-------|
| W2-0 | Domain iskeleti: `report`/`task`/`evidence`/`identity` tipleri + `AnalysisResult.report_id` | derlenir, testli |
| W2-1 | **Anonymizer adapter (yüz/plaka blur, ingest)** + KVKK sertleştirme | blur pipeline ucu uca |
| W2-2 | Intake use case: anonymize→analyze→classify(metin sinyali)→dedup→route; `POST /reports` | bildirim → AI → kuyruk |
| W2-3 | Tasking: report→task, atama, durum makinesi; saha görev listesi | görev akışı |
| W2-4 | Completion verification: after foto + before/after diff | kanıtla doğrulama |
| W2-5 | Persistence (Postgres repo) + Redis dedup/cache | kalıcı yaşam döngüsü |
| W2-6 | Kafka event omurgası (lifecycle events) | event-driven koordinasyon |
| W2-7 | Analytics dashboard + manager raporları + SLA | yönetici görünümü |
| W2-8 | Web/mobil rol bazlı arayüzler (citizen/staff/operator/manager) | uçtan uca demo |
| W2-9 | Auth/RBAC (masterfabric-go) + audit trail | ürün güvenliği |

> Sıra riski en aza indirir: önce KVKK (blur) ve intake (en gerçek veri), sonra görev/doğrulama, en son altyapı (Postgres/Kafka) ve analytics. Her faz öncekini bozmadan ekler; Wave 1 vision çekirdeği hiç değişmez (yalnızca `report_id` eklenir).

---

## 11. Riskler / kapsam dışı (MVP)

- **Risk:** yüz/plaka detector doğruluğu — düşük güvende tam-kare blur + `needs_human` fallback.
- **Risk:** duplicate yanlış birleştirme — önce konservatif (dar zaman+mesafe), operator manuel merge.
- **Kapsam dışı (MVP):** embedding tabanlı duplicate, Street View/Mapillary recheck otomasyonu, çok-belediye (multi-tenant), gelişmiş SLA eskalasyon.
- **Korunan ilke:** KVKK ve priority kararları deterministik; LLM yalnızca prose/sinyal.

---

## 12. Sonuç

Bu dalga CivicLens'i "sokak görüntüsü analiz aracı"ndan **belediye operasyon platformuna** çevirir: bildirim → anonimleştir → analiz → önceliklendir → görev → kanıtla doğrula → raporla. Gerçek veri sistemin kendi akışından gelir; KVKK tasarımla değil **aktif blur ile** sağlanır; Kafka/Redis/RBAC artık gerçek ihtiyaç olduğu için devreye girer. Wave 1 çekirdeği zarar görmeden genişler.
```
