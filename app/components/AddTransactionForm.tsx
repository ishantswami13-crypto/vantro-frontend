"use client";

import { useMemo, useState } from "react";
import { cn } from "@/lib/utils";

export default function AddTransactionForm({
  onDone,
}: {
  onDone?: () => void;
}) {
  const [type, setType] = useState<"income" | "expense">("expense");
  const [title, setTitle] = useState("");
  const [category, setCategory] = useState("General");
  const [amount, setAmount] = useState("");

  const valid = useMemo(() => {
    const a = Number(amount);
    return title.trim().length >= 2 && category.trim().length >= 2 && a > 0;
  }, [title, category, amount]);

  return (
    <div className="space-y-4">
      {/* Type toggle */}
      <div className="flex gap-2 rounded-2xl border border-white/10 bg-white/[0.03] p-1">
        <button
          type="button"
          onClick={() => setType("expense")}
          className={cn(
            "flex-1 rounded-xl px-3 py-2 text-sm transition",
            type === "expense" ? "bg-white/10" : "hover:bg-white/5 text-white/70"
          )}
        >
          Expense
        </button>
        <button
          type="button"
          onClick={() => setType("income")}
          className={cn(
            "flex-1 rounded-xl px-3 py-2 text-sm transition",
            type === "income" ? "bg-white/10" : "hover:bg-white/5 text-white/70"
          )}
        >
          Income
        </button>
      </div>

      {/* Inputs */}
      <div className="grid grid-cols-1 gap-3">
        <Field label="Title">
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g. Coffee, Client payment"
            className="w-full rounded-xl border border-white/10 bg-black/40 px-3 py-2 text-sm outline-none focus:border-white/20"
          />
        </Field>

        <div className="grid grid-cols-2 gap-3">
          <Field label="Category">
            <input
              value={category}
              onChange={(e) => setCategory(e.target.value)}
              placeholder="e.g. Food, Sales"
              className="w-full rounded-xl border border-white/10 bg-black/40 px-3 py-2 text-sm outline-none focus:border-white/20"
            />
          </Field>

          <Field label="Amount (₹)">
            <input
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              inputMode="numeric"
              placeholder="0"
              className="w-full rounded-xl border border-white/10 bg-black/40 px-3 py-2 text-sm outline-none focus:border-white/20"
            />
          </Field>
        </div>
      </div>

      {/* Actions */}
      <div className="flex items-center justify-between pt-1">
        <div className="text-xs text-white/50">
          Stored locally for now — API wiring next.
        </div>

        <button
          type="button"
          disabled={!valid}
          onClick={() => {
            // for now: log it (next step: call API + update UI)
            console.log({ type, title, category, amount: Number(amount) });
            onDone?.();
          }}
          className={cn(
            "rounded-xl px-4 py-2 text-sm font-medium transition",
            valid
              ? "bg-white text-black hover:opacity-90"
              : "bg-white/10 text-white/40 cursor-not-allowed"
          )}
        >
          Save
        </button>
      </div>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1">
      <div className="text-xs text-white/60">{label}</div>
      {children}
    </div>
  );
}
