"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import DetectionCard, { formatObjectType } from "@/components/DetectionCard";
import type {
  AnalysisResult,
  MaintenanceReport,
  Priority,
  SummaryResponse,
} from "./types";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const PRIORITY_BORDER: Record<Priority, string> = {
  critical: "border-red-500",
  high: "border-orange-500",
  medium: "border-blue-500",
  low: "border-slate-400",
};

type DemoScene = {
  label: string;
  imagePath: string;
  sourceRef: string;
  mode?: string;
};

const DEMO_SCENES: DemoScene[] = [
  {
    label: "Trafik sahnesi",
    imagePath: "/samples/street_traffic_01.webp",
    sourceRef: "sample_street_traffic",
  },
  {
    label: "Yol hasarı",
    imagePath: "/samples/road_pothole_01.webp",
    sourceRef: "sample_road_damage",
    mode: "road_damage",
  },
];

function BoolBadge({ ok, okLabel, failLabel }: { ok: boolean; okLabel: string; failLabel: string }) {
  return (
    <span
      className={`inline-flex rounded-full px-2 py-0.5 text-xs font-semibold ${
        ok ? "bg-emerald-100 text-emerald-800" : "bg-amber-100 text-amber-800"
      }`}
    >
      {ok ? okLabel : failLabel}
    </span>
  );
}

export default function Home() {
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [result, setResult] = useState<AnalysisResult | null>(null);
  const [summary, setSummary] = useState<MaintenanceReport | null>(null);
  const [loading, setLoading] = useState(false);
  const [reviewingId, setReviewingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const blobUrlRef = useRef<string | null>(null);

  const revokeBlobUrl = useCallback(() => {
    if (blobUrlRef.current) {
      URL.revokeObjectURL(blobUrlRef.current);
      blobUrlRef.current = null;
    }
  }, []);

  useEffect(() => {
    return () => revokeBlobUrl();
  }, [revokeBlobUrl]);

  const setPreview = useCallback(
    (src: string, isBlob = false) => {
      revokeBlobUrl();
      if (isBlob) {
        blobUrlRef.current = src;
      }
      setImagePreview(src);
    },
    [revokeBlobUrl],
  );

  const fetchSummary = useCallback(async () => {
    const res = await fetch(`${API_URL}/api/v1/vision/summary`, {
      cache: "no-store",
    });
    if (!res.ok) {
      throw new Error(`Özet alınamadı (${res.status})`);
    }
    const data: SummaryResponse = await res.json();
    setSummary(data.report);
  }, []);

  const runAnalyze = useCallback(
    async (url: string, previewSrc: string) => {
      setLoading(true);
      setError(null);
      setPreview(previewSrc);
      setResult(null);
      setSummary(null);

      try {
        const res = await fetch(url, { method: "POST" });
        if (!res.ok) {
          const text = await res.text();
          throw new Error(text || `Analiz başarısız (${res.status})`);
        }
        const data: AnalysisResult = await res.json();
        setResult(data);
        await fetchSummary();
      } catch (err) {
        setError(err instanceof Error ? err.message : "Bilinmeyen hata");
      } finally {
        setLoading(false);
      }
    },
    [fetchSummary, setPreview],
  );

  const handleDemo = useCallback(
    (scene: DemoScene) => {
      const params = new URLSearchParams({ source_ref: scene.sourceRef });
      if (scene.mode) params.set("mode", scene.mode);
      void runAnalyze(
        `${API_URL}/api/v1/vision/analyze?${params.toString()}`,
        scene.imagePath,
      );
    },
    [runAnalyze],
  );

  const handleUpload = useCallback(
    (file: File) => {
      const preview = URL.createObjectURL(file);
      const form = new FormData();
      form.append("image", file);
      void (async () => {
        setLoading(true);
        setError(null);
        setPreview(preview, true);
        setResult(null);
        setSummary(null);

        try {
          const res = await fetch(`${API_URL}/api/v1/vision/analyze`, {
            method: "POST",
            body: form,
          });
          if (!res.ok) {
            const text = await res.text();
            throw new Error(text || `Yükleme başarısız (${res.status})`);
          }
          const data: AnalysisResult = await res.json();
          setResult(data);
          await fetchSummary();
        } catch (err) {
          setError(err instanceof Error ? err.message : "Bilinmeyen hata");
        } finally {
          setLoading(false);
        }
      })();
    },
    [fetchSummary, setPreview],
  );

  const handleReview = useCallback(
    async (detectionId: string, decision: "accepted" | "rejected") => {
      setReviewingId(detectionId);
      setError(null);

      try {
        const res = await fetch(
          `${API_URL}/api/v1/vision/reviews/${detectionId}`,
          {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              decision,
              reviewed_by: "web_dashboard",
            }),
          },
        );
        if (!res.ok) {
          const text = await res.text();
          throw new Error(text || `İnceleme başarısız (${res.status})`);
        }
        const updated: AnalysisResult = await res.json();
        setResult(updated);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Bilinmeyen hata");
      } finally {
        setReviewingId(null);
      }
    },
    [],
  );

  const isPrecomputed =
    result?.model_mode === "precomputed" ||
    result?.model_id.includes("precomputed");

  return (
    <div className="mx-auto min-h-screen max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <header className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl">
          CivicLens Core
        </h1>
        <p className="mt-2 max-w-3xl text-base leading-relaxed text-slate-600">
          KVKK uyumlu kentsel bakım triyaj paneli — yüz ve plaka verisi
          saklanmaz; tespitler yalnızca cansız kentsel nesneleri hedefler.
        </p>
      </header>

      {/* Controls */}
      <section className="mb-6 rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
        <h2 className="mb-4 text-sm font-semibold uppercase tracking-wider text-slate-500">
          Analiz kaynağı
        </h2>
        <div className="flex flex-wrap items-center gap-3">
          {DEMO_SCENES.map((scene) => (
            <button
              key={scene.sourceRef}
              type="button"
              disabled={loading}
              onClick={() => handleDemo(scene)}
              className="rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {scene.label}
            </button>
          ))}
          <label className="cursor-pointer rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm transition hover:bg-slate-50 has-disabled:cursor-not-allowed has-disabled:opacity-50">
            Görüntü yükle
            <input
              type="file"
              accept="image/*"
              className="sr-only"
              disabled={loading}
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (file) handleUpload(file);
                e.target.value = "";
              }}
            />
          </label>
        </div>
        {loading && (
          <p className="mt-4 text-sm text-slate-500">Analiz çalışıyor…</p>
        )}
        {error && (
          <p className="mt-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-800">
            {error}
          </p>
        )}
      </section>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Image viewer */}
        <section className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
          <h2 className="mb-4 text-lg font-semibold text-slate-900">
            Görüntü ve tespitler
          </h2>
          {imagePreview ? (
            <div className="relative inline-block max-w-full overflow-hidden rounded-lg border border-slate-100">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={imagePreview}
                alt="Analiz görüntüsü"
                className="block h-auto max-w-full"
              />
              {result?.detections.map((d) => (
                <div
                  key={d.id}
                  className={`absolute border-2 ${PRIORITY_BORDER[d.priority]}`}
                  style={{
                    left: `${(d.bbox.xmin / result.image_width) * 100}%`,
                    top: `${(d.bbox.ymin / result.image_height) * 100}%`,
                    width: `${((d.bbox.xmax - d.bbox.xmin) / result.image_width) * 100}%`,
                    height: `${((d.bbox.ymax - d.bbox.ymin) / result.image_height) * 100}%`,
                  }}
                >
                  <span className="absolute -top-5 left-0 whitespace-nowrap rounded bg-black/75 px-1.5 py-0.5 text-[10px] font-medium text-white">
                    {formatObjectType(d.normalized_object_type)}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-slate-500">
              Demo sahne seçin veya bir görüntü yükleyin.
            </p>
          )}

          {result && (
            <div className="mt-4 flex flex-wrap items-center gap-2">
              <span className="rounded-full bg-indigo-100 px-2.5 py-0.5 text-xs font-semibold text-indigo-800">
                Mod: {result.model_mode}
              </span>
              <span className="rounded-full bg-slate-100 px-2.5 py-0.5 font-mono text-xs text-slate-700">
                {result.model_id}
              </span>
              {isPrecomputed && (
                <span className="rounded-full bg-amber-100 px-2.5 py-0.5 text-xs font-medium text-amber-800">
                  Önceden hesaplanmış demo yolu
                </span>
              )}
            </div>
          )}
        </section>

        {/* Side panels */}
        <div className="space-y-6">
          {/* KVKK privacy */}
          {result && (
            <section className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
              <h2 className="mb-4 text-lg font-semibold text-slate-900">
                KVKK / Gizlilik
              </h2>
              <dl className="space-y-3 text-sm">
                <div className="flex items-center justify-between gap-3">
                  <dt className="text-slate-600">KVKK güvenli</dt>
                  <dd>
                    <BoolBadge
                      ok={result.privacy.kvkk_safe}
                      okLabel="Evet"
                      failLabel="Hayır"
                    />
                  </dd>
                </div>
                <div className="flex items-center justify-between gap-3">
                  <dt className="text-slate-600">Ham görüntü saklandı</dt>
                  <dd>
                    <BoolBadge
                      ok={!result.privacy.raw_image_stored}
                      okLabel="Hayır"
                      failLabel="Evet"
                    />
                  </dd>
                </div>
                <div className="flex items-center justify-between gap-3">
                  <dt className="text-slate-600">Anonimleştirildi</dt>
                  <dd>
                    <BoolBadge
                      ok={result.privacy.anonymized}
                      okLabel="Evet"
                      failLabel="Hayır"
                    />
                  </dd>
                </div>
                <div className="flex items-center justify-between gap-3">
                  <dt className="text-slate-600">Engellenen PII</dt>
                  <dd className="font-medium text-slate-900">
                    {result.privacy.blocked_count}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-600">PII stratejisi</dt>
                  <dd className="mt-1 text-slate-800">
                    {result.privacy.pii_strategy}
                  </dd>
                </div>
              </dl>
            </section>
          )}

          {/* Summary */}
          <section className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
            <h2 className="mb-4 text-lg font-semibold text-slate-900">
              Bakım özeti
            </h2>
            {summary ? (
              <dl className="space-y-3 text-sm">
                <div>
                  <dt className="font-medium text-slate-500">Özet</dt>
                  <dd className="mt-1 text-slate-800">{summary.summary}</dd>
                </div>
                <div>
                  <dt className="font-medium text-slate-500">
                    Önerilen eylem
                  </dt>
                  <dd className="mt-1 text-slate-800">
                    {summary.recommended_action}
                  </dd>
                </div>
                <div className="flex items-center gap-2">
                  <dt className="font-medium text-slate-500">Risk</dt>
                  <dd>
                    <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-semibold uppercase text-slate-800">
                      {summary.risk_level}
                    </span>
                  </dd>
                </div>
                <div>
                  <dt className="font-medium text-slate-500">KVKK notu</dt>
                  <dd className="mt-1 text-slate-700">{summary.kvkk_note}</dd>
                </div>
              </dl>
            ) : result ? (
              <p className="text-sm text-slate-500">
                Bu analiz için özet raporu üretilmedi.
              </p>
            ) : (
              <p className="text-sm text-slate-500">
                Analiz sonrası özet burada görünecek.
              </p>
            )}
          </section>
        </div>
      </div>

      {/* Detections list */}
      {result && result.detections.length > 0 && (
        <section className="mt-6">
          <h2 className="mb-4 text-lg font-semibold text-slate-900">
            Tespitler ({result.detections.length})
          </h2>
          <div className="grid gap-4 sm:grid-cols-2">
            {result.detections.map((detection) => (
              <DetectionCard key={detection.id} detection={detection}>
                {detection.review_status === "needs_review" && (
                  <>
                    <button
                      type="button"
                      disabled={reviewingId === detection.id}
                      onClick={() => handleReview(detection.id, "accepted")}
                      className="rounded-md bg-emerald-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-emerald-700 disabled:opacity-50"
                    >
                      Onayla
                    </button>
                    <button
                      type="button"
                      disabled={reviewingId === detection.id}
                      onClick={() => handleReview(detection.id, "rejected")}
                      className="rounded-md bg-slate-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-slate-700 disabled:opacity-50"
                    >
                      Reddet
                    </button>
                  </>
                )}
              </DetectionCard>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
