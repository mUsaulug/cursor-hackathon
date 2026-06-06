const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const PRIVACY_CHECKLIST = [
  "No raw camera images in git",
  "Faces and license plates are not stored",
  "Anonymize before persistence or transmission",
  "Raw folders (data/raw/, uploads/raw/, private/) stay empty",
  "Deletion flow documented in docs/PRIVACY.md",
] as const;

const PROJECT_MODULES = [
  {
    name: "Web",
    stack: "Next.js",
    folder: "web/",
    command: "npm run dev",
    port: "localhost:3000",
  },
  {
    name: "Mobile",
    stack: "Expo",
    folder: "mobile/",
    command: "npx expo start",
    port: "Expo dev server",
  },
  {
    name: "Backend",
    stack: "Go",
    folder: "backend/",
    command: "go run main.go",
    port: "localhost:8080",
  },
] as const;

async function getHealth(): Promise<{ status: string } | null> {
  try {
    const res = await fetch(`${API_URL}/health`, { cache: "no-store" });
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

function StatusBadge({
  ok,
  okLabel,
  failLabel,
}: {
  ok: boolean;
  okLabel: string;
  failLabel: string;
}) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${
        ok
          ? "bg-emerald-100 text-emerald-800"
          : "bg-amber-100 text-amber-800"
      }`}
    >
      {ok ? okLabel : failLabel}
    </span>
  );
}

export default async function Home() {
  const health = await getHealth();
  const backendOk = health?.status === "ok";

  const demoChecks = [
    { label: "Web app serving", ok: true },
    { label: "Backend /health reachable", ok: backendOk },
    { label: "API URL configured", ok: Boolean(API_URL) },
    { label: "Privacy docs in repo", ok: true },
  ];

  const demoReady = demoChecks.every((item) => item.ok);

  return (
    <div className="mx-auto min-h-screen max-w-6xl px-4 py-10 sm:px-6 lg:px-8">
      <header className="mb-10">
        <p className="text-sm font-semibold uppercase tracking-wider text-slate-500">
          Urban AI Hackathon
        </p>
        <h1 className="mt-2 text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl">
          Cursor Hackathon Istanbul
        </h1>
        <p className="mt-3 max-w-2xl text-base leading-relaxed text-slate-600">
          Minimal dashboard for local demo readiness. No raw imagery, faces, or
          plates — see{" "}
          <code className="rounded bg-slate-200 px-1.5 py-0.5 font-mono text-sm text-slate-800">
            docs/PRIVACY.md
          </code>
          .
        </p>
      </header>

      <div className="grid gap-6 sm:grid-cols-2">
        {/* Backend health */}
        <section className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="mb-4 flex items-center justify-between gap-3">
            <h2 className="text-lg font-semibold text-slate-900">
              Backend health
            </h2>
            <StatusBadge
              ok={backendOk}
              okLabel="Online"
              failLabel="Offline"
            />
          </div>
          <dl className="space-y-3 text-sm">
            <div>
              <dt className="text-slate-500">API URL</dt>
              <dd className="mt-0.5 font-mono text-slate-800">{API_URL}</dd>
            </div>
            <div>
              <dt className="text-slate-500">Health endpoint</dt>
              <dd className="mt-0.5 font-mono text-slate-800">GET /health</dd>
            </div>
            <div>
              <dt className="text-slate-500">Response</dt>
              <dd className="mt-0.5 font-mono text-slate-800">
                {backendOk ? '{"status":"ok"}' : "unreachable"}
              </dd>
            </div>
          </dl>
          {!backendOk && (
            <p className="mt-4 rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-900">
              Start the API:{" "}
              <code className="font-mono">cd backend && go run main.go</code>
            </p>
          )}
        </section>

        {/* Demo readiness */}
        <section className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="mb-4 flex items-center justify-between gap-3">
            <h2 className="text-lg font-semibold text-slate-900">
              Demo readiness
            </h2>
            <StatusBadge
              ok={demoReady}
              okLabel="Ready"
              failLabel="Pending"
            />
          </div>
          <ul className="space-y-2.5">
            {demoChecks.map((item) => (
              <li
                key={item.label}
                className="flex items-center justify-between gap-3 text-sm"
              >
                <span className="text-slate-700">{item.label}</span>
                <span
                  className={
                    item.ok ? "text-emerald-600" : "text-amber-600"
                  }
                  aria-hidden
                >
                  {item.ok ? "✓" : "○"}
                </span>
              </li>
            ))}
          </ul>
        </section>

        {/* Privacy / KVKK */}
        <section className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm sm:col-span-2 lg:col-span-1">
          <h2 className="mb-4 text-lg font-semibold text-slate-900">
            Privacy / KVKK checklist
          </h2>
          <ul className="space-y-2.5">
            {PRIVACY_CHECKLIST.map((item) => (
              <li
                key={item}
                className="flex items-start gap-2.5 text-sm text-slate-700"
              >
                <span className="mt-0.5 text-emerald-600" aria-hidden>
                  ✓
                </span>
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </section>

        {/* Project modules */}
        <section className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm sm:col-span-2 lg:col-span-1">
          <h2 className="mb-4 text-lg font-semibold text-slate-900">
            Project modules
          </h2>
          <div className="space-y-4">
            {PROJECT_MODULES.map((mod) => (
              <article
                key={mod.name}
                className="rounded-lg border border-slate-100 bg-slate-50 p-4"
              >
                <div className="flex items-center justify-between gap-2">
                  <h3 className="font-semibold text-slate-900">{mod.name}</h3>
                  <span className="text-xs font-medium text-slate-500">
                    {mod.stack}
                  </span>
                </div>
                <dl className="mt-2 space-y-1 text-sm">
                  <div className="flex gap-2">
                    <dt className="text-slate-500">Folder</dt>
                    <dd className="font-mono text-slate-800">{mod.folder}</dd>
                  </div>
                  <div className="flex gap-2">
                    <dt className="text-slate-500">Run</dt>
                    <dd className="font-mono text-slate-800">{mod.command}</dd>
                  </div>
                  <div className="flex gap-2">
                    <dt className="text-slate-500">Target</dt>
                    <dd className="font-mono text-slate-800">{mod.port}</dd>
                  </div>
                </dl>
              </article>
            ))}
          </div>
        </section>
      </div>
    </div>
  );
}
