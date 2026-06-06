import type { Priority } from "@/app/types";

export const PRIORITY_STYLES: Record<Priority, string> = {
  critical: "bg-red-100 text-red-800",
  high: "bg-orange-100 text-orange-800",
  medium: "bg-blue-100 text-blue-800",
  low: "bg-slate-100 text-slate-700",
};

export const PRIORITY_LABELS: Record<Priority, string> = {
  critical: "Kritik",
  high: "Yüksek",
  medium: "Orta",
  low: "Düşük",
};

const STATUS_STYLES: Record<string, string> = {
  waiting_for_review: "bg-amber-100 text-amber-800",
  task_created: "bg-blue-100 text-blue-800",
  merged: "bg-purple-100 text-purple-800",
  rejected: "bg-red-100 text-red-800",
  created: "bg-slate-100 text-slate-700",
  open: "bg-blue-100 text-blue-800",
  in_progress: "bg-indigo-100 text-indigo-800",
  completed: "bg-emerald-100 text-emerald-800",
  cancelled: "bg-slate-200 text-slate-600",
};

const STATUS_LABELS: Record<string, string> = {
  waiting_for_review: "İnceleme Bekliyor",
  task_created: "Görev Oluşturuldu",
  merged: "Birleştirildi",
  rejected: "Reddedildi",
  created: "Oluşturuldu",
  open: "Açık",
  in_progress: "Devam Ediyor",
  completed: "Tamamlandı",
  cancelled: "İptal",
};

type BadgeProps = {
  className: string;
  children: React.ReactNode;
};

export function Badge({ className, children }: BadgeProps) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${className}`}
    >
      {children}
    </span>
  );
}

export function PriorityBadge({ priority }: { priority: string }) {
  const style =
    PRIORITY_STYLES[priority as Priority] ?? "bg-slate-100 text-slate-700";
  const label =
    PRIORITY_LABELS[priority as Priority] ?? priority.replace(/_/g, " ");
  return <Badge className={style}>{label}</Badge>;
}

export function StatusBadge({ status }: { status: string }) {
  const style = STATUS_STYLES[status] ?? "bg-slate-100 text-slate-700";
  const label = STATUS_LABELS[status] ?? status.replace(/_/g, " ");
  return <Badge className={style}>{label}</Badge>;
}

export function formatDate(iso: string): string {
  try {
    return new Intl.DateTimeFormat("tr-TR", {
      day: "2-digit",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    }).format(new Date(iso));
  } catch {
    return iso;
  }
}
