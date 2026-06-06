"use client";

import { useCallback, useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import type { Report, ReviewResponse, UserRole } from "@/app/types";
import ReportCard from "./ReportCard";
import TaskCard from "./TaskCard";

type OperatorViewProps = {
  role: UserRole;
};

export default function OperatorView({ role }: OperatorViewProps) {
  const [reports, setReports] = useState<Report[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [reviewingId, setReviewingId] = useState<string | null>(null);
  const [lastReview, setLastReview] = useState<ReviewResponse | null>(null);

  const fetchReports = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch("/api/v1/reports", role, { cache: "no-store" });
      if (!res.ok) {
        throw new Error(`Bildirimler alınamadı (${res.status})`);
      }
      const data: Report[] = await res.json();
      setReports(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Bilinmeyen hata");
    } finally {
      setLoading(false);
    }
  }, [role]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void fetchReports();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [fetchReports]);

  const handleReview = useCallback(
    async (reportId: string, decision: "accepted" | "rejected") => {
      setReviewingId(reportId);
      setError(null);

      try {
        const body =
          decision === "accepted"
            ? { decision: "accepted", assigned_to: "saha_ekip_1" }
            : { decision: "rejected" };

        const res = await apiFetch(`/api/v1/reports/${reportId}/review`, role, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body),
        });

        if (!res.ok) {
          const text = await res.text();
          throw new Error(text || `İnceleme başarısız (${res.status})`);
        }

        const result: ReviewResponse = await res.json();
        setLastReview(result);
        await fetchReports();
      } catch (err) {
        setError(err instanceof Error ? err.message : "Bilinmeyen hata");
      } finally {
        setReviewingId(null);
      }
    },
    [fetchReports, role],
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-lg font-semibold text-slate-900">
          Bildirim Kuyruğu
          {reports.length > 0 ? (
            <span className="ml-2 text-sm font-normal text-slate-500">
              ({reports.length})
            </span>
          ) : null}
        </h2>
        <button
          type="button"
          disabled={loading}
          onClick={() => void fetchReports()}
          className="rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 shadow-sm transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? "Yükleniyor…" : "Yenile"}
        </button>
      </div>

      {loading && reports.length === 0 ? (
        <p className="text-sm text-slate-500">Bildirimler yükleniyor…</p>
      ) : null}

      {error ? (
        <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
        </p>
      ) : null}

      {!loading && !error && reports.length === 0 ? (
        <div className="rounded-xl border border-slate-200 bg-white p-10 text-center shadow-sm">
          <p className="text-sm text-slate-500">Henüz bildirim yok.</p>
        </div>
      ) : null}

      {reports.length > 0 ? (
        <div className="grid gap-4">
          {reports.map((report) => (
            <ReportCard
              key={report.report_id}
              report={report}
              highlight={report.status === "waiting_for_review"}
            >
              {report.status === "waiting_for_review" ? (
                <>
                  <button
                    type="button"
                    disabled={reviewingId === report.report_id}
                    onClick={() => void handleReview(report.report_id, "accepted")}
                    className="rounded-md bg-emerald-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-emerald-700 disabled:opacity-50"
                  >
                    Kabul et
                  </button>
                  <button
                    type="button"
                    disabled={reviewingId === report.report_id}
                    onClick={() => void handleReview(report.report_id, "rejected")}
                    className="rounded-md bg-red-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-red-700 disabled:opacity-50"
                  >
                    Reddet
                  </button>
                </>
              ) : null}
            </ReportCard>
          ))}
        </div>
      ) : null}

      {lastReview ? (
        <section className="space-y-4">
          <h2 className="text-lg font-semibold text-slate-900">
            Son İnceleme Sonucu
          </h2>
          <ReportCard report={lastReview.report} />
          {lastReview.task ? <TaskCard task={lastReview.task} /> : null}
        </section>
      ) : null}
    </div>
  );
}
