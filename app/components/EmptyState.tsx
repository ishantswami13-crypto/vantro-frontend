import { ArrowUpRight } from "lucide-react";

export function EmptyState({
  title,
  subtitle,
  action,
  onAction,
}: {
  title: string;
  subtitle: string;
  action?: string;
  onAction?: () => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center rounded-2xl border border-white/10 bg-white/[0.02] py-14 text-center">
      <div className="mb-3 h-10 w-10 rounded-2xl border border-white/10 bg-white/[0.06]" />
      <div className="text-sm font-semibold tracking-tight">{title}</div>
      <div className="mt-1 max-w-xs text-xs text-white/50">{subtitle}</div>

      {action && (
        <button
          onClick={onAction}
          className="mt-5 inline-flex items-center gap-1 rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-xs hover:bg-white/10 transition"
        >
          {action}
          <ArrowUpRight size={12} />
        </button>
      )}
    </div>
  );
}
