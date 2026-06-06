"use client";

import { useState } from "react";
import type { UserRole } from "@/app/types";
import AiAnalysisDashboard from "@/components/AiAnalysisDashboard";
import CitizenView from "@/components/CitizenView";
import FieldStaffView from "@/components/FieldStaffView";
import ManagerView from "@/components/ManagerView";
import OperatorView from "@/components/OperatorView";

type ConsoleView = "operations" | "ai";

const ROLE_OPTIONS: { value: UserRole; label: string }[] = [
  { value: "citizen", label: "Vatandaş" },
  { value: "field_staff", label: "Saha" },
  { value: "operator", label: "Operatör" },
  { value: "manager", label: "Yönetici" },
];

function RoleOperationsView({ role }: { role: UserRole }) {
  switch (role) {
    case "citizen":
      return (
        <CitizenView
          role={role}
          sourceType="citizen_mobile"
          title="Vatandaş Bildirimi"
          description="Kentsel sorunları fotoğraf ve konum bilgisiyle bildirin."
        />
      );
    case "field_staff":
      return (
        <>
          <CitizenView
            role={role}
            sourceType="staff_mobile"
            title="Saha Bildirimi"
            description="Saha ekibi olarak tespit ettiğiniz sorunları bildirin."
          />
          <FieldStaffView role={role} />
        </>
      );
    case "operator":
      return <OperatorView role={role} />;
    case "manager":
      return <ManagerView role={role} />;
  }
}

export default function Home() {
  const [role, setRole] = useState<UserRole>("citizen");
  const [view, setView] = useState<ConsoleView>("operations");

  return (
    <div className="mx-auto min-h-screen max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <header className="mb-8">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl">
              CivicLens
            </h1>
            <p className="mt-2 max-w-3xl text-base leading-relaxed text-slate-600">
              Belediye operasyon konsolu — rol tabanlı bildirim, inceleme ve
              analitik paneli.
            </p>
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="role-select"
              className="text-xs font-semibold uppercase tracking-wider text-slate-500"
            >
              Rol
            </label>
            <select
              id="role-select"
              value={role}
              onChange={(e) => setRole(e.target.value as UserRole)}
              className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm font-medium text-slate-900 shadow-sm outline-none transition focus:border-slate-500 focus:ring-2 focus:ring-slate-200"
            >
              {ROLE_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
        </div>

        <nav className="mt-6 flex gap-2 border-b border-slate-200">
          <button
            type="button"
            onClick={() => setView("operations")}
            className={`border-b-2 px-4 py-2 text-sm font-medium transition ${
              view === "operations"
                ? "border-slate-900 text-slate-900"
                : "border-transparent text-slate-500 hover:text-slate-700"
            }`}
          >
            Operasyon
          </button>
          <button
            type="button"
            onClick={() => setView("ai")}
            className={`border-b-2 px-4 py-2 text-sm font-medium transition ${
              view === "ai"
                ? "border-slate-900 text-slate-900"
                : "border-transparent text-slate-500 hover:text-slate-700"
            }`}
          >
            AI Analiz
          </button>
        </nav>
      </header>

      {view === "operations" ? (
        <RoleOperationsView role={role} />
      ) : (
        <AiAnalysisDashboard role={role} />
      )}
    </div>
  );
}
