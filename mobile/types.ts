export type Role = 'citizen' | 'field_staff';

export type Task = {
  task_id: string;
  report_id: string;
  assigned_department: string;
  assigned_to: string;
  priority: string;
  status: string;
  sla: string;
};

export type Report = {
  report_id: string;
  problem_type: string;
  priority: string;
  assigned_department: string;
  status: string;
};

export type Evidence = {
  evidence_id: string;
  task_id: string;
  ai_verification: string;
  manager_approval: string;
};

export type Detection = {
  id: string;
  label: string;
  normalized_object_type: string;
  confidence: number;
  review_status: string;
  priority: string;
  reason: string;
};

export type AnalysisResult = {
  analysis_id: string;
  model_mode: string;
  model_id: string;
  kvkk_safe: boolean;
  detections: Detection[];
  created_at: string;
};

export type ReactNativeFile = {
  uri: string;
  name: string;
  type: string;
};
