import type { Detection, Priority, ReviewStatus } from "@/app/types";

const PRIORITY_STYLES: Record<Priority, string> = {
  critical: "bg-red-100 text-red-800",
  high: "bg-orange-100 text-orange-800",
  medium: "bg-blue-100 text-blue-800",
  low: "bg-slate-100 text-slate-700",
};

const PRIORITY_LABELS: Record<Priority, string> = {
  critical: "Kritik",
  high: "Yüksek",
  medium: "Orta",
  low: "Düşük",
};

const REVIEW_STYLES: Record<ReviewStatus, string> = {
  auto_accepted: "bg-emerald-100 text-emerald-800",
  needs_review: "bg-amber-100 text-amber-800",
  rejected: "bg-slate-200 text-slate-600",
};

const REVIEW_LABELS: Record<ReviewStatus, string> = {
  auto_accepted: "Otomatik onay",
  needs_review: "İnceleme gerekli",
  rejected: "Reddedildi",
};

function formatObjectType(type: string): string {
  return type
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

type DetectionCardProps = {
  detection: Detection;
  children?: React.ReactNode;
};

export default function DetectionCard({
  detection,
  children,
}: DetectionCardProps) {
  const confidencePct = Math.round(detection.confidence * 100);

  return (
    <article className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <h3 className="font-semibold text-slate-900">
          {formatObjectType(detection.normalized_object_type)}
        </h3>
        <div className="flex flex-wrap gap-1.5">
          <span
            className={`inline-flex rounded-full px-2 py-0.5 text-xs font-semibold ${PRIORITY_STYLES[detection.priority]}`}
          >
            {PRIORITY_LABELS[detection.priority]}
          </span>
          <span
            className={`inline-flex rounded-full px-2 py-0.5 text-xs font-semibold ${REVIEW_STYLES[detection.review_status]}`}
          >
            {REVIEW_LABELS[detection.review_status]}
          </span>
        </div>
      </div>
      <p className="mt-2 text-sm text-slate-600">
        Güven:{" "}
        <span className="font-medium text-slate-800">{confidencePct}%</span>
      </p>
      <p className="mt-2 text-sm leading-relaxed text-slate-700">
        {detection.reason}
      </p>
      {children ? <div className="mt-3 flex gap-2">{children}</div> : null}
    </article>
  );
}

export { formatObjectType };
