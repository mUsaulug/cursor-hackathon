import type { Task } from "@/app/types";
import { formatDate, PriorityBadge, StatusBadge } from "./badges";

export default function TaskCard({
  task,
  children,
}: {
  task: Task;
  children?: React.ReactNode;
}) {
  return (
    <article className="flex flex-col gap-3 rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <span className="font-mono text-xs text-slate-400">{task.task_id}</span>
          <p className="mt-0.5 text-sm font-semibold text-slate-800">
            Rapor: {task.report_id}
          </p>
        </div>
        <div className="flex flex-wrap gap-1.5">
          <StatusBadge status={task.status} />
          <PriorityBadge priority={task.priority} />
        </div>
      </div>
      <dl className="mt-3 grid grid-cols-2 gap-x-4 gap-y-1.5 text-xs sm:grid-cols-3">
        <div>
          <dt className="text-slate-400">Birim</dt>
          <dd className="font-medium text-slate-700">{task.assigned_department}</dd>
        </div>
        <div>
          <dt className="text-slate-400">Atanan</dt>
          <dd className="font-medium text-slate-700">{task.assigned_to}</dd>
        </div>
        <div>
          <dt className="text-slate-400">SLA</dt>
          <dd className="font-medium text-slate-700">{task.sla}</dd>
        </div>
        <div>
          <dt className="text-slate-400">Oluşturulma</dt>
          <dd className="font-medium text-slate-700">{formatDate(task.created_at)}</dd>
        </div>
        <div>
          <dt className="text-slate-400">Güncelleme</dt>
          <dd className="font-medium text-slate-700">{formatDate(task.updated_at)}</dd>
        </div>
      </dl>
      {children ? (
        <div className="flex flex-wrap items-center gap-2 border-t border-slate-100 pt-3">
          {children}
        </div>
      ) : null}
    </article>
  );
}
