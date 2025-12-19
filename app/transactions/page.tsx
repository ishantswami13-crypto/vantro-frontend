"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import AuthGuard from "@/components/AuthGuard";
import Navbar from "@/components/Navbar";
import { apiFetch } from "@/lib/api";
import { downloadCsv } from "@/lib/exportCsv";
import { getActiveBusinessId } from "@/lib/business";
import { getToken, logout } from "@/lib/auth";

type Txn = {
  id: number | string;
  type: "income" | "expense";
  amount: number;
  note: string;
  created_at: string;
};

type Summary = {
  income: number;
  expense: number;
  net: number;
};

export default function TransactionsPage() {
  const router = useRouter();

  const [businessId, setBusinessId] = useState<number | null>(null);
  const [type, setType] = useState<"income" | "expense">("income");
  const [amount, setAmount] = useState<string>("");
  const [note, setNote] = useState<string>("");
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string>("");

  const [summary, setSummary] = useState<Summary | null>(null);
  const [txns, setTxns] = useState<Txn[]>([]);

  const canSubmit = useMemo(() => {
    const n = Number(amount);
    return Number.isFinite(n) && n > 0 && !loading && businessId !== null;
  }, [amount, loading, businessId]);

  useEffect(() => {
    if (!getToken()) router.replace("/login");
  }, [router]);

  useEffect(() => {
    const bid = getActiveBusinessId();
    if (!bid) {
      router.replace("/business");
      return;
    }
    setBusinessId(bid);
  }, [router]);

  async function refreshAll(bid?: number | null) {
    const id = bid ?? businessId;
    if (!id) {
      router.replace("/business");
      return;
    }
    const [s, list] = await Promise.all([
      apiFetch<Summary>(`/api/transactions/summary?business_id=${id}`),
      apiFetch<Txn[]>(`/api/transactions?business_id=${id}`),
    ]);
    setSummary(s);
    setTxns(list);
  }

  useEffect(() => {
    if (businessId === null) return;
    refreshAll(businessId).catch((e) => setErr(e?.message || "Failed to load"));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [businessId]);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    setLoading(true);

    const id = businessId ?? getActiveBusinessId();
    if (!id) {
      router.replace("/business");
      setLoading(false);
      return;
    }

    try {
      await apiFetch<{ id: number }>("/api/transactions", {
        method: "POST",
        body: JSON.stringify({
          type,
          amount: Number(amount),
          note: note || "",
          business_id: id,
        }),
      });

      setAmount("");
      setNote("");
      await refreshAll(id);
    } catch (e: any) {
      setErr(e?.message || "Failed to create transaction");
    } finally {
      setLoading(false);
    }
  }

  async function exportCsv() {
    const id = businessId ?? getActiveBusinessId();
    if (!id) {
      router.replace("/business");
      return;
    }

    try {
      const all = await apiFetch<Txn[]>(`/api/transactions?business_id=${id}`);
      downloadCsv("vantro-transactions.csv", all);
    } catch (e: any) {
      setErr(e?.message || "Export failed");
    }
  }

  return (
    <AuthGuard>
      <div className="min-h-screen bg-black text-white">
        <Navbar />
        <div className="max-w-4xl mx-auto p-6 space-y-6">
          <header className="flex items-center justify-between">
            <h1 className="text-2xl font-bold">Transactions</h1>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={exportCsv}
                className="px-4 py-2 rounded bg-neutral-800 hover:bg-neutral-700"
              >
                Export CSV
              </button>
              <button
                onClick={() => router.push("/business")}
                className="px-4 py-2 rounded bg-neutral-800 hover:bg-neutral-700"
              >
                Switch Business
              </button>
              <button
                onClick={() => router.push("/dashboard")}
                className="px-4 py-2 rounded bg-neutral-800 hover:bg-neutral-700"
              >
                Back to Dashboard
              </button>
              <button
                onClick={logout}
                className="px-4 py-2 rounded border border-neutral-700"
              >
                Logout
              </button>
            </div>
          </header>

          {/* Summary */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="rounded-xl bg-neutral-900 p-4">
              <div className="text-sm text-neutral-400">Income</div>
              <div className="text-xl font-semibold">
                Rs. {(summary?.income ?? 0).toLocaleString("en-IN")}
              </div>
            </div>
            <div className="rounded-xl bg-neutral-900 p-4">
              <div className="text-sm text-neutral-400">Expense</div>
              <div className="text-xl font-semibold">
                Rs. {(summary?.expense ?? 0).toLocaleString("en-IN")}
              </div>
            </div>
            <div className="rounded-xl bg-neutral-900 p-4">
              <div className="text-sm text-neutral-400">Net</div>
              <div className="text-xl font-semibold">
                Rs. {(summary?.net ?? 0).toLocaleString("en-IN")}
              </div>
            </div>
          </div>

          {/* Create form */}
          <div className="rounded-2xl bg-neutral-900 p-5">
            <h2 className="text-lg font-semibold mb-3">Add Transaction</h2>

            <form onSubmit={submit} className="space-y-4">
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={() => setType("income")}
                  className={`px-4 py-2 rounded ${
                    type === "income" ? "bg-white text-black" : "bg-neutral-800"
                  }`}
                >
                  Income
                </button>
                <button
                  type="button"
                  onClick={() => setType("expense")}
                  className={`px-4 py-2 rounded ${
                    type === "expense" ? "bg-white text-black" : "bg-neutral-800"
                  }`}
                >
                  Expense
                </button>
              </div>

              <input
                className="w-full p-2 rounded bg-neutral-800 outline-none"
                placeholder="Amount (e.g. 5000)"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                inputMode="decimal"
              />

              <input
                className="w-full p-2 rounded bg-neutral-800 outline-none"
                placeholder="Note (optional)"
                value={note}
                onChange={(e) => setNote(e.target.value)}
              />

              {err && <p className="text-red-400 text-sm">{err}</p>}

              <button
                disabled={!canSubmit}
                className="w-full bg-white text-black p-2 rounded disabled:opacity-60"
              >
                {loading ? "Saving..." : "Add"}
              </button>
            </form>
          </div>

          {/* List */}
          <div className="rounded-2xl bg-neutral-900 p-5">
            <div className="flex items-center justify-between mb-3">
              <h2 className="text-lg font-semibold">Recent</h2>
              <button
                onClick={() => refreshAll().catch((e) => setErr(e?.message || "Failed"))}
                className="px-3 py-2 rounded bg-neutral-800 hover:bg-neutral-700 text-sm"
              >
                Refresh
              </button>
            </div>

            <div className="divide-y divide-neutral-800">
              {txns.length === 0 ? (
                <div className="text-neutral-400 text-sm py-6">No transactions yet.</div>
              ) : (
                txns.map((t) => (
                  <div key={t.id} className="py-3 flex items-center justify-between">
                    <div>
                      <div className="font-medium">
                        {t.type === "income" ? "Income" : "Expense"} - Rs.{" "}
                        {Number(t.amount).toLocaleString("en-IN")}
                      </div>
                      <div className="text-xs text-neutral-400">
                        {t.note || "-"} - {new Date(t.created_at).toLocaleString()}
                      </div>
                    </div>
                    <span
                      className={`text-xs px-2 py-1 rounded ${
                        t.type === "income" ? "bg-green-900/40" : "bg-red-900/40"
                      }`}
                    >
                      {t.type}
                    </span>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      </div>
    </AuthGuard>
  );
}
