"use client";

import { useCallback, useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import type { AnalyticsSummary, Task, UserRole } from "@/app/types";
import TaskCard from "./TaskCard";

type ManagerViewProps = {
  role: UserRole;
};

function StatCard({
  label,
  value,
  accent = "text-slate-900",
}: {
  label: string;
  value: string | number;
  accent?: string;
}) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
      <p className="text-sm font-medium text-slate-500">{label}</p>
      <p className={`mt-2 text-3xl font-bold tracking-tight ${accent}`}>
        {value}
      </p>
    </div>
  );
}

function BreakdownCard({
  title,
  data,
}: {
  title: string;
  data: Record<string, number>;
}) {
  const entries = Object.entries(data);
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
      <h3 className="mb-3 text-sm font-semibold text-slate-700">{title}</h3>
      {entries.length === 0 ? (
        <p className="text-sm text-slate-500">Veri yok</p>
      ) : (
        <ul className="space-y-2">
          {entries.map(([key, count]) => (
            <li
              key={key}
              className="flex items-center justify-between text-sm"
            >
              <span className="text-slate-600">{key.replace(/_/g, " ")}</span>
              <span className="font-semibold text-slate-900">{count}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

export default function ManagerView({ role }: ManagerViewProps) {
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [closingId, setClosingId] = useState<string | null>(null);
  const [auditLines, setAuditLines] = useState<string[]>([]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [summaryRes, tasksRes] = await Promise.all([
        apiFetch("/api/v1/analytics/summary", role, { cache: "no-store" }),
        apiFetch("/api/v1/tasks", role, { cache: "no-store" }),
      ]);

      if (!summaryRes.ok) {
        throw new Error(`Analitik alınamadı (${summaryRes.status})`);
      }
      if (!tasksRes.ok) {
        throw new Error(`Görevler alınamadı (${tasksRes.status})`);
      }

      const auditRes = await apiFetch("/api/v1/audit", role, {
        cache: "no-store",
      });
      const summaryData: AnalyticsSummary = await summaryRes.json();
      const tasksData: Task[] = await tasksRes.json();
      setSummary(summaryData);
      setTasks(tasksData);
      if (auditRes.ok) {
        const audit = (await auditRes.json()) as Array<{
          method: string;
          path: string;
          actor_role: string;
          status: number;
          at: string;
        }>;
        setAuditLines(
          audit.slice(-8).map(
            (e) =>
              `${e.at} ${e.actor_role} ${e.method} ${e.path} → ${e.status}`,
          ),
        );
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Bilinmeyen hata");
    } finally {
      setLoading(false);
    }
  }, [role]);

  const handleClose = useCallback(
    async (taskId: string) => {
      setClosingId(taskId);
      setError(null);
      try {
        const res = await apiFetch(`/api/v1/tasks/${taskId}/close`, role, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ decision: "approved" }),
        });
        if (!res.ok) {
          throw new Error(`Görev kapatılamadı (${res.status})`);
        }
        await fetchData();
      } catch (err) {
        setError(err instanceof Error ? err.message : "Bilinmeyen hata");
      } finally {
        setClosingId(null);
      }
    },
    [fetchData, role],
  );

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void fetchData();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [fetchData]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-lg font-semibold text-slate-900">
          Yönetici Panosu
        </h2>
        <button
          type="button"
          disabled={loading}
          onClick={() => void fetchData()}
          className="rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 shadow-sm transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? "Yükleniyor…" : "Yenile"}
        </button>
      </div>

      {loading && !summary ? (
        <p className="text-sm text-slate-500">Veriler yükleniyor…</p>
      ) : null}

      {error ? (
        <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </p>
      ) : null}

      {summary ? (
        <>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <StatCard label="Toplam Bildirim" value={summary.total_reports} />
            <StatCard
              label="İnceleme Bekleyen"
              value={summary.needs_review}
              accent="text-amber-700"
            />
            <StatCard
              label="Tamamlanan Görev"
              value={summary.completed_tasks}
              accent="text-emerald-700"
            />
            <StatCard
              label="Ort. Çözüm (saat)"
              value={summary.avg_resolution_hours.toFixed(1)}
              accent="text-indigo-700"
            />
          </div>

          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <BreakdownCard
              title="Bildirimler — Durum"
              data={summary.reports_by_status}
            />
            <BreakdownCard
              title="Bildirimler — Tür"
              data={summary.reports_by_type}
            />
            <BreakdownCard
              title="Bildirimler — Birim"
              data={summary.reports_by_department}
            />
            <BreakdownCard
              title="Görevler — Durum"
              data={summary.tasks_by_status}
            />
          </div>
        </>
      ) : null}

      <section>
        <h3 className="mb-4 text-base font-semibold text-slate-900">
          Görevler
          {tasks.length > 0 ? (
            <span className="ml-2 text-sm font-normal text-slate-500">
              ({tasks.length})
            </span>
          ) : null}
        </h3>
        {tasks.length === 0 && !loading ? (
          <p className="text-sm text-slate-500">Henüz görev yok.</p>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2">
            {tasks.map((task) => (
              <TaskCard key={task.task_id} task={task}>
                {task.status === "ai_verified" ||
                task.status === "evidence_uploaded" ? (
                  <button
                    type="button"
                    disabled={closingId === task.task_id}
                    onClick={() => void handleClose(task.task_id)}
                    className="rounded-md bg-emerald-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-emerald-700 disabled:opacity-50"
                  >
                    Kapat
                  </button>
                ) : null}
              </TaskCard>
            ))}
          </div>
        )}
      </section>

      {auditLines.length > 0 ? (
        <section className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
          <h3 className="mb-3 text-base font-semibold text-slate-900">
            Denetim günlüğü (son kayıtlar)
          </h3>
          <ul className="space-y-1 font-mono text-xs text-slate-600">
            {auditLines.map((line) => (
              <li key={line}>{line}</li>
            ))}
          </ul>
        </section>
      ) : null}
    </div>
  );
}
