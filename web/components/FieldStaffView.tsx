"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { apiFetch } from "@/lib/api";
import type { Evidence, Task, UserRole } from "@/app/types";
import TaskCard from "./TaskCard";

const ASSIGNED_TO = "saha_ekip_1";

type FieldStaffViewProps = {
  role: UserRole;
};

export default function FieldStaffView({ role }: FieldStaffViewProps) {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [busyTaskId, setBusyTaskId] = useState<string | null>(null);
  const [lastEvidence, setLastEvidence] = useState<Evidence | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const pendingEvidenceTaskId = useRef<string | null>(null);

  const fetchTasks = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch(
        `/api/v1/tasks?assigned_to=${ASSIGNED_TO}`,
        role,
        { cache: "no-store" },
      );
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

  const handleStart = useCallback(
    async (taskId: string) => {
      setBusyTaskId(taskId);
      setError(null);
      try {
        const res = await apiFetch(`/api/v1/tasks/${taskId}/start`, role, {
          method: "POST",
        });
        if (!res.ok) {
          throw new Error(`Görev başlatılamadı (${res.status})`);
        }
        await fetchTasks();
      } catch (err) {
        setError(err instanceof Error ? err.message : "Bilinmeyen hata");
      } finally {
        setBusyTaskId(null);
      }
    },
    [fetchTasks, role],
  );

  const handleEvidencePick = useCallback((taskId: string) => {
    pendingEvidenceTaskId.current = taskId;
    fileInputRef.current?.click();
  }, []);

  const handleEvidenceFile = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      const taskId = pendingEvidenceTaskId.current;
      e.target.value = "";
      if (!file || !taskId) return;

      setBusyTaskId(taskId);
      setError(null);
      try {
        const body = new FormData();
        body.append("image", file);
        const res = await apiFetch(`/api/v1/tasks/${taskId}/evidence`, role, {
          method: "POST",
          body,
        });
        if (!res.ok) {
          const text = await res.text();
          throw new Error(text || `Kanıt yüklenemedi (${res.status})`);
        }
        const ev: Evidence = await res.json();
        setLastEvidence(ev);
        await fetchTasks();
      } catch (err) {
        setError(err instanceof Error ? err.message : "Bilinmeyen hata");
      } finally {
        setBusyTaskId(null);
        pendingEvidenceTaskId.current = null;
      }
    },
    [fetchTasks, role],
  );

  return (
    <div className="space-y-6">
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        className="hidden"
        onChange={(e) => void handleEvidenceFile(e)}
      />

      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-slate-900">Saha Görevleri</h2>
          <p className="mt-1 text-sm text-slate-600">
            Atanan görevleri başlatın ve tamamlama kanıtı yükleyin.
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
            <TaskCard key={task.task_id} task={task}>
              {task.status === "assigned" ? (
                <button
                  type="button"
                  disabled={busyTaskId === task.task_id}
                  onClick={() => void handleStart(task.task_id)}
                  className="rounded-md bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700 disabled:opacity-50"
                >
                  Başlat
                </button>
              ) : null}
              {task.status === "started" ? (
                <button
                  type="button"
                  disabled={busyTaskId === task.task_id}
                  onClick={() => handleEvidencePick(task.task_id)}
                  className="rounded-md bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
                >
                  Kanıt Yükle
                </button>
              ) : null}
              {lastEvidence?.task_id === task.task_id ? (
                <span className="text-xs text-emerald-700">
                  Kanıt yüklendi — {lastEvidence.ai_verification}
                </span>
              ) : null}
            </TaskCard>
          ))}
        </div>
      ) : null}
    </div>
  );
}
