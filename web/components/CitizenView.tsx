"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { apiFetch } from "@/lib/api";
import type { Report, UserRole } from "@/app/types";
import ReportCard from "./ReportCard";

// Istanbul center — fallback only when the browser cannot provide a location.
const DEFAULT_LAT = "41.0082";
const DEFAULT_LNG = "28.9784";

type LocStatus = "idle" | "loading" | "granted" | "denied";

type CitizenViewProps = {
  role: UserRole;
  sourceType: "citizen_mobile" | "staff_mobile";
  title: string;
  description: string;
};

export default function CitizenView({
  role,
  sourceType,
  title,
  description,
}: CitizenViewProps) {
  const [formDescription, setFormDescription] = useState("");
  const [lat, setLat] = useState("");
  const [lng, setLng] = useState("");
  const [locStatus, setLocStatus] = useState<LocStatus>("idle");
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdReport, setCreatedReport] = useState<Report | null>(null);
  const blobUrlRef = useRef<string | null>(null);

  const revokeBlobUrl = useCallback(() => {
    if (blobUrlRef.current) {
      URL.revokeObjectURL(blobUrlRef.current);
      blobUrlRef.current = null;
    }
  }, []);

  const requestLocation = useCallback(() => {
    if (typeof navigator === "undefined" || !navigator.geolocation) {
      setLocStatus("denied");
      return;
    }
    setLocStatus("loading");
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setLat(pos.coords.latitude.toFixed(6));
        setLng(pos.coords.longitude.toFixed(6));
        setLocStatus("granted");
      },
      () => setLocStatus("denied"),
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 0 },
    );
  }, []);

  // Ask for location on mount (deferred); the user can retry with the button.
  useEffect(() => {
    const timer = window.setTimeout(() => {
      requestLocation();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [requestLocation]);

  const handleImageChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0] ?? null;
      revokeBlobUrl();
      if (file) {
        const url = URL.createObjectURL(file);
        blobUrlRef.current = url;
        setImagePreview(url);
        setImageFile(file);
      } else {
        setImagePreview(null);
        setImageFile(null);
      }
      e.target.value = "";
    },
    [revokeBlobUrl],
  );

  const handleSubmit = useCallback(
    async (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      if (!imageFile) return;

      setSubmitting(true);
      setError(null);
      setCreatedReport(null);

      try {
        const body = new FormData();
        body.append("image", imageFile);
        body.append("description", formDescription);
        body.append("lat", lat || DEFAULT_LAT);
        body.append("lng", lng || DEFAULT_LNG);
        body.append("source_type", sourceType);

        const res = await apiFetch("/api/v1/reports", role, {
          method: "POST",
          body,
        });

        if (!res.ok) {
          const text = await res.text();
          throw new Error(text || `Bildirim gönderilemedi (${res.status})`);
        }

        const report: Report = await res.json();
        setCreatedReport(report);
        setFormDescription("");
        setImageFile(null);
        revokeBlobUrl();
        setImagePreview(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Bilinmeyen hata");
      } finally {
        setSubmitting(false);
      }
    },
    [formDescription, imageFile, lat, lng, revokeBlobUrl, role, sourceType],
  );

  return (
    <div className="space-y-6">
      <section className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
        <h2 className="mb-1 text-lg font-semibold text-slate-900">{title}</h2>
        <p className="mb-5 text-sm text-slate-600">{description}</p>

        <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              Fotoğraf <span className="text-red-500">*</span>
            </label>
            <div className="flex flex-wrap items-center gap-3">
              <label className="cursor-pointer rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm transition hover:bg-slate-50">
                {imageFile ? "Fotoğrafı değiştir" : "Fotoğraf seç"}
                <input
                  type="file"
                  accept="image/*"
                  className="sr-only"
                  disabled={submitting}
                  onChange={handleImageChange}
                />
              </label>
              {imageFile ? (
                <span className="max-w-xs truncate text-sm text-slate-500">
                  {imageFile.name}
                </span>
              ) : null}
            </div>
            {imagePreview ? (
              <div className="mt-3">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={imagePreview}
                  alt="Seçilen fotoğraf"
                  className="h-40 w-auto rounded-lg border border-slate-200 object-cover shadow-sm"
                />
              </div>
            ) : null}
          </div>

          <div>
            <label
              htmlFor="report-description"
              className="mb-1.5 block text-sm font-medium text-slate-700"
            >
              Açıklama <span className="text-red-500">*</span>
            </label>
            <textarea
              id="report-description"
              rows={3}
              required
              disabled={submitting}
              value={formDescription}
              onChange={(e) => setFormDescription(e.target.value)}
              placeholder="Sorunu kısaca açıklayın…"
              className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 placeholder-slate-400 shadow-sm outline-none transition focus:border-slate-500 focus:ring-2 focus:ring-slate-200 disabled:opacity-50"
            />
          </div>

          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              Konum
            </label>
            <div className="flex flex-wrap items-center gap-3">
              <button
                type="button"
                onClick={requestLocation}
                disabled={submitting || locStatus === "loading"}
                className="inline-flex items-center gap-2 rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm transition hover:bg-slate-50 disabled:opacity-50"
              >
                {locStatus === "loading" ? "Konum alınıyor…" : "📍 Konumumu kullan"}
              </button>
              {locStatus === "granted" ? (
                <span className="inline-flex items-center rounded-full bg-emerald-100 px-3 py-1 text-xs font-medium text-emerald-800">
                  Konum alındı ({lat}, {lng})
                </span>
              ) : null}
              {locStatus === "denied" ? (
                <span className="inline-flex items-center rounded-full bg-amber-100 px-3 py-1 text-xs font-medium text-amber-800">
                  Konum izni yok — İstanbul merkez kullanılacak
                </span>
              ) : null}
            </div>
          </div>

          <button
            type="submit"
            disabled={submitting || !formDescription.trim() || !imageFile}
            className="rounded-lg bg-slate-900 px-5 py-2.5 text-sm font-medium text-white shadow-sm transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {submitting ? "Gönderiliyor…" : "Bildir"}
          </button>

          {error ? (
            <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-800">
              {error}
            </p>
          ) : null}
        </form>
      </section>

      {createdReport ? (
        <section>
          <h2 className="mb-4 text-lg font-semibold text-slate-900">
            Oluşturulan Bildirim
          </h2>
          <ReportCard report={createdReport} />
        </section>
      ) : null}
    </div>
  );
}
