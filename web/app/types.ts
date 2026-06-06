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

// ─── Wave 2 operations types ─────────────────────────────────────────────────

export type UserRole = "citizen" | "field_staff" | "operator" | "manager";

export type Location = { lat: number; lng: number };

export type Report = {
  report_id: string;
  source_type: string;
  reporter_role: string;
  description: string;
  location: Location | null;
  image_ref: string;
  analysis_id: string;
  problem_type: string;
  priority: string;
  review_status: string;
  assigned_department: string;
  duplicate_of?: string;
  duplicate_count: number;
  status: string;
  created_at: string;
};

export type Task = {
  task_id: string;
  report_id: string;
  assigned_department: string;
  assigned_to: string;
  priority: string;
  status: string;
  sla: string;
  created_at: string;
  updated_at: string;
};

export type AnalyticsSummary = {
  total_reports: number;
  reports_by_status: Record<string, number>;
  reports_by_type: Record<string, number>;
  reports_by_department: Record<string, number>;
  needs_review: number;
  total_tasks: number;
  tasks_by_status: Record<string, number>;
  completed_tasks: number;
  avg_resolution_hours: number;
};

export type ReviewDecision = "accepted" | "rejected";

export type ReviewResponse = {
  report: Report;
  task?: Task;
};
