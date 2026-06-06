"use client";

import { useCallback, useRef, useState } from "react";
import { apiFetch } from "@/lib/api";
import type { Report, UserRole } from "@/app/types";
import ReportCard from "./ReportCard";

const DEFAULT_LAT = "41.0082";
const DEFAULT_LNG = "28.9784";

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
  const [lat, setLat] = useState(DEFAULT_LAT);
  const [lng, setLng] = useState(DEFAULT_LNG);
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
        body.append("lat", lat);
        body.append("lng", lng);
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
        setLat(DEFAULT_LAT);
        setLng(DEFAULT_LNG);
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

          <div className="flex flex-wrap gap-3">
            <div className="min-w-[140px] flex-1">
              <label
                htmlFor="report-lat"
                className="mb-1.5 block text-sm font-medium text-slate-700"
              >
                Enlem
              </label>
              <input
                id="report-lat"
                type="text"
                inputMode="decimal"
                disabled={submitting}
                value={lat}
                onChange={(e) => setLat(e.target.value)}
                className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-slate-500 focus:ring-2 focus:ring-slate-200 disabled:opacity-50"
              />
            </div>
            <div className="min-w-[140px] flex-1">
              <label
                htmlFor="report-lng"
                className="mb-1.5 block text-sm font-medium text-slate-700"
              >
                Boylam
              </label>
              <input
                id="report-lng"
                type="text"
                inputMode="decimal"
                disabled={submitting}
                value={lng}
                onChange={(e) => setLng(e.target.value)}
                className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-slate-500 focus:ring-2 focus:ring-slate-200 disabled:opacity-50"
              />
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
