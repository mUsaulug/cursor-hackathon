# CivicLens Core — Architecture Design Document

> **Date:** 2026-06-06  
> **Status:** In Progress — iterative enrichment  
> **Hackathon:** Urban AI Solutions — 6-hour sprint  
> **Mandatory stack:** Go backend (masterfabric-go), Next.js web, Expo mobile, Hugging Face models, **Google Street View API** (external data — katı kural)  
> **Context repos analysed:** `evamcep/evam-saas-aihub` · `evamcep/evam-saas-insight`

---

## 0. One-Sentence Definition

> **CivicLens Core** — street-level görüntülerden gelen AI tespitlerini KVKK-güvenli, insan tarafından doğrulanabilir ve bakım önceliği atanmış kentsel aksiyon çıktılarına dönüştüren modüler Go çekirdeği.

---

## 1. Temel Fikir: "Core AI Karar Katmanı"

Bu proje bir "core model" değil, bir **"core AI karar katmanı"** kuruyor.

| Katman | Ne yapar? | Araç |
|--------|-----------|------|
| Görsel Algılama Motorları | Görseli alır, nesne tespiti yapar | Hugging Face modelleri |
| Açıklama / Reasoning | Rapor, policy reasoning, kalite kontrol | OpenRouter / LLM |
| **CivicLens Core** | Normalize eder, güvenli kılar, aksiyona dönüştürür | Go çekirdeği |
| Backend İskeleti | Clean/hexagonal backend altyapısı | masterfabric-go |

**Model sadece "ne gördüğünü" söyler. Core ise "bu belediye açısından ne anlama geliyor?" sorusunu cevaplar.**

Örnek dönüşüm:
```
HF ham çıktı → CivicLens Core → Belediye aksiyon çıktısı

{                               {
  "label": "traffic light",       "label": "traffic light",
  "score": 0.91,        →         "normalized_object_type": "traffic_signal",
  "box": {...}                    "confidence": 0.91,
}                                 "priority": "medium",
                                  "review_status": "auto_accepted",
                                  "kvkk_safe": true,
                                  "reason": "Trafik altyapısı yüksek güvenle tespit edildi."
                                }
```

---

## 2. Mimari Genel Görünüm

```
┌─────────────────────────────────────────────────────────────────────┐
│                        IMAGE SOURCE LAYER                           │
│   ┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐   │
│   │ Upload Image │  │ Sample Image │  │ Street View Adapter    │   │
│   └──────────────┘  └──────────────┘  └────────────────────────┘   │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────────┐
│                    PERCEPTION / MODEL LAYER                         │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│   │ HF DETR      │  │ HF Mask2     │  │ Road Damage YOLO Adapter │  │
│   │ Adapter      │  │ Former Adptr │  │ (rezzzq/yolo12s-rdd2022) │  │
│   └──────────────┘  └──────────────┘  └──────────────────────────┘  │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│   │ Traffic Sign │  │ Zero-shot    │  │ Precomputed Adapter      │  │
│   │ Adapter      │  │ Adapter      │  │ (demo/fallback)          │  │
│   └──────────────┘  └──────────────┘  └──────────────────────────┘  │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────────┐
│                       CIVICLENS CORE                                │
│                                                                     │
│  ┌───────────────────┐    ┌────────────────────┐                   │
│  │ Detection         │    │ Urban Object       │                   │
│  │ Normalizer        │────│ Ontology           │                   │
│  └───────────────────┘    └────────────────────┘                   │
│  ┌───────────────────┐    ┌────────────────────┐                   │
│  │ Privacy Guard     │    │ Confidence         │                   │
│  │ (KVKK)            │    │ Evaluator          │                   │
│  └───────────────────┘    └────────────────────┘                   │
│  ┌───────────────────┐    ┌────────────────────┐  ┌─────────────┐ │
│  │ Human Review      │    │ Priority Engine    │  │ Action      │ │
│  │ Router            │    │                    │  │ Report      │ │
│  └───────────────────┘    └────────────────────┘  │ Builder     │ │
│                                                    └─────────────┘ │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────────┐
│                      INTERFACE LAYER                                │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│   │ Next.js      │  │ Expo Field   │  │ API Consumers            │  │
│   │ Dashboard    │  │ View         │  │                          │  │
│   └──────────────┘  └──────────────┘  └──────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

**Bağımlılık yönü (hexagonal):**
```
HTTP Handler
  → Application Use Case
    → Domain Model / Domain Rule
      → Port Interface
        ← HF Adapter / OpenRouter Adapter / StreetView Adapter / Precomputed Adapter
```

HTTP handler asla doğrudan Hugging Face veya OpenRouter çağırmamalı.

---

## 3. masterfabric-go Uyumu

> **[ASSUMPTION — verify against delivered repo at 11:00]**  
> masterfabric-go saat 11:00'da teslim edilecek. Aşağıdaki klasör yapısı, README açıklamalarına dayanan tahmindir. İlk 30 dakika bu yapıyı gerçek repo ile karşılaştır; isimlendirme farklıysa §4'ü güncelle. Chi router da varsayım — mevcut stub `net/http ServeMux` kullanıyor.

### Beklenen Yapı (Teyit Edilecek)

masterfabric-go kendisini "enterprise-grade, multi-tenant, RBAC-driven SaaS backend platform" olarak tanımlar. Go + clean/hexagonal architecture ile inşa edilmiş. Beklenen:

- `internal/application` → Use case'ler
- `internal/domain` → Domain modelleri (dış bağımlılıksız)
- `internal/gateway` → Dış sistemlere port interfaces
- `internal/infrastructure` → Adapter implementasyonları
- `internal/shared` → Çapraz kesişim

Bu yapı vision bounded context eklemek için uygun — **teslim sonrası doğrulama şart**.

### Hackathon'da Kullanılacak Parçalar

| Kullanılacak | Kullanılmayacak |
|-------------|-----------------|
| Go server yapısı | Tam RBAC |
| Chi routing | Multi-tenant registration |
| Health endpoints | PostgreSQL persistence |
| Clean/hexagonal klasör ayrımı | Full Kafka consumer pipeline |
| Env config | Gateway policy engine |
| Structured logging | Audit log persistence (DB) |
| Postman/README disiplini | Redis Sentinel/Cluster |
| domain/application/infrastructure ayrımı | — |
| **Redis (cache katmanı — §20)** | — |
| **Kafka (event publisher — §20, opsiyonel)** | — |

---

## 4. vision Bounded Context — Dosya Yapısı

> **[ASSUMPTION — verify at 11:00]** Aşağıdaki dosya yolları masterfabric-go'nun beklenen klasör isimlerine dayanır. Gerçek repo teslim edildikten sonra ilk 30 dakikada karşılaştır ve farklı isimler varsa güncelle. Commit etmeden önce gerçek dizin yapısına uy.

```
internal/
  domain/
    vision/
      analysis.go           ← AnalysisResult, Detection, BoundingBox tipleri
      detection.go          ← Detection domain mantığı
      bbox.go               ← BoundingBox yardımcıları
      priority.go           ← Priority enum + scoring kuralları
      review_status.go      ← ReviewStatus enum
      privacy.go            ← PrivacyReport, blocked label listesi

  application/
    vision/
      ports.go              ← InferencePort, ReasonerPort interface'leri
      analyze_image_usecase.go
      normalize_detection.go
      privacy_guard.go
      priority_engine.go
      review_router.go
      action_report_builder.go
      model_router.go

  infrastructure/
    huggingface/
      client.go
      detr_adapter.go       ← facebook/detr-resnet-50
      mask2former_adapter.go
      road_damage_adapter.go ← rezzzq/yolo12s-rdd2022

    openrouter/
      client.go
      report_reasoner_adapter.go
      policy_reasoner_adapter.go

    demo/
      precomputed_adapter.go ← Fallback; gerçek model çıktıları
      sample_images.go

    streetview/
      streetview_adapter.go  ← Opsiyonel; API key riski

    http/
      handler/
        vision/
          handler.go
          dto.go
          mapper.go

  shared/
    config/
      ai_config.go
```

---

## 5. CivicLens Core İç Detayı

### 5.1 Detection Normalizer

Her AI adapter çıktısını tek `Detection` struct'ına dönüştürür.

```go
// HF DETR raw output — box koordinatları float döner, normalize edilir
{
  "label": "traffic light",
  "score": 0.91,
  "box": {"xmin": 120.4, "ymin": 79.8, "xmax": 190.1, "ymax": 220.3}
}

// YOLO raw output — class kodu ontology'e map edilir
{
  "class": "D40",          // D40 → road_damage (pothole)
  "confidence": 0.73,
  "bbox": [120.4, 79.8, 190.1, 220.3]  // [x_min, y_min, x_max, y_max]
}

// CivicLens normalized Detection — ID zorunlu, bbox float, model_id per-detection
type Detection struct {
    ID                   string      `json:"id"`                    // UUID — review flow için zorunlu
    Label                string      `json:"label"`
    NormalizedObjectType string      `json:"normalized_object_type"`
    Confidence           float64     `json:"confidence"`
    BBox                 BoundingBox `json:"bbox"`
    ModelID              string      `json:"model_id"`              // hangi adapter ürettiyse
    ReviewStatus         ReviewStatus `json:"review_status"`        // ayrı enum — priority değil
    Priority             Priority    `json:"priority"`              // high/medium/low/critical
    Reason               string      `json:"reason"`
}
```

> **Önemli:** `Priority` (high/medium/low/critical) ve `ReviewStatus` (auto_accepted/needs_review/rejected) ayrı kavramlardır ve ayrı enum olarak tutulur. `priority: "needs_review"` **yanlış** kullanımdır.

### 5.2 Urban Object Ontology

```
traffic_signal     ← traffic light, traffic sign, stop sign
road_damage        ← pothole, crack, D00, D10, D20, D40
sidewalk           ← curb, sidewalk, pedestrian path
street_furniture   ← bench, trash can, street light, bus stop
waste_asset        ← garbage bin, container, waste pile
unknown            ← tanımlanamayan veya düşük güvenli
```

Ontology mapping `docs/rules/ontology.yaml` dosyasında tutulur — model değiştikçe kod değil, kural güncellenir.

### 5.3 Privacy Guard (KVKK)

**KVKK stratejisi: PII-avoidance-by-design**

MVP'de Street View kullanılmaz. Yalnızca inanimate/cansız kentsel nesne tespiti yapılır. Böylece yüz/plaka işleme meselesi ortadan kalkar ve blur/anonymization koduna gerek kalmaz.

Jüriye anlatış: *"Sisteme kişisel veri girmiyor — algılama hedefi yalnızca altyapı nesneleri. PII riskini reddetme yoluyla değil, tasarımla engelliyoruz."*

```go
// COCO sınıf gerçeği:
// DETR (COCO eğitimli) → "person" VAR, "face" YOK, "license_plate" YOK, "pedestrian" YOK
// Yani sadece "person" ve araç sınıfları pratikte tetiklenecek

// MVP için gerçekçi blocked + hide listeleri:
var blockedLabels = map[string]bool{
    "person":     true,  // COCO'da var — filtrelenir
    "motorcycle": true,  // takip riski
    "bicycle":    true,  // sürücü takip riski
}

var defaultHideLabels = map[string]bool{
    "car":   true,  // plaka/takip riski
    "truck": true,
    "bus":   true,
}

// PIIStrategy, PrivacyReport'ta jüriye açıkça gösterilir
const PIIStrategyAvoidanceByDesign = "avoidance_by_design"
const PIIStrategyBlurApplied       = "blur_applied"       // Street View eklenirse
```

**Önemli:** KVKK ve privacy kararları LLM'e bırakılmaz. Deterministik kurallar.

**Street View eklenirse:** o zaman ve ancak o zaman blur/anonymization gerekir. `anonymized=false` ile "yapmadığımız şeyi yaptık demiyoruz" tutumunu koru; ama PII-avoidance stratejisiyle MVP'de bu satır hiç tetiklenmez.

### 5.4 Confidence Evaluator

```go
func EvaluateConfidence(score float64) ReviewStatus {
    switch {
    case score >= 0.80:
        return AutoAccepted
    case score >= 0.50:
        return NeedsReview
    default:
        return Rejected
    }
}
```

Gelişmiş sürüm: model bazlı ve nesne tipi bazlı threshold.
```yaml
# docs/rules/confidence_thresholds.yaml
road_damage:
  auto_accept: 0.85   # daha sıkı — yüksek yanlış pozitif riski
traffic_signal:
  auto_accept: 0.75   # daha esnek — altyapı tespiti daha stabil
unknown:
  auto_accept: 0.95
```

### 5.5 Priority Engine

`Priority` ve `ReviewStatus` ayrı kavram — karıştırma.

```go
// Priority: ne kadar acil bakım gerekiyor?
type Priority string
const (
    PriorityHigh     Priority = "high"     // road_damage
    PriorityMedium   Priority = "medium"   // traffic_signal, sidewalk
    PriorityLow      Priority = "low"      // street_furniture, waste_asset
    PriorityCritical Priority = "critical" // tehlike oluşturan hasar (gelecek)
)

// ReviewStatus: insan onayı gerekiyor mu?
type ReviewStatus string
const (
    ReviewAutoAccepted ReviewStatus = "auto_accepted"
    ReviewNeedsReview  ReviewStatus = "needs_review"
    ReviewRejected     ReviewStatus = "rejected"
)

var priorityMap = map[string]Priority{
    "road_damage":      PriorityHigh,
    "traffic_signal":   PriorityMedium,
    "sidewalk":         PriorityMedium,
    "street_furniture": PriorityLow,
    "waste_asset":      PriorityLow,
    "unknown":          PriorityLow,   // unknown → ayrıca needs_review verilir, ama priority düşük
}
```

Priority kuralları `docs/rules/priority_policy.yaml` dosyasında — Evidence-Grounded Decision Layer.

`unknown` nesne tipi → priority=low + review_status=needs_review. **Priority `"needs_review"` asla olamaz.**

### 5.6 Human Review Router

```
confidence >= 0.80 → auto_accepted
0.50 <= confidence < 0.80 → needs_review
confidence < 0.50 → rejected

+ unknown object type → needs_review
+ blocked label → filtered (KVKK, not queued)
```

Review queue'ya gidecek detections için:
```go
type ReviewRecord struct {
    DetectionID  string
    ReviewedBy   string
    Decision     string // accepted | rejected
    Note         string
}
```

### 5.7 Action Report Builder

```json
{
  "analysis_id": "ana_001",
  "source_type": "upload",
  "source_ref": "sample_istanbul_street_01",
  "location": {"lat": 41.0082, "lng": 28.9784},
  "model_id": "facebook/detr-resnet-50",
  "model_mode": "live_hf",
  "raw_image_stored": false,
  "anonymized": false,
  "kvkk_safe": true,
  "detections": [
    {
      "label": "traffic light",
      "normalized_object_type": "traffic_signal",
      "confidence": 0.91,
      "bbox": {"xmin": 120, "ymin": 80, "xmax": 190, "ymax": 220},
      "review_status": "auto_accepted",
      "priority": "medium",
      "reason": "Trafik altyapısı yüksek güvenle tespit edildi."
    }
  ],
  "created_at": "2026-06-06T10:30:00+03:00",
  "deletion_status": "raw_image_not_persisted"
}
```

---

## 6. Go Data Structures

```go
// AnalysisResult — web dashboard ve Expo aynı şemayı tüketir
type AnalysisResult struct {
    SchemaVersion  string      `json:"schema_version"`            // "1.0" — breaking change takibi
    AnalysisID     string      `json:"analysis_id"`               // UUID
    SourceType     string      `json:"source_type"`               // "upload" | "sample" | "streetview"
    SourceRef      string      `json:"source_ref"`
    Location       *Location   `json:"location,omitempty"`
    ModelID        string      `json:"model_id"`                  // birincil adapter model ID
    ModelMode      string      `json:"model_mode"`                // "live_hf" | "precomputed" | "road_damage"
    ImageWidth     int         `json:"image_width"`               // bbox overlay için zorunlu
    ImageHeight    int         `json:"image_height"`              // bbox overlay için zorunlu
    RawImageStored bool        `json:"raw_image_stored"`          // daima false
    Anonymized     bool        `json:"anonymized"`
    KVKKSafe       bool        `json:"kvkk_safe"`
    Privacy        PrivacyReport `json:"privacy"`                 // tek kaynak — AnalysisResult içinde
    Detections     []Detection `json:"detections"`
    CreatedAt      string      `json:"created_at"`
    DeletionStatus string      `json:"deletion_status"`
}

// Detection — her tespit kendi model kaynağını ve ID'sini bilir
type Detection struct {
    ID                   string       `json:"id"`                  // UUID — PATCH /reviews/{id} için zorunlu
    Label                string       `json:"label"`               // model'in döndürdüğü ham etiket
    NormalizedObjectType string       `json:"normalized_object_type"` // ontology'e map edilmiş
    Confidence           float64      `json:"confidence"`
    BBox                 BoundingBox  `json:"bbox"`
    ModelID              string       `json:"model_id"`            // bu detection'ı hangi adapter ürettiyse
    ReviewStatus         ReviewStatus `json:"review_status"`       // auto_accepted | needs_review | rejected
    Priority             Priority     `json:"priority"`            // high | medium | low | critical
    Reason               string       `json:"reason"`
}

// BoundingBox — HF float koordinatları korunur (int dönüşümü kayıp yaratır)
type BoundingBox struct {
    XMin float64 `json:"xmin"`
    YMin float64 `json:"ymin"`
    XMax float64 `json:"xmax"`
    YMax float64 `json:"ymax"`
}

type Location struct {
    Lat float64 `json:"lat"`
    Lng float64 `json:"lng"`
}

// PrivacyReport — AnalysisResult.Privacy içinde; bağımsız endpoint değil
type PrivacyReport struct {
    KVKKSafe       bool   `json:"kvkk_safe"`
    RawImageStored bool   `json:"raw_image_stored"` // daima false
    Anonymized     bool   `json:"anonymized"`
    DeletionStatus string `json:"deletion_status"`
    BlockedCount   int    `json:"blocked_count"`     // filtrenen detection sayısı
    PIIStrategy    string `json:"pii_strategy"`      // "avoidance_by_design" | "blur_applied"
}

// ReviewRecord — human review flow
type ReviewRecord struct {
    DetectionID  string `json:"detection_id"`  // Detection.ID ile eşleşmeli
    ReviewedBy   string `json:"reviewed_by"`
    Decision     string `json:"decision"`      // "accepted" | "rejected"
    Note         string `json:"note,omitempty"`
}
```

---

## 7. Hugging Face Model Stratejisi

### 7.1 Model Seçim Tablosu

| Model | HF ID | Rol | Hackathon Önceliği |
|-------|--------|-----|-------------------|
| DETR ResNet-50 | `facebook/detr-resnet-50` | Canlı baseline — trafik lambası, bank, vb. (COCO sınıfları) | **Zorunlu (MVP live)** |
| **Precomputed RDD-YOLO** | `rezzzq/yolo12s-road-damage-rdd2022` çıktıları | **Demo kahramanı** — D40=pothole, yüksek öncelik | **Zorunlu (precomputed)** |
| Mask2Former Mapillary | `facebook/mask2former-swin-large-mapillary-vistas-semantic` | Urban segmentation | Feature 5 (advanced) |
| Road Damage YOLO (live) | `rezzzq/yolo12s-road-damage-rdd2022` | HF Inference API olmayabilir — riskli | Live deneyin; yoksa precomputed |
| Pothole Detection | `peterhdd/pothole-detection-yolov8` | Tek sınıf pothole | Precomputed fallback |
| Grounding DINO | `IDEA-Research/grounding-dino-base` | Zero-shot / open vocab | Opsiyonel |
| OWLv2 | `google/owlv2-base-patch16-ensemble` | Zero-shot detection | Opsiyonel |

> **Kritik:** DETR **pothole/çatlak tespit edemez** (COCO'da bu sınıf yok). "High priority road_damage" demo tespiti precomputed RDD-YOLO çıktısından gelir — bunu jüriye açıkça söyle: *"Yol hasarı precomputed inference, trafik altyapısı live HF inference."*
>
> **HF cold start:** DETR bile ilk istekte HTTP 503 + `estimated_time` dönebilir. HF client'ta retry-with-wait (max 15s timeout) zorunlu. Demo öncesi DETR'i ısıt.

### 7.2 Dataset Dayanakları

| Dataset | Projedeki Rol | Lisans Notu |
|---------|--------------|-------------|
| Mapillary Vistas v2 | Street-level segmentation referansı | CC BY-NC-SA — araştırma dayanağı |
| RDD2022 (47.420 görsel) | Yol hasarı referansı; D40 = pothole | Belediye kullanım uyumlu |
| BDD100K | Driving/street scene referansı | README araştırma |
| Cityscapes | Urban semantic segmentation | Model card açıklaması |
| COCO | DETR genel baseline zemini | Ana gerekçe |

### 7.3 Model Router

```go
type ModelRouter struct {
    adapters map[string]InferencePort
}

// Routing tablosu
const (
    ModeLiveBaseline  = "live_hf"          // → HFDETRAdapter
    ModeRoadDamage    = "road_damage"       // → RoadDamageAdapter
    ModeSegmentation  = "segmentation"      // → Mask2FormerAdapter
    ModeOpenVocab     = "open_vocabulary"   // → GroundingDINOAdapter
    ModeDemo          = "precomputed"       // → PrecomputedAdapter (fallback)
)
```

---

## 8. OpenRouter Entegrasyonu

### Nerede Kullanılacak?

OpenRouter, HF'nin yerine geçmez. Yalnızca **reasoning / explanation** katmanında.

```go
type OpenRouterReasonerPort interface {
    GenerateActionExplanation(ctx context.Context, result AnalysisResult) (string, error)
    GeneratePrivacySummary(ctx context.Context, report PrivacyReport) (string, error)
    GenerateMaintenanceReport(ctx context.Context, result AnalysisResult) (*MaintenanceReport, error)
    // ValidateDecisionRationale KALDIRILDI — KVKK/priority kararları LLM'e bırakılamaz, deterministik olmalı
}
```

Bu port olmasa da sistem çalışmaya devam etmeli (degraded gracefully).

> **Demo path:** OpenRouter raporu demo öncesinde pre-generate et, canlı tıklamada anlık LLM çağrısı yapma. LLM gecikmesi demo sırasında görünmemeli.

### Doğru Kullanım Alanları

1. Detection result → kısa belediye bakım raporu
2. Priority reason açıklamasını doğal dile çevirme
3. KVKK checklist metni üretme
4. Human review kararını açıklama
5. Demo sonunda "analysis summary"

### Yanlış Kullanım

- OpenRouter VLM'ye görüntü gönderip asıl object detection yaptırmak
- KVKK/priority kararlarını LLM'e bırakmak (deterministik olmalı)

### AI Maintenance Report Endpoint

```
POST /api/v1/vision/report

Input:
{
  "analysis_id": "ana_001",
  "detections": [...],
  "privacy_report": {...},
  "priority_summary": {...}
}

Output:
{
  "summary": "Görüntüde trafik altyapısı ve düşük öncelikli şehir mobilyası tespit edilmiştir.",
  "recommended_action": "Trafik altyapısı envanter kaydına alınmalı; düşük güvenli tespitler saha personeli tarafından doğrulanmalıdır.",
  "risk_level": "medium",
  "kvkk_note": "Ham görüntü saklanmamış, kişisel veri riski taşıyan sınıflar filtrelenmiştir."
}
```

Fallback: OpenRouter fail → deterministic local report template.

---

## 9. Text AI Yöntemlerinin CivicLens Karşılıkları

### 9.1 RAG → Evidence-Grounded Decision Layer

```
Detected: pothole
Retrieved rules (from docs/rules/):
  road_damage → high priority
  confidence 0.40–0.80 → needs_review
  raw image must not be stored
Output:
  priority=high, review_status=needs_review, reason=<rule-based>
```

Kaynaklar: `docs/rules/priority_policy.yaml`, `docs/KVKK.md`, `docs/MODEL_CARD.md`

### 9.2 Structured Output → JSON Schema Zorumu

Her AI adapter sonucu tek `AnalysisResult` şemasına çevrilir. Frontend daima aynı şemayı görür.

### 9.3 Model Router → Adapter Registry

Ucuz model rutin işler için, detaylı model karmaşık işler için.

### 9.4 Guardrails → Privacy Guard + Label Policy

```
person → block
face → block
license_plate → block
car/truck/bus → default hide
low confidence → needs_review
unknown → needs_review
```

### 9.5 Confidence Calibration

Model bazlı ve nesne tipi bazlı threshold — tek genel eşik yok.

### 9.6 Human-in-the-Loop

needs_review kuyruğu → review_decision → döngü tamamlanır.

### 9.7 Precomputed Cache

HF live başarısız → precomputed adapter. Bu mock değil, önceden alınmış gerçek inference.

### 9.8 Evaluation Harness

```
eval/
  sample_images/
  expected_outputs.json
  smoke_test_cases.json
```

Kontrol noktaları:
- `/api/v1/vision/analyze` çalışıyor mu?
- JSON schema doğru mu?
- `raw_image_stored=false` mı?
- person/car filtrelenmiş mi?
- priority atanmış mı?
- low confidence review'a düşmüş mü?

---

## 10. API Tasarımı

### Zorunlu Endpointler

```
GET  /health/live
GET  /health/ready

POST /api/v1/vision/analyze        ← Core: görsel al → detection JSON dön
GET  /api/v1/vision/demo-results   ← Precomputed demo sonuçları
GET  /api/v1/vision/model-info     ← Hangi model? Live mi precomputed mu?
GET  /api/v1/vision/privacy-report ← KVKK durum raporu
GET  /api/v1/vision/summary        ← Son analizin özeti
```

### Yetişirse

```
GET   /api/v1/vision/reviews
PATCH /api/v1/vision/reviews/{detectionId}
POST  /api/v1/vision/report        ← OpenRouter maintenance raporu
```

---

## 11. Feature Sıralaması

Bu sırayı bozma.

| # | Feature | Zorunluluk | Detay |
|---|---------|-----------|-------|
| 0 | Core Image Analysis | **Zorunlu** | upload/sample → Go endpoint → HF/precomputed → detection JSON → dashboard |
| 1 | Privacy Guard | **Zorunlu** | blocked labels, `raw_image_stored=false`, kvkk_safe flag |
| 2 | Priority Engine | Zorunluya yakın | road_damage→high, traffic_signal→medium, unknown→needs_review |
| 3 | Human Review Queue | Zorunluya yakın | confidence threshold + auto_accepted/needs_review/rejected |
| 4 | Model Info / Transparency | Yapılmalı | live_hf mi, precomputed mu, hangi model, limitasyonlar |
| 5 | OpenRouter Report Reasoner | Opsiyonel | detection → kısa belediye raporu (core'u buna bağlama) |
| 6 | Street View Adapter | **Zorunlu (Hackathon Kuralı)** | `hackathon.mdc`: "External data → Google Street View API". API key `.env` içinde. Billing riski → sınırlı sample koordinat listesi kullan. PII-avoidance-by-design → Street View eklenir eklenmez blur pipeline aktif olmalı. |
| 7 | Mobile Field View | Opsiyonel | Son analiz + needs_review listesi |

---

## 12. Demo / Fallback Stratejisi

| Durum | Ne yapılacak? | Jüriye nasıl anlatılır? |
|-------|--------------|------------------------|
| HF çalışıyor | `model_mode=live_hf` | "Canlı HF inference sonucu" |
| HF yavaş | timeout → precomputed | "Dış servis riski için dokümante fallback" |
| HF hata | precomputed adapter | "Mock değil, önceden alınmış inference" |
| OpenRouter hata | local report template | "LLM rapor opsiyonel" |
| Street View yok | sample image | "Core sistem Street View'e bağımlı değil" |
| Mobile yetişmedi | web dashboard ana demo | "Mobile saha görünümü opsiyonel" |
| Street View yok (MVP) | PII-avoidance-by-design — `pii_strategy="avoidance_by_design"` | "Kişisel veri sisteme girmiyor; tasarım gereği engelledik, blur'a gerek yok" |
| DB yok | in-memory result | "MVP'de persistence ertelendi" |

**Kavram farkı (jüri sorarsa):**

| Terim | Açıklama |
|-------|----------|
| Mock data | Elle uydurulmuş sahte sonuç |
| Precomputed inference | Gerçek modelden önceden alınmış sonuç |
| Demo sample output | UI test için sabit örnek |
| Live inference | O anda model çağrısı |

---

## 13. Zaman Planı

### İlk 60 Dakika

| Zaman | İş | Commit |
|-------|----|--------|
| 0–10  | Yarışma checklist oku, masterfabric-go repo teslim alındı mı kontrol et | — |
| 10–30 | **masterfabric-go'yu oku** — klasör yapısı, router kaydı, env config, DI/wiring nasıl çalışıyor; §4'ü gerçek yapıya göre güncelle | — |
| 30–40 | vision bounded context klasörlerini aç; domain/vision type'larını yaz | **commit #1** |
| 40–50 | `/api/v1/vision/analyze` skeleton + precomputed adapter | — |
| 50–60 | Demo sample JSON + privacy guard iskelet | **commit #2** |

> **Uyarı:** masterfabric-go'yu okumadan kod yazmaya başlamak en yüksek zaman kaybı riskidir. İlk 30 dakikanın tamamı buna ayrılmalı.

### İlk 120 Dakika

| Zaman | İş | Commit |
|-------|----|--------|
| 60–80  | HFDETRAdapter dene (live) — 503 retry + 15s timeout ekle | — |
| 80–90  | privacy_guard.go tamamla (PII-avoidance-by-design) | **commit #3** |
| 90–110 | priority_engine.go + review_router.go + enum'lar | — |
| 110–120 | Web dashboard result card | **commit #4** |

### İlk 180 Dakika

| Zaman | İş | Commit |
|-------|----|--------|
| 120–140 | model-info + precomputed road-damage sample | **commit #5** |
| 140–155 | demo-results + summary endpointleri | — |
| 155–170 | README + docs/KVKK.md + docs/MODEL_CARD.md | **commit #6** |
| 170–180 | Smoke test (eval/ klasörü) | **commit #7** |

> **Toplu commit = disqualifikasyon riski.** Her ~20-30 dakikada anlamlı commit. Commit mesajları jüri tarafından değerlendirilir.

### Kalan Süre

- Render deploy
- Vercel deploy
- OpenRouter report endpoint
- Expo field view
- Street View adapter
- UI polish
- Demo rehearsal

---

## 14. Hackathon Skorlama Haritası

| Puan | Kriter | CivicLens Karşılığı |
|------|--------|---------------------|
| 30 | Teknik çalışabilirlik | masterfabric-go mimarisine uyum, hatasız çalışan demo |
| 25 | Doğruluk & güvenilirlik | HF detection + precomputed fallback + eval smoke tests |
| 20 | Kamu yararı | Kentsel bakım önceliği + insan onayı kuyruğu |
| 10 | AI adaptasyonu | Cursor agentic ruleset, CLI/SDK kullanım dokümantasyonu |
| 10 | KVKK & etik | Privacy Guard + KVKK-COMPLIANCE.md + deletion evidence |
| 5 | Sunum & dokümantasyon | README kalitesi, MODEL_CARD, bu mimari doc |

---

## 15. Kesinlikle Yapılacaklar

- [ ] vision bounded context (`internal/domain/vision`, `internal/application/vision`, `internal/infrastructure/huggingface`)
- [ ] HFDETRAdapter veya PrecomputedAdapter
- [ ] Privacy Guard (deterministik KVKK kuralları)
- [ ] Priority Engine (YAML rule-based)
- [ ] Review Status (auto_accepted / needs_review / rejected)
- [ ] AnalysisResult JSON (standart schema)
- [ ] Web dashboard result card
- [ ] README + docs/KVKK.md + docs/MODEL_CARD.md

## 16. Projeyi Riske Sokacaklar

- Tam DB persistence (PostgreSQL tam entegrasyon)
- Kafka event flow
- Full auth/RBAC
- Gerçek-time map tracking
- Plaka/yüz blur iddiası ama kodsuz
- Tek başına YOLO road damage canlı inference'a bağımlılık
- OpenRouter VLM ile bbox tespiti yaptırmak

---

## 17. evam-saas-aihub'dan Alınan Dersler

> **Repo:** `evamcep/evam-saas-aihub` — Java/Spring Boot multi-tenant SaaS AI platformu

### Uygulanabilir Patternler

| aihub Paterni | CivicLens Karşılığı |
|--------------|---------------------|
| `LlmFeatureConfig` — feature bazlı model routing | `ModelRouter` — urban detection mode bazlı adapter seçimi |
| MCP tool integration — LLM'in harici araçları çağırması | Urban decision tools (priority rule engine, KVKK checker) |
| Streaming SSE response (Flux) | Go'da chunked response / SSE (gelecek feature) |
| Conversation history + branching | Analysis history + human review flow |
| Hybrid search (exact + semantic) | Urban RAG: regulation lookup + precedent similarity |
| Usage & cost tracking per feature | Model mode tracking (live_hf vs precomputed) |
| Tool execution audit trail | Detection decision audit (hangi model, hangi kural, neden) |
| Multi-tenancy via tenant context | Multi-belediye desteği (gelecek) |
| Interrupt & cancellation | Long-running inference timeout + precomputed fallback |

### Kritik Fark

aihub reactive/async Java kullanır (Webflux). CivicLens ise Go ile senkron HTTP + goroutine eşzamanlılığı kullanacak. Go'da interface'ler üzerinden hexagonal ayrım aynı felsefeyi uygular.

---

## 18. evam-saas-insight'tan Alınan Dersler

> **Repo:** `evamcep/evam-saas-insight` — Java/Spring Boot event-driven analytics platformu

### Uygulanabilir Patternler

| insight Paterni | CivicLens Karşılığı |
|----------------|---------------------|
| STG → DWH → DM 3 katman | Raw detection → Normalized detection → Priority report |
| JSONB staging + schema evolution | Detection metadata JSONB (model version, bbox, confidence histogram) |
| SCD Type 2 dimensions | Urban asset versioning (bakım geçmişi zaman serisi) |
| Fingerprint-based dedup | Detection dedup: (geohash, timestamp, model_version) |
| Continuous Aggregates (CAGG) | Pre-computed: "son 30 gün ilçe bazlı çukur sayısı" |
| Distributed ETL lock | Çakışan detection merge'i (2 kamera aynı çukuru görüyor) |
| Dynamic SQL query builder | Civic dashboard filter builder (no SQL injection) |
| Redis 1000-row cache boundary | "İlçe bazlı öncelikli tespit" → cache, "tüm asset listesi" → pass-through |
| 3-phase ETL scheduler | Phase 1: Reference dimensions, Phase 2: Fact tables, Phase 3: Aggregates |
| Dual datasource (read DWH, write config) | Okuma: detection results, yazma: review decisions |

### Hackathon MVP İçin Basitleştirilmiş Versiyon

insight full analytics pipeline kurmak hackathon için fazla. Ama key pattern alınır:

```
Kafka / HF Model Output
  → STG: in-memory slice ([]AnalysisResult)    ← insight'ın stg_schema karşılığı
    → Normalize: detection mapper               ← insight'ın ETL transformer karşılığı
      → DWH: processed results JSON            ← insight'ın DWH karşılığı (in-memory MVP)
        → API: /summary endpoint               ← insight'ın GraphQL API karşılığı
```

---

## 19. Sonuç

**CivicLens Core, bu projenin ana fikri ve ürünleşebilir temelidir.**

Bu core katman sayesinde proje tek bir çukur tespit demosu olmaktan çıkıyor; Hugging Face modellerini, OpenRouter reasoning'i, KVKK kurallarını, human review mantığını ve belediye bakım önceliğini bir araya getiren genişletilebilir bir kentsel AI karar platformu haline geliyor.

Temel prensipler:

1. **Core küçük, deterministik ve test edilebilir kalır.**
2. **Her yeni özellik core'a adapter veya rule olarak bağlanır.**
3. **KVKK ve priority kararları LLM'e bırakılmaz.**
4. **Fallback her katmanda var; canlı bağımlılık minimum.**
5. **masterfabric-go mimarisine zarar verilmez; sadece vision bounded context eklenir.**

---

## 20. Redis + Kafka Integration (Speed + Extensibility Layer)

> **⚠️ go.mod CONSTRAINT (`backend.mdc` katı kural):** `backend/go.mod` şu an yalnızca `go 1.22` içeriyor — sıfır üçüncü-taraf paket. `backend.mdc` kuralı: *"Standard library only unless a package already exists in go.mod."*  
> Bu, `go-redis` ve `kafka-go` kullanmadan **önce** bunları `go.mod`'a eklemeniz gerektiği anlamına gelir.  
> **Uygulama adımı:** `cd backend && go get github.com/redis/go-redis/v9 && go get github.com/segmentio/kafka-go` — bunu Redis/Kafka adapter dosyaları yazılmadan önce çalıştır ve commit et.  
> **Alternatif (paket eklemek istemezsen):** Upstash Redis'in REST API'sini stdlib `net/http` ile çağırabilirsin — SDK gerekmez; `hackathon.mdc`'de Go için "standard library only" zorunluluğunu karşılar.

> **Authoritative flow decision:** Redis IS in the response path (cache check **before** the HF call). Kafka is **NOT** in the response path (async fan-out **after** the HTTP response is built). Both can fail without breaking the HTTP response.
>
> **Grounded in:** `aihub-central-api` Redis pub/sub (`journey:interrupt:broadcast`), `aihub-core/kafka-configuration` (`KafkaTopicProperty`, `KafkaProducerProperty`), `insight-api` `RedisCacheAdapter`, `insight-data-collector` `RawEventsConsumerAdapter` (`rawEvents` → STG).

### 20.1 Architecture — Where Redis and Kafka Sit

```
                         POST /api/v1/vision/analyze  (image bytes)
                                        │
                                        ▼
                          ┌──────────────────────────┐
                          │  Vision HTTP Handler       │
                          └─────────────┬─────────────┘
                                        │  image_hash = sha256(bytes)
                                        ▼
                          ┌──────────────────────────┐
                          │  AnalyzeImageUseCase       │
                          └─────────────┬─────────────┘
                                        │
            ┌───────────────────────────┼─────────────────── RESPONSE PATH (sync) ──────┐
            │                           ▼                                                │
            │  (1) ┌────────────────────────────────────┐  vision:analysis:{hash}:{mdl} │
            │      │  AnalysisCachePort.Get()  ── REDIS ─┼──► HIT ──┐  TTL 30m–2h        │
            │      └────────────────┬───────────────────┘          │                    │
            │                       │ MISS / Redis down (skip)      │                    │
            │                       ▼                               │                    │
            │  (2) ┌────────────────────────────────────┐          │                    │
            │      │  InferencePort  (HF DETR / RDD /    │          │                    │
            │      │  Precomputed)  — ONLY on cache miss │          │                    │
            │      └────────────────┬───────────────────┘          │                    │
            │                       ▼                               │                    │
            │  (3) ┌────────────────────────────────────┐          │                    │
            │      │  CivicLens Core:                    │          │                    │
            │      │  normalize → privacy → confidence → │          │                    │
            │      │  priority → review                  │          │                    │
            │      └────────────────┬───────────────────┘          │                    │
            │                       ▼                               │                    │
            │  (4) ┌────────────────────────────────────┐          │                    │
            │      │  AnalysisCachePort.Set()  ── REDIS ─┼◄─────────┘                    │
            │      └────────────────┬───────────────────┘  (write-through on miss)       │
            │                       ▼                                                    │
            │  (5)         HTTP 200  AnalysisResult  ──────────────────────────────────► │ client
            └───────────────────────┬─────────────────────────────────────────────────┘
                                    │   (response already returned — caller not blocked)
                                    ▼
            ┌───────────────────── ASYNC PATH (fire-and-forget goroutine) ──────────────┐
            │  (6) ┌────────────────────────────────────┐                                │
            │      │  EventPublisherPort.Publish() ─ KAFKA                               │
            │      │  topic civiclens.vision.events  key = analysis_id                   │
            │      │  events: analysis.completed / review.required /                     │
            │      │          priority.high_detected / privacy.filtered                  │
            │      └────────────────┬───────────────────┘  Kafka down → log + continue   │
            └───────────────────────┼─────────────────────────────────────────────────┘
                                    ▼
            ┌────────────────┐  ┌────────────────┐  ┌──────────────────────────────┐
            │ Analytics       │  │ Review Queue   │  │ Notification / Dashboard      │
            │ consumer → STG  │  │ consumer        │  │ live feed (future)            │
            └────────────────┘  └────────────────┘  └──────────────────────────────┘
```

**One rule to remember:** the dotted ASYNC box runs *after* step (5). If Kafka is dead, the user already has their `AnalysisResult`.

### 20.2 Port Interfaces (hexagonal — `internal/application/vision/ports.go`)

```go
package vision

import (
	"context"
	"time"
)

// AnalysisCachePort — Redis is in the response path. The use case calls Get()
// before inference and Set() after Core processing. Implementations MUST be
// non-fatal: on any Redis error return (nil,false,nil) for Get and nil for Set
// so the use case continues without the cache.
type AnalysisCachePort interface {
	// Get returns (result, true, nil) on hit; (nil, false, nil) on miss OR on
	// Redis failure (graceful degradation — never blocks the request).
	Get(ctx context.Context, key string) (*AnalysisResult, bool, error)

	// Set write-through caches the result. Errors are swallowed/logged by the
	// adapter and returned as nil so callers never branch on cache-write failure.
	Set(ctx context.Context, key string, result *AnalysisResult, ttl time.Duration) error
}

// JobStatusPort — only used in Mode 2 (async job). Backed by the same Redis.
type JobStatusPort interface {
	SetStatus(ctx context.Context, analysisID, status string, ttl time.Duration) error
	GetStatus(ctx context.Context, analysisID string) (string, bool, error)
}

// DomainEvent — the envelope every Kafka message uses. Payload carries ONLY
// metadata (never image bytes).
type DomainEvent struct {
	EventID     string         `json:"event_id"`     // UUID per event
	EventType   string         `json:"event_type"`   // vision.analysis.completed, ...
	AggregateID string         `json:"aggregate_id"` // analysis_id — also the Kafka partition key
	OccurredAt  time.Time      `json:"occurred_at"`  // RFC3339
	SchemaVer   string         `json:"schema_version"`
	Payload     map[string]any `json:"payload"`
}

// EventPublisherPort — Kafka is NOT in the response path. Publish is invoked
// from a detached goroutine after the HTTP response is built. Implementations
// log-and-swallow transport errors; a returned error is informational only.
type EventPublisherPort interface {
	Publish(ctx context.Context, event DomainEvent) error
}

// Event type constants — topic: civiclens.vision.events
const (
	EventAnalysisCompleted = "vision.analysis.completed"
	EventReviewRequired    = "vision.review.required"
	EventPriorityHigh      = "vision.priority.high_detected"
	EventPrivacyFiltered   = "vision.privacy.filtered"

	VisionEventsTopic = "civiclens.vision.events"
	EventSchemaVer    = "1.0"
)
```

### 20.3 Redis Adapter — `internal/infrastructure/cache/redis/analysis_cache.go`

Mirrors `insight-api`'s `RedisCacheAdapter`: a `redisAvailable` flag, connection-failure → mark-unavailable + skip, sliding TTL on read. Key composition includes `model_id` + `policy_version` so a model swap or a privacy-policy bump never serves a stale decision.

```go
package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"yourorg/internal/application/vision"
)

// TTL constants — tune per use case (design §Redis use cases).
const (
	TTLAnalysisCache = 30 * time.Minute // vision:analysis:{hash}:{model}:{policy}
	TTLJobStatus     = 30 * time.Minute // vision:job:{analysis_id}:status
	TTLDemoResult    = 24 * time.Hour   // vision:demo:{demo_id}
	TTLModelInfo     = 24 * time.Hour   // vision:model:{model_id}:info
)

// CacheKey builds the response-path cache key. Same image + same model +
// same policy version = same decision = cache hit. Any change invalidates.
func CacheKey(imageBytes []byte, modelID, policyVersion string) string {
	sum := sha256.Sum256(imageBytes)
	return fmt.Sprintf("vision:analysis:%s:%s:%s",
		hex.EncodeToString(sum[:]), modelID, policyVersion)
}

type AnalysisCacheAdapter struct {
	rdb       *redis.Client
	available atomic.Bool // graceful-degradation flag (insight pattern)
	log       *slog.Logger
}

func NewAnalysisCacheAdapter(rdb *redis.Client, log *slog.Logger) *AnalysisCacheAdapter {
	a := &AnalysisCacheAdapter{rdb: rdb, log: log}
	a.available.Store(true)
	return a
}

func (a *AnalysisCacheAdapter) Get(ctx context.Context, key string) (*vision.AnalysisResult, bool, error) {
	if !a.available.Load() {
		return nil, false, nil // Redis known-down → skip, never block request
	}
	raw, err := a.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil // clean MISS
	}
	if err != nil {
		a.available.Store(false) // connection failure → degrade (health check re-arms)
		a.log.Warn("redis get failed, cache disabled", "key", key, "err", err)
		return nil, false, nil
	}
	var res vision.AnalysisResult
	if err := json.Unmarshal(raw, &res); err != nil {
		a.log.Warn("redis cache corrupt, treating as miss", "key", key, "err", err)
		return nil, false, nil
	}
	a.rdb.Expire(ctx, key, TTLAnalysisCache) // sliding TTL (insight pattern)
	return &res, true, nil
}

func (a *AnalysisCacheAdapter) Set(ctx context.Context, key string, res *vision.AnalysisResult, ttl time.Duration) error {
	if !a.available.Load() {
		return nil
	}
	b, err := json.Marshal(res)
	if err != nil {
		a.log.Warn("redis cache marshal failed", "err", err)
		return nil // never propagate cache errors to the response path
	}
	if err := a.rdb.Set(ctx, key, b, ttl).Err(); err != nil {
		a.available.Store(false)
		a.log.Warn("redis set failed, cache disabled", "key", key, "err", err)
	}
	return nil
}

// HealthPing re-arms availability — wire to a 30s ticker (insight @Scheduled).
func (a *AnalysisCacheAdapter) HealthPing(ctx context.Context) {
	if err := a.rdb.Ping(ctx).Err(); err == nil {
		if a.available.CompareAndSwap(false, true) {
			a.log.Info("redis connection restored")
		}
	}
}
```

### 20.4 Kafka Publisher — `internal/infrastructure/events/kafka/publisher.go`

The publisher builds the `vision.analysis.completed` event from an `AnalysisResult` carrying **only counts and metadata** — no bbox-free guarantee needed because we never serialize the image. Async dispatch is a detached goroutine with its own short context timeout so it cannot outlive or block the request.

```go
package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"yourorg/internal/application/vision"
)

type Publisher struct {
	w   *kafka.Writer // RequireOne acks, like aihub KafkaProducerProperty acks="1"
	log *slog.Logger
}

func NewPublisher(brokers []string, log *slog.Logger) *Publisher {
	return &Publisher{
		w: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        vision.VisionEventsTopic,
			Balancer:     &kafka.Hash{}, // partition by Key = analysis_id (ordering per analysis)
			RequiredAcks: kafka.RequireOne,
			BatchTimeout: 300 * time.Millisecond, // mirrors aihub linger.ms=300
		},
		log: log,
	}
}

func (p *Publisher) Publish(ctx context.Context, e vision.DomainEvent) error {
	body, _ := json.Marshal(e)
	err := p.w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(e.AggregateID), // analysis_id = partition key
		Value: body,
		Time:  e.OccurredAt,
	})
	if err != nil {
		p.log.Warn("kafka publish failed (swallowed)", "type", e.EventType, "id", e.AggregateID, "err", err)
	}
	return err // informational only — caller already returned HTTP 200
}

// CompletedEvent builds the completed event. NOTE: NO image bytes, NO base64 —
// only metadata, counts, and the analysis_id needed to re-fetch the full result.
func CompletedEvent(r *vision.AnalysisResult) vision.DomainEvent {
	highCount, reviewCount := 0, 0
	for _, d := range r.Detections {
		if d.Priority == vision.PriorityHigh || d.Priority == vision.PriorityCritical {
			highCount++
		}
		if d.ReviewStatus == vision.ReviewNeedsReview {
			reviewCount++
		}
	}
	return vision.DomainEvent{
		EventID:     uuid.NewString(),
		EventType:   vision.EventAnalysisCompleted,
		AggregateID: r.AnalysisID,
		OccurredAt:  time.Now().UTC(),
		SchemaVer:   vision.EventSchemaVer,
		Payload: map[string]any{
			"model_id":         r.ModelID,
			"model_mode":       r.ModelMode,
			"detection_count":  len(r.Detections),
			"high_priority":    highCount,
			"needs_review":     reviewCount,
			"blocked_count":    r.Privacy.BlockedCount,
			"kvkk_safe":        r.KVKKSafe,
			"raw_image_stored": r.RawImageStored, // always false — no bytes anywhere
			"source_type":      r.SourceType,
			// location intentionally coarse / omitted — see "No PII" rule
		},
	}
}
```

**Async dispatch in the use case (after the response is built):**

```go
// In AnalyzeImageUseCase.Execute — AFTER result is ready, BEFORE returning.
func (uc *AnalyzeImageUseCase) emit(result *vision.AnalysisResult) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = uc.events.Publish(ctx, kafka.CompletedEvent(result))
		if hasHighPriority(result) {
			_ = uc.events.Publish(ctx, kafka.HighPriorityEvent(result))
		}
		if needsReview(result) {
			_ = uc.events.Publish(ctx, kafka.ReviewRequiredEvent(result))
		}
	}()
}
```

### 20.5 evam-saas-aihub Patterns Applied

aihub uses **two distinct Redis/Kafka idioms**; CivicLens borrows the exact shapes:

| aihub mechanism | Concrete aihub code | CivicLens mapping |
|---|---|---|
| **Redis pub/sub for control signals** | `InterruptBroadcastAdapter` publishes JSON to channel `journey:interrupt:broadcast` via `stringRedisTemplate.convertAndSend(...)`, **failures logged & swallowed** ("best-effort") | CivicLens uses Redis as a **key/value cache** (response path), not pub/sub for MVP. The *pattern* we copy is the **best-effort, swallow-and-log** discipline — see `AnalysisCacheAdapter.Set` returning `nil` on every failure. If we later add live-dashboard fan-out without Kafka, we replicate the `convertAndSend` channel idiom with a `vision:events:broadcast` channel. |
| **`RedisMessageListenerContainer` + `ChannelTopic`** subscriber wiring | `InterruptBroadcastConfig` registers a container per channel | Our `HealthPing`/ticker wiring is the Go equivalent of aihub's `@Scheduled` re-arm; a future SSE dashboard would add one Go subscriber goroutine per channel. |
| **Kafka producer tuning** | `KafkaProducerProperty`: `acks="1"`, `lingerMs=300`, `batchSize=16384`, `StringSerializer` key / `JsonSerializer` value | `kafka.Writer`: `RequiredAcks: RequireOne`, `BatchTimeout: 300ms`, key = `[]byte(analysis_id)` (string), JSON value. **Same throughput/ordering trade-off.** |
| **Kafka topic config structure** | `KafkaTopicProperty`: `name`, `group`, `partition=32`, `replication`, `retentionMs`, `autoOffsetReset=earliest`, `enableAutoCommit=false` | Our `civiclens.vision.events` config object carries the same fields. For the hackathon: `partition=1`, `replication=1`, `retentionMs=86400000` (24h). Consumers use manual commit (`enableAutoCommit=false`) — see §20.6. |
| **`STANDALONE` Redis** | central-api `redis.type: STANDALONE, url: localhost:6379` | Hackathon Redis = single `localhost:6379` standalone. No Sentinel/Cluster. |

**Key insight from aihub:** interrupt broadcasting is *explicitly best-effort* ("Failures are logged and swallowed"). CivicLens applies that same posture to **both** Redis cache writes and Kafka publishes — neither is allowed to fail the request.

### 20.6 evam-saas-insight Patterns Applied

insight's spine is **`rawEvents` Kafka topic → consumer → batch INSERT into `stg_*` tables**. This is exactly the shape of CivicLens's analytics fan-out.

| insight mechanism | Concrete insight code | CivicLens mapping |
|---|---|---|
| **`rawEvents` → STG consumer** | `RawEventsConsumerAdapter` `@KafkaListener(topics="rawEvents")` consumes a **batch** of records, transforms each (1:N journey explosion), then `stgBatchInsertService.batchInsert(...)` | A `VisionEventsConsumer` subscribes to `civiclens.vision.events`, and for each `vision.analysis.completed` writes one analytics row (`stg_vision_analysis`: analysis_id, model_id, detection_count, high_priority, needs_review, kvkk_safe, occurred_at). **No image — metadata only**, which is already guaranteed by §20.4. |
| **Manual ack, no-ack-on-failure = at-least-once** | `acknowledgment.acknowledge()` only after a successful batch insert; on exception it logs and **does NOT ack** so Kafka redelivers | CivicLens consumer commits offset only after the analytics write succeeds. A crashed analytics DB just replays events — the user-facing `/analyze` path is untouched. |
| **STG → DWH → DM layering** | raw_event → `stg_digital_studio_event` → aggregates | `vision.analysis.completed` → `stg_vision_analysis` → DWH (`fact_detection`) → DM/CAGG ("son 30 gün ilçe bazlı yüksek-öncelik tespit sayısı"). The `/api/v1/vision/summary` endpoint reads the DM layer. |
| **Redis 1000-row cache boundary / `insight:dwh:*` keys** | `RedisCacheAdapter` caches bounded query results, prefix-namespaced, with sliding TTL + `redisAvailable` flag | Our cache keys are namespaced `vision:analysis:*`, `vision:demo:*`, `vision:model:*` (same prefix discipline) with the same `available` flag + health re-arm. |
| **Batch consumption tuning** | `max-pool-records: 500`, `concurrency: 16` on `rawEvents` | Hackathon: `maxPoolRecords` small, `concurrency: 1`. The analytics consumer is **out of scope for the live demo** but the topic + a stub consumer prove extensibility. |

**Mapping in one line:** insight's `rawEvents → stg` becomes CivicLens's `vision.analysis.completed → stg_vision_analysis`, with the same manual-ack at-least-once contract, feeding the `/summary` endpoint via the same STG→DWH→DM idea (in-memory for MVP).

### 20.7 Hackathon Implementation Priority

| Tier | Work | Time | Why |
|---|---|---|---|
| **MVP (do first, 30–45 min)** | 1. `AnalysisCachePort` + `AnalysisCacheAdapter` (go-redis, standalone `localhost:6379`). 2. `CacheKey(sha256, model_id, policy_version)`. 3. Wire `Get` before inference, `Set` after Core in `AnalyzeImageUseCase`. 4. **Demo-result cache** `vision:demo:{id}` (TTL 24h) so the demo path is instant. 5. Graceful-degradation flag + 30s health ping. | 30–45m | Redis is in the response path and makes the demo visibly fast (cache HIT on repeat). Visible, low-risk, high-value. |
| **MVP Kafka (only if cache is solid)** | 6. `EventPublisherPort` + `kafka.Writer` publisher. 7. `CompletedEvent` builder (metadata only). 8. Detached-goroutine `emit()` after response. **No consumer needed** — publishing alone demonstrates the event-driven seam. | +15m | Proves extensibility ("every new feature attaches as an event consumer") without consumer complexity. Safe because it's off the response path. |
| **Advanced (Mode 2, if time remains)** | 9. `JobStatusPort` (`vision:job:{id}:status`, TTL 30m). 10. `POST /analyze` returns `analysis_id` immediately + publishes a `vision.job.queued` event. 11. Worker goroutine/consumer runs inference, updates Redis status. 12. `GET /analyze/{id}` polling endpoint reads Redis status. 13. Stub `VisionEventsConsumer` → in-memory `stg_vision_analysis` slice feeding `/summary` (insight pattern). | 60–90m | True async job mode + analytics fan-out. High value for scoring but only after MVP is rock-solid. |
| **Skip entirely** | Redis Sentinel/Cluster; Kafka SASL/SSL; multi-partition/replication; Kafka in the response path; pub/sub live-dashboard fan-out; DWH/DM persistence in a real DB; consumer auto-scaling. | — | All listed in §16 as risk. Keep Kafka partition=1, replication=1, retention 24h, single standalone Redis. |

**Demo narration:** *"İlk analiz canlı HF inference — ikinci kez aynı görsel Redis cache'ten anında dönüyor. Sonuç döndükten sonra Kafka'ya `vision.analysis.completed` event'i async gidiyor; analytics ve review consumer'ları buna abone. Redis veya Kafka düşse bile `/analyze` cevabı dönmeye devam ediyor."*

### 20.8 Key Design Rules (binding)

1. **Redis = speed.** Cache (image-hash decision reuse) and state (job status). In the response path, but every Redis call is guarded by the `available` flag and degrades to a clean MISS.
2. **Kafka = extensibility.** Async, event-driven, **never in the response path**. Published from a detached goroutine *after* the HTTP response is built. New features (analytics, notifications, review queue) attach as consumers — zero changes to `/analyze`.
3. **Graceful degradation, both sides.** Redis down → skip cache, run inference, still respond. Kafka down → log and continue, response already returned. The HTTP response **always** returns. (Copied verbatim from aihub's "best-effort, swallow-and-log" interrupt broadcasting.)
4. **No PII in Kafka payloads.** Never serialize image bytes or base64. Payloads carry `analysis_id` + counts + model/privacy metadata only. `raw_image_stored=false` is asserted in every event. Location is coarse or omitted. KVKK/privacy decisions stay deterministic and inside Core — events only *report* them.
5. **Cache key correctness.** `vision:analysis:{sha256}:{model_id}:{policy_version}` — a model swap or privacy-policy bump changes the key, so a stale decision can never be served after a rule change.
6. **At-least-once on the consumer side.** Any future consumer (per insight) commits the offset only after its write succeeds; replay is safe because events are idempotent on `event_id` / `analysis_id`.

---

## 21. Ürün Vizyonu & Go-to-Market

### Belediye Ne Satın Alır?

Belediye "AI modeli" satın almaz. Satın aldığı şeyler:

- Daha hızlı saha tespiti
- Daha az manuel kontrol
- Daha düzenli bakım planı
- Vatandaş şikâyetlerine daha hızlı cevap
- Harita tabanlı sorun envanteri
- Denetlenebilir rapor
- KVKK uyumlu veri akışı
- Akıllı şehir hedeflerine uyum

**Satış cümlesi:** *"CivicLens, belediyelerin sahadan veya vatandaş başvurularından gelen görselleri otomatik analiz ederek bakım ihtiyacını sınıflandırır, önceliklendirir, insan onayıyla doğrular ve ilgili belediye birimine aksiyonlanabilir iş kaydı olarak sunar."*

### 9 Ürün Katmanı (MVP vs Roadmap)

| # | Katman | Hackathon'da | MD Bölümü |
|---|--------|-------------|-----------|
| 1 | Visual Detection Layer | MVP ✅ | §5.1, §7 |
| 2 | Municipal Priority Engine | MVP ✅ | §5.5 |
| 3 | Human Review & Validation Layer | MVP ✅ | §5.6 |
| 4 | Privacy & Compliance Layer | MVP ✅ | §5.3 |
| 5 | Map-based Issue Inventory | Pitch only | — |
| 6 | Department Routing Layer | Pitch only | — |
| 7 | Citizen Complaint Integration (153) | Roadmap | — |
| 8 | Reporting & KPI Layer | Roadmap | — |
| 9 | Model Improvement Layer | Roadmap | — |

### Satış Paketleri

| Paket | Kapsam |
|-------|--------|
| **Pilot** | 1 ilçe, 1000–3000 görüntü, web dashboard, AI analiz, KVKK raporu |
| **Belediye Operasyon** | + mobil saha uygulaması, harita envanteri, birim yönlendirme |
| **Büyükşehir Akıllı Şehir** | + API entegrasyonu, CBS/GIS, 153 entegrasyonu, multi-model adapter, araç kamerası |

### Ticari Risk Cevapları (Jüri Q&A)

| Risk | Cevap |
|------|-------|
| "Bizde zaten şikâyet sistemi var" | Biz şikâyet sistemi değiliz; görüntüden otomatik doğrulama, önceliklendirme ve saha karar desteği katmanıyız |
| "Bunu sahada kim kullanacak?" | Bilgi İşlem kurar, Fen İşleri kullanır, Yol Bakım aksiyon alır, çağrı merkezi entegre olur |
| "Veri gizliliği riski?" | Yüz tanıma yok, plaka okuma yok, ham görüntü saklama yok, PII-avoidance-by-design |
| "Model doğruluğu düşük olabilir" | Human review + confidence threshold + precomputed fallback + pilot calibration roadmap |

---

## 22. Jüri Pitch — Demo Bugün vs Roadmap Yarın

| Demo'da gösterilen | Jüriye pitch edilen roadmap |
|-------------------|----------------------------|
| Upload → analyze → priority → KVKK → review | Harita tabanlı sorun envanteri |
| Live DETR + precomputed road-damage | Saha mobil uygulaması (field crew) |
| Human review queue (needs_review) | Belediye birim yönlendirme (Fen, Ulaşım, Temizlik) |
| Privacy report (pii_strategy) | 153/vatandaş şikâyeti entegrasyonu |
| Redis cache (ikinci çağrı anında döner) | KPI dashboards (müdahale süresi, çözüm oranı) |
| Kafka event publish (async) | Belediye verisiyle model fine-tuning |

**Scripted pitch line:** *"Bugün gördüğünüz çekirdek; haritalı envanter, birim yönlendirme ve 153 entegrasyonu aynı `AnalysisResult` şemasının üzerine adapter olarak eklenir — core değişmez."*

---

## 23. docs/rules/ YAML Şemaları (Decision Layer Kaynağı)

Bu dosyalar "Evidence-Grounded Decision Layer"ın kaynağıdır. **Henüz diskte yok — hackathon başında oluşturulmalı.**

### `docs/rules/ontology.yaml`

```yaml
version: "1.0"
mappings:
  traffic_signal:   [traffic light, traffic sign, stop sign]
  road_damage:      [pothole, crack, D00, D10, D20, D40, Repair]
  sidewalk:         [curb, sidewalk, pedestrian path]
  street_furniture: [bench, trash can, street light, bus stop]
  waste_asset:      [garbage bin, container, waste pile]
default: unknown
```

### `docs/rules/priority_policy.yaml`

```yaml
version: "1.0"
priority_by_type:
  road_damage:      high
  traffic_signal:   medium
  sidewalk:         medium
  street_furniture: low
  waste_asset:      low
  unknown:          low
overrides:
  - if_type: road_damage
    if_label_in: [D40, pothole]   # severe pothole → critical
    set_priority: critical
review_overrides:
  - if_type: unknown
    set_review: needs_review
  - if_priority: critical
    set_review: needs_review      # critical her zaman human review'a gider
```

### `docs/rules/confidence_thresholds.yaml`

```yaml
version: "1.0"
# YAML wins at startup; Go switch is compile-time default only
default:
  auto_accept: 0.80
  reject_below: 0.50
by_type:
  road_damage:    { auto_accept: 0.85, reject_below: 0.50 }
  traffic_signal: { auto_accept: 0.75, reject_below: 0.40 }
  unknown:        { auto_accept: 0.95, reject_below: 0.60 }
# Research-backed (RDD2022/YOLO-RD 2025):
# Standard images: 0.50 baseline; high-priority escalation at 0.75+
# Large/wide-angle: lower threshold 0.35 due to perspective distortion
```

> **Kural:** YAML startup'ta yüklenir, Go switch sadece parse hatasında fallback. Bu sayede model değişince veya kural güncellenince **kod değil, YAML değişir**.

---

## 24. Araştırma Bulguları — Teknik Referans

> *Bu bölüm web araştırması ile doğrulanmış, hallucination içermeyen teknik verilerdir.*

### 24.1 HuggingFace Inference API — Doğru Kullanım

**Token izni:** `inference.serverless.write` scope zorunlu. Eksikse 401 döner (auth hatası gibi görünür, gerçekte izin eksikliği).

**Doğru Go HTTP çağrısı:**
```go
// Content-Type: application/octet-stream ile raw bytes gönder — base64 gereksiz
req, _ := http.NewRequestWithContext(ctx, "POST",
    "https://router.huggingface.co/hf-inference/models/facebook/detr-resnet-50",
    bytes.NewReader(imageBytes))
req.Header.Set("Authorization", "Bearer "+hfToken)
req.Header.Set("Content-Type", "application/octet-stream")
```

**503 Cold Start retry — zorunlu:**
```go
for attempt := 0; attempt < 5; attempt++ {
    resp, err := client.Do(req)
    if resp.StatusCode == 503 {
        var cold struct{ EstimatedTime float64 `json:"estimated_time"` }
        json.NewDecoder(resp.Body).Decode(&cold)
        wait := time.Duration(cold.EstimatedTime) * time.Second
        if wait == 0 { wait = time.Duration(math.Pow(2, float64(attempt))) * time.Second }
        time.Sleep(wait)
        continue
    }
    break
}
```

**HF response format** (float coords — §6 BoundingBox float64 bunu karşılıyor):
```json
[{"label": "car", "score": 0.97, "box": {"xmin":10.4,"ymin":20.1,"xmax":100.3,"ymax":80.7}}]
```

**Önemli:** `hupe1980/go-huggingface` paketi **yalnızca NLP** destekliyor — object detection için kullanma. Raw `net/http` yaz (~30 satır).

**DETR'i ısıt:** Demo öncesi `T-10 dakika` boş bir curl atarak cold start'ı temizle.

### 24.2 Research-Backed Confidence Thresholds

Kaynak: RDD2022 / YOLO-RD 2025 (PMC), RDD-YOLO MDPI 2024:

| Koşul | Threshold | Priority |
|-------|-----------|----------|
| Standard road image | ≥ 0.75 | HIGH |
| Standard road image | 0.50–0.74 | MEDIUM |
| Wide-angle/perspective | ≥ 0.35 | LOW |
| General auto-accept | ≥ 0.80 | auto_accepted |
| General needs_review | 0.50–0.79 | needs_review |
| General reject | < 0.50 | rejected |

DETR default threshold (HF API): `0.5` — road-irrelevant COCO sınıfları (person, bicycle) label bazlı filtrele.

### 24.3 OpenRouter — Doğru Model ve Yapı

**Önerilen model:** `google/gemma-4-26b-a4b-it:free`
- 256K context, native structured output, multimodal
- Fallback: `openai/gpt-oss-20b:free`
- **Kullanma:** `openrouter/free` — deterministik değil

**Strict JSON schema çağrısı:**
```json
{
  "model": "google/gemma-4-26b-a4b-it:free",
  "response_format": {
    "type": "json_schema",
    "json_schema": {
      "name": "civic_report",
      "strict": true,
      "schema": {
        "type": "object",
        "properties": {
          "summary":        {"type": "string"},
          "recommended_action": {"type": "string"},
          "risk_level":     {"type": "string", "enum": ["low","medium","high"]},
          "kvkk_note":      {"type": "string"}
        },
        "required": ["summary","recommended_action","risk_level","kvkk_note"],
        "additionalProperties": false
      }
    }
  }
}
```

**Rate limit:** Free tier = 50 req/day. Demo için yeterli. $10 yükle → 1000 req/day.
**Outage:** 401 like auth error → önce status page kontrol et, hardcoded fallback template dön.

### 24.4 KVKK — Araştırmayla Doğrulanmış Gereksinimler

- KVKK Md. 3: kamusal alan gizliliği ortadan kaldırmaz — tanımlanabilir kişiler kişisel veri kapsamında
- Yüz blur: **geri döndürülemez** olmalı (şifreli blur KVKK'yı karşılamaz)
- Google Street View yaklaşımı: yüz >89%, plaka 94–96% otomatik blur
- **Cross-border:** HF API'ye görsel gönderimi KVKK Md. 9 kapsamında (US'e transfer). Hackathon için belgelenmiş risk; üretim için hukuki inceleme şart.
- **PII-avoidance-by-design:** MVP'de Street View kullanılmıyor → anonymization meselesi yok

### 24.5 Go Hexagonal Architecture — Doğru Port Konumu

ThreeDots Labs canonical pattern (en çok referans alınan Go clean arch):

```go
// Portlar consumer'a aittir — infra'da değil, app/ports/out/'da
// application/vision/ports.go
type VisionModel interface {
    Detect(ctx context.Context, imageBytes []byte, threshold float64) ([]Detection, error)
}

// Domain hataları: adapter bunları wrap eder
var ErrVisionUnavailable = errors.New("vision service unavailable")
var ErrModelColdStart    = errors.New("model loading, retry")
```

Adapter boundary'de domain error wrap:
```go
if resp.StatusCode == 503 {
    return nil, fmt.Errorf("hf adapter: %w", domain.ErrModelColdStart)
}
```

Use case'de:
```go
if errors.Is(err, domain.ErrModelColdStart) {
    // retry or fallback to precomputed
}
```

### 24.6 Civic Tech AI — Sektör Bulguları

- Commercial platforms (RoadBotics, Pavesight, Malibu): AI detects → **human approves** → work order. Hiçbir production sistemi otomatik iş emri açmıyor.
- PaveX benchmark (ASCE Jan 2026): PCI ±5 puan insan değerlendirmesine yakın = belediye kabul eşiği
- 2025 EY survey: %67 belediye AI integrasyon yapıyor — gap: explainable output + GIS/311 entegrasyonu
- **Açık kaynak civic AI road inspection aracı yok** — CivicLens open source yayınlanırsa gerçekten yeni

---

## 25. API Error Contract

Tüm endpointler tek hata zarfı:

```json
{
  "error": {
    "code": "UNSUPPORTED_MEDIA_TYPE",
    "message": "Only image/jpeg and image/png are accepted.",
    "correlation_id": "ana_001"
  }
}
```

| Durum | HTTP | code |
|-------|------|------|
| Yanlış content-type | 415 | `UNSUPPORTED_MEDIA_TYPE` |
| Görsel > 10 MB | 413 | `PAYLOAD_TOO_LARGE` |
| Eksik / bozuk body | 400 | `INVALID_REQUEST` |
| HF down → precomputed devreye girdi | 200 | — (`model_mode:"precomputed"`) |
| HF down, fallback yok | 503 | `INFERENCE_UNAVAILABLE` |
| Internal hata | 500 | `INTERNAL_ERROR` |

**Input limitleri:** max 10 MB, yalnızca `image/jpeg` ve `image/png`, tek görsel per request.

---

## 26. MODEL_CARD.md Şablonu

> Bu dosya `docs/MODEL_CARD.md` olarak hackathon sırasında oluşturulmalı. Şu an diskte yok.

### Zorunlu İçerik

**Kullanılan modeller:**

| Model | HF ID | Rol | Sınır |
|-------|--------|-----|-------|
| DETR ResNet-50 | `facebook/detr-resnet-50` | Live genel nesne tespiti | **COCO'da pothole/crack yok** |
| RDD-YOLO | `rezzzq/yolo12s-road-damage-rdd2022` | Yol hasarı (D00/D10/D20/D40) | Precomputed; live Inference API olmayabilir |

**Doğruluk beyanı (zorunlu):**
> *"Alan doğruluğu yüzdesi iddia edilmemektedir. Hiçbir mAP/precision değeri belirtilmemiştir; belediye onaylı test seti kullanılmamıştır. Güven skorları model raporlamasıdır, doğrulanmış ground truth değildir."*

**Dataset & Lisanslar:**

| Dataset | Kullanım | Lisans |
|---------|---------|--------|
| RDD2022 | Road damage reference | Belediye kullanım uyumlu |
| Mapillary Vistas v2 | Segmentation referansı | CC BY-NC-SA — yalnızca araştırma dayanağı |
| COCO | DETR baseline | Apache 2.0 |
| Cityscapes | Urban seg referansı | Non-commercial research |
| BDD100K | Scene referansı | BSD |

**Hedeflenen kullanım:** Yalnızca inanimate kentsel nesneler. Kişi/araç tanımlama yasak.

**Bilinen hata modları:**
- HF cold-start 503
- DETR'in COCO class gap'i (pothole, crack yok)
- Precomputed sample sabit görsellerdir
- Small object (<32×32px) tespiti düşük (YOLO-RD 2025 verisi: mAP_s ~17.66)

**PII stratejisi:** `avoidance_by_design` — MVP'de Street View yok, PII sisteme girmiyor.

---

## 27. Deployment & Demo Runbook

### Render (Go Backend)

```bash
# Build command
go build -o app ./cmd/server

# Start command
./app

# Health check path (Render dashboard'da ayarla)
/health/live

# Zorunlu env vars
PORT=8080
HF_API_TOKEN=hf_...             # inference.serverless.write scope!
OPENROUTER_API_KEY=sk-or-...
REDIS_URL=redis://...            # Render Redis add-on veya Upstash
ALLOWED_ORIGINS=https://civiclens.vercel.app,http://localhost:3000
```

**Cold start uyarısı:** Render free tier 15 dk sonra uyuyor → demo öncesi T-5 dk `curl https://<render-url>/health/ready` ile ısıt.

### Vercel (Next.js Web)

```bash
# Root directory: web/
# Framework: Next.js (auto-detected)
# Env var (Vercel dashboard'da):
NEXT_PUBLIC_API_URL=https://<render-app>.onrender.com
```

**CORS:** Backend `ALLOWED_ORIGINS`'e Vercel URL (`*.vercel.app`) eklenmeli — şu anki stub yalnızca localhost'a izin veriyor.

### HF Pre-Warm

```bash
# Demo öncesi T-10 dakika çalıştır
curl -X POST https://router.huggingface.co/hf-inference/models/facebook/detr-resnet-50 \
  -H "Authorization: Bearer $HF_API_TOKEN" \
  -H "Content-Type: application/octet-stream" \
  --data-binary @docs/sample_images/warmup.jpg
# 503 gelirse estimated_time kadar bekle ve tekrar dene
```

### Env Var Master Tablosu

| Değişken | Kullanım yeri | `.env.example` değeri |
|----------|--------------|----------------------|
| `HF_API_TOKEN` | Backend HF adapter | `hf_YOUR_TOKEN` |
| `OPENROUTER_API_KEY` | Backend OpenRouter adapter | `sk-or-YOUR_KEY` |
| `STREETVIEW_API_KEY` | Street View (opsiyonel) | `YOUR_STREETVIEW_KEY` |
| `REDIS_URL` | Backend Redis adapter | `redis://localhost:6379` |
| `KAFKA_BROKERS` | Backend Kafka publisher | `localhost:9092` |
| `PORT` | Backend HTTP server | `8080` |
| `ALLOWED_ORIGINS` | CORS middleware | `http://localhost:3000` |
| `NEXT_PUBLIC_API_URL` | Web frontend | `http://localhost:8080` |
| `EXPO_PUBLIC_API_URL` | Mobile frontend | `http://localhost:8080` |

### Pre-Demo Checklist

- [ ] `curl https://<render>/health/ready` → `{"status":"ok"}`
- [ ] HF DETR ısınmış (T-10 min curl yapıldı)
- [ ] Precomputed road-damage sample `/demo-results` dönüyor
- [ ] OpenRouter raporu pre-generated (canlı tıklamada gecikmez)
- [ ] Redis cache: ikinci analiz anında dönüyor (göster)
- [ ] CORS: web frontend backend'e ulaşıyor
- [ ] KVKK-COMPLIANCE.md doldurulmuş ve imzalanmış
- [ ] `git log --oneline` incremental commit var (toplu değil)
- [ ] `.env` commitleri yok (`git status` temiz)

---

## 28. Cursor AI-Assisted Development

> *10 puan + bonus — şu an neredeyse sıfır puanla bırakılıyor. Bu bölüm README'ye ve `docs/AI_USAGE.md`'e taşınmalı.*

### Mevcut Cursor Ruleset (`.cursor/rules/`)

| Dosya | Kapsam |
|-------|--------|
| `hackathon.mdc` | Genel hackathon kuralları, stack, commit disiplini |
| `workflow.mdc` | Geliştirme workflow'u |
| `backend.mdc` | Go backend kuralları |
| `web.mdc` | Next.js kuralları |
| `mobile.mdc` | Expo kuralları |
| `.cursor/mcp.json` | MCP server konfigürasyonu |

### Model Stratejisi (AGENTS.md'den)

| Görev | Model |
|-------|-------|
| Mimari / çok dosya planlama | Claude Opus 4 |
| Tek dosya implementasyon | Claude Sonnet 4.5 |
| 2+ dosya değişikliği | Plan Mode (Shift+Tab) + Opus |

### Kanıtlanabilir AI Kullanımı

README ve `docs/AI_USAGE.md`'e şunları yaz:
- Bu mimari doc Opus Ask Mode ile iteratif zenginleştirildi
- vision bounded context scaffolding'i Cursor Agent Mode + `@Codebase` ile üretildi
- Commit mesajları `@Commit` komutu ile oluşturuldu
- Precomputed sample JSON Cursor üretimi
- docs/rules/*.yaml şemaları Cursor ile draft edildi

### Claude Code CLI Kullanımı (Bonus Puan)

```bash
# Eğer hackathon sırasında Claude Code CLI kullanılırsa:
claude "Generate the HF DETR adapter for the vision bounded context"
claude "Review the privacy_guard.go for KVKK compliance"
claude "Add 503 retry logic to the HF client"
```

Her CLI komutu `docs/AI_USAGE.md`'e logla → jüri kanıtı.

---

## 29. Dokümantasyon Durumu

> Hangi dosyalar var, hangisi eksik — diskte kontrol edildi.

| Dosya | Durum | Öncelik |
|-------|-------|---------|
| `docs/KVKK-COMPLIANCE.md` | ✅ var (template doldurulacak) | **Kritik — prize prerequisite** |
| `docs/PRIVACY.md` | ✅ var (kapsamlı) | — |
| `docs/DECISIONS.md` | ✅ var | — |
| `docs/DEMO.md` | ✅ var | — |
| `docs/2026-06-06-civiclens-core-design.md` | ✅ bu doc | — |
| `docs/MODEL_CARD.md` | ❌ eksik (4x referans) | **Kritik** |
| `docs/AI_USAGE.md` | ❌ eksik (Cursor bonus pts) | Yüksek |
| `docs/rules/ontology.yaml` | ❌ eksik (§5.2, §23) | Yüksek |
| `docs/rules/priority_policy.yaml` | ❌ eksik (§5.5, §9.1, §23) | Yüksek |
| `docs/rules/confidence_thresholds.yaml` | ❌ eksik (§5.4, §23) | Yüksek |
| `docs/HF_RESEARCH.md` | ❌ eksik | Orta |
| `docs/EVALUATION.md` | ❌ eksik | Orta |
| `docs/OPENROUTER_REASONING.md` | ❌ eksik | Düşük |
| `docs/ANONYMIZATION_AND_DELETION.md` | ❌ eksik | Orta |

**KVKK-COMPLIANCE.md doldurmak için CivicLens bilgileri:**
- Purpose limitation: *"Yalnızca inanimate kentsel nesneler (yol, trafik altyapısı, şehir mobilyası). Kimlik tespiti, yüz tanıma, plaka okuma yasak."*
- PII strategy: `avoidance_by_design` — raw image sisteme girmiyor, HF API'ye anonymize edilmeden görsel gönderilmez
- Deletion status: `raw_image_not_persisted` — her `AnalysisResult`'ta belgeleniyor

---

*Bu doc yaşayan bir belgedir. Hackathon boyunca her önemli karar buraya eklenir.*
*Son güncelleme: 2026-06-06 — §21-29 eklendi (gap analizi + araştırma bulguları + Kafka/Redis)*

*Bu doc yaşayan bir belgedir. Hackathon boyunca her önemli karar buraya eklenir.*
