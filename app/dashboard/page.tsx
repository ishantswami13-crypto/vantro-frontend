"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import AuthGuard from "@/components/AuthGuard";
import Navbar from "@/components/Navbar";
import { apiFetch } from "@/lib/api";
import { logout } from "@/lib/auth";

type Summary = {
  income: number;
  expense: number;
  net: number;
};

type Txn = {
  id: number | string;
  type: "income" | "expense";
  amount: number;
  note?: string;
  created_at?: string;
};

function formatINR(n: number) {
  try {
    return new Intl.NumberFormat("en-IN", {
      style: "currency",
      currency: "INR",
      maximumFractionDigits: 0,
    }).format(n);
  } catch {
    return `₹${Math.round(n)}`;
  }
}

export default function DashboardPage() {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [recent, setRecent] = useState<Txn[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  async function load() {
    setErr("");
    setLoading(true);
    try {
      const [s, list] = await Promise.all([
        apiFetch<Summary>("/api/transactions/summary"),
        apiFetch<Txn[]>("/api/transactions"),
      ]);
      setSummary(s);
      setRecent(Array.isArray(list) ? list.slice(0, 8) : []);
    } catch (e: any) {
      setErr(e?.message || "Failed to load dashboard");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, []);

  const income = summary?.income ?? 0;
  const expense = summary?.expense ?? 0;
  const net = summary?.net ?? 0;

  return (
    <AuthGuard>
      <div className="min-h-screen bg-black text-white">
        <Navbar />

        <div className="max-w-4xl mx-auto p-6 space-y-6">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-bold">Dashboard</h1>
            <div className="flex gap-3">
              <Link
                href="/transactions"
                className="px-4 py-2 rounded bg-white text-black hover:opacity-90"
              >
                Add / View Transactions
              </Link>
              <button
                onClick={load}
                className="px-4 py-2 rounded border border-neutral-700 hover:border-neutral-500"
              >
                Refresh
              </button>
              <button
                onClick={logout}
                className="px-4 py-2 rounded border border-neutral-700"
              >
                Logout
              </button>
            </div>
          </div>

          {err ? (
            <div className="bg-red-900/40 border border-red-500/40 p-3 rounded">
              {err}
            </div>
          ) : null}

          {/* Summary cards */}
          <section className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="bg-neutral-900 border border-neutral-800 rounded-lg p-4">
              <div className="text-sm text-neutral-400">Income</div>
              <div className="text-xl font-semibold">
                {loading ? "..." : formatINR(income)}
              </div>
            </div>

            <div className="bg-neutral-900 border border-neutral-800 rounded-lg p-4">
              <div className="text-sm text-neutral-400">Expense</div>
              <div className="text-xl font-semibold">
                {loading ? "..." : formatINR(expense)}
              </div>
            </div>

            <div className="bg-neutral-900 border border-neutral-800 rounded-lg p-4">
              <div className="text-sm text-neutral-400">Net</div>
              <div
                className={`text-xl font-semibold ${
                  net >= 0 ? "text-green-400" : "text-red-400"
                }`}
              >
                {loading ? "..." : formatINR(net)}
              </div>
            </div>
          </section>

          {/* Recent transactions */}
          <section className="bg-neutral-900 border border-neutral-800 rounded-lg p-5">
            <div className="flex items-center justify-between mb-3">
              <h2 className="text-lg font-semibold">Recent</h2>
              <Link href="/transactions" className="text-sm text-neutral-300 hover:text-white">
                View all →
              </Link>
            </div>

            {loading ? (
              <div className="text-neutral-400">Loading...</div>
            ) : recent.length === 0 ? (
              <div className="text-neutral-400">No transactions yet.</div>
            ) : (
              <div className="divide-y divide-neutral-800">
                {recent.map((t) => (
                  <div key={String(t.id)} className="py-3 flex justify-between gap-4">
                    <div>
                      <div className="font-medium">
                        <span className={t.type === "income" ? "text-green-400" : "text-red-400"}>
                          {t.type.toUpperCase()}
                        </span>
                        <span className="text-neutral-400"> · </span>
                        <span>{formatINR(Number(t.amount || 0))}</span>
                      </div>
                      {t.note ? <div className="text-sm text-neutral-400">{t.note}</div> : null}
                    </div>

                    <div className="text-xs text-neutral-500">
                      {t.created_at ? new Date(t.created_at).toLocaleString() : ""}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>
        </div>
      </div>
    </AuthGuard>
  );
}

