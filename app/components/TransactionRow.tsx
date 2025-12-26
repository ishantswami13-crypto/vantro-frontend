import { ArrowDownLeft, ArrowUpRight } from "lucide-react";
import { cn } from "@/lib/utils";

export function TransactionRow({
  title,
  category,
  amount,
  type,
}: {
  title: string;
  category: string;
  amount: number;
  type: "income" | "expense";
}) {
  const positive = type === "income";

  return (
    <div className="flex items-center justify-between rounded-xl border border-white/10 bg-white/[0.02] px-4 py-3 hover:bg-white/[0.04] transition">
      <div className="flex items-center gap-3">
        <div
          className={cn(
            "flex h-9 w-9 items-center justify-center rounded-xl",
            positive ? "bg-emerald-500/15 text-emerald-400" : "bg-red-500/15 text-red-400"
          )}
        >
          {positive ? <ArrowDownLeft size={16} /> : <ArrowUpRight size={16} />}
        </div>

        <div>
          <div className="text-sm font-medium">{title}</div>
          <div className="text-xs text-white/50">{category}</div>
        </div>
      </div>

      <div
        className={cn(
          "text-sm font-semibold tabular-nums",
          positive ? "text-emerald-400" : "text-red-400"
        )}
      >
        {positive ? "+" : "-"}₹{amount}
      </div>
    </div>
  );
}
