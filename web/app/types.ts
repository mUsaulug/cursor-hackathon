export type BoundingBox = {
  xmin: number;
  ymin: number;
  xmax: number;
  ymax: number;
};

export type Priority = "critical" | "high" | "medium" | "low";

export type ReviewStatus = "auto_accepted" | "needs_review" | "rejected";

export type Detection = {
  id: string;
  label: string;
  normalized_object_type: string;
  confidence: number;
  bbox: BoundingBox;
  model_id: string;
  review_status: ReviewStatus;
  priority: Priority;
  reason: string;
};

export type PrivacyReport = {
  kvkk_safe: boolean;
  raw_image_stored: boolean;
  anonymized: boolean;
  deletion_status: string;
  blocked_count: number;
  pii_strategy: string;
};

export type AnalysisResult = {
  schema_version: string;
  analysis_id: string;
  source_type: string;
  source_ref: string;
  location: { lat: number; lng: number } | null;
  model_id: string;
  model_mode: string;
  image_width: number;
  image_height: number;
  raw_image_stored: boolean;
  anonymized: boolean;
  kvkk_safe: boolean;
  privacy: PrivacyReport;
  detections: Detection[];
  created_at: string;
  deletion_status: string;
};

export type MaintenanceReport = {
  summary: string;
  recommended_action: string;
  risk_level: string;
  kvkk_note: string;
};

export type ReviewItem = {
  analysis_id: string;
  detection_id: string;
  label: string;
  normalized_object_type: string;
  confidence: number;
  priority: string;
  reason: string;
};

export type SummaryResponse = {
  analysis_id: string;
  report: MaintenanceReport | null;
};
