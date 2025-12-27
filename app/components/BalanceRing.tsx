"use client";

import { motion } from "framer-motion";

export function BalanceRing({
  income,
  expense,
}: {
  income: number;
  expense: number;
}) {
  const total = income - expense;
  const max = Math.max(income, expense, 1);
  const progress = Math.min(Math.abs(total) / max, 1);

  const size = 140;
  const stroke = 10;
  const radius = (size - stroke) / 2;
  const circumference = 2 * Math.PI * radius;

  return (
    <div className="flex items-center gap-6">
      <svg width={size} height={size} className="-rotate-90">
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          stroke="rgba(255,255,255,0.08)"
          strokeWidth={stroke}
          fill="none"
        />
        <motion.circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          stroke={total >= 0 ? "#34d399" : "#f87171"}
          strokeWidth={stroke}
          fill="none"
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={circumference}
          animate={{
            strokeDashoffset: circumference * (1 - progress),
          }}
          transition={{ duration: 1.1, ease: "easeOut" }}
        />
      </svg>

      <div>
        <div className="text-xs text-white/50">Net balance</div>
        <div className="mt-1 text-2xl font-semibold tabular-nums tracking-tight">
          ₹{total.toLocaleString()}
        </div>
        <div className="mt-1 text-xs text-white/40">
          {total >= 0 ? "Positive cash flow" : "Overspending"}
        </div>
      </div>
    </div>
  );
}
