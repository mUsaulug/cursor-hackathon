"use client";

import { useCallback, useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import type { Task, UserRole } from "@/app/types";
import TaskCard from "./TaskCard";

type FieldStaffViewProps = {
  role: UserRole;
};

export default function FieldStaffView({ role }: FieldStaffViewProps) {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchTasks = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch("/api/v1/tasks", role, { cache: "no-store" });
      if (!res.ok) {
        throw new Error(`Görevler alınamadı (${res.status})`);
      }
      const data: Task[] = await res.json();
      setTasks(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Bilinmeyen hata");
    } finally {
      setLoading(false);
    }
  }, [role]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void fetchTasks();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [fetchTasks]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-slate-900">Saha Görevleri</h2>
          <p className="mt-1 text-sm text-slate-600">
            Size atanan saha görevlerini görüntüleyin.
          </p>
        </div>
        <button
          type="button"
          disabled={loading}
          onClick={() => void fetchTasks()}
          className="rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 shadow-sm transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? "Yükleniyor…" : "Yenile"}
        </button>
      </div>

      {loading && tasks.length === 0 ? (
        <p className="text-sm text-slate-500">Görevler yükleniyor…</p>
      ) : null}

      {error ? (
        <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </p>
      ) : null}

      {!loading && !error && tasks.length === 0 ? (
        <div className="rounded-xl border border-slate-200 bg-white p-10 text-center shadow-sm">
          <p className="text-sm text-slate-500">Atanmış görev yok.</p>
        </div>
      ) : null}

      {tasks.length > 0 ? (
        <div className="grid gap-4 sm:grid-cols-2">
          {tasks.map((task) => (
            <TaskCard key={task.task_id} task={task} />
          ))}
        </div>
      ) : null}
    </div>
  );
}
