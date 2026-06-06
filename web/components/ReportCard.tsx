import type { Report } from "@/app/types";
import { formatDate, PriorityBadge, StatusBadge, Badge } from "./badges";

type ReportCardProps = {
  report: Report;
  highlight?: boolean;
  children?: React.ReactNode;
};

export default function ReportCard({
  report,
  highlight = false,
  children,
}: ReportCardProps) {
  return (
    <article
      className={`rounded-xl border bg-white p-5 shadow-sm flex flex-col gap-3 ${
        highlight
          ? "border-amber-400 ring-2 ring-amber-100"
          : "border-slate-200"
      }`}
    >
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div className="flex flex-col gap-0.5">
          <span className="font-mono text-xs text-slate-400">
            {report.report_id}
          </span>
          <span className="text-sm font-semibold text-slate-800">
            {report.problem_type || "Sorun türü belirsiz"}
          </span>
        </div>
        <div className="flex flex-wrap gap-1.5">
          <StatusBadge status={report.status} />
          {report.priority ? <PriorityBadge priority={report.priority} /> : null}
          {report.duplicate_count > 0 ? (
            <Badge className="bg-violet-100 text-violet-800">
              {report.duplicate_count} yineleme
            </Badge>
          ) : null}
        </div>
      </div>

      <p className="text-sm leading-relaxed text-slate-600">
        {report.description}
      </p>

      <dl className="grid grid-cols-2 gap-x-4 gap-y-1.5 text-xs sm:grid-cols-3">
        <div>
          <dt className="text-slate-400">Kaynak</dt>
          <dd className="font-medium text-slate-700">{report.source_type}</dd>
        </div>
        <div>
          <dt className="text-slate-400">İnceleme</dt>
          <dd className="font-medium text-slate-700">{report.review_status}</dd>
        </div>
        {report.assigned_department ? (
          <div>
            <dt className="text-slate-400">Birim</dt>
            <dd className="font-medium text-slate-700">
              {report.assigned_department}
            </dd>
          </div>
        ) : null}
        {report.location ? (
          <div>
            <dt className="text-slate-400">Konum</dt>
            <dd className="font-medium text-slate-700">
              {report.location.lat.toFixed(4)}, {report.location.lng.toFixed(4)}
            </dd>
          </div>
        ) : null}
        <div className="col-span-2 sm:col-span-1">
          <dt className="text-slate-400">Tarih</dt>
          <dd className="font-medium text-slate-700">
            {formatDate(report.created_at)}
          </dd>
        </div>
      </dl>

      {children ? <div className="flex flex-wrap gap-2 pt-1">{children}</div> : null}
    </article>
  );
}
