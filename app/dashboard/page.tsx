"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Sidebar from "../components/Sidebar";
import Navbar from "../components/Navbar";

function toCSV(rows: Record<string, any>[]) {
  if (!rows || rows.length === 0) return "";
  const headers = Object.keys(rows[0]);

  const esc = (v: any) => {
    const s = v === null || v === undefined ? "" : String(v);
    // wrap in quotes + escape quotes if needed
    if (/[,"\n]/.test(s)) return `"${s.replace(/"/g, '""')}"`;
    return s;
  };

  const lines = [
    headers.join(","),
    ...rows.map((r) => headers.map((h) => esc(r[h])).join(",")),
  ];

  return lines.join("\n");
}

function downloadText(filename: string, content: string) {
  const blob = new Blob([content], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

function Card({ title, value }: { title: string; value: string }) {
  return (
    <div className="bg-[#111116] rounded-xl border border-[#222] p-5 shadow-sm">
      <p className="text-xs text-gray-400 uppercase tracking-wide">{title}</p>
      <p className="text-2xl font-semibold mt-3">{value}</p>
    </div>
  );
}

type Summary = {
  total_income: number;
  total_expense: number;
  net: number;
  currency: string;
};

export default function DashboardPage() {
  const router = useRouter();
  const [month, setMonth] = useState(() => {
    const d = new Date();
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`;
  });

  const [summary, setSummary] = useState<Summary>({
    total_income: 0,
    total_expense: 0,
    net: 0,
    currency: "INR",
  });

  const [incomes, setIncomes] = useState<any[]>([]);
  const [expenses, setExpenses] = useState<any[]>([]);

  const [error, setError] = useState("");

  useEffect(() => {
    const token =
      typeof window !== "undefined" ? localStorage.getItem("token") : null;

    if (!token) {
      router.push("/login");
    }
  }, [router]);

  async function loadSummary() {
    const token = localStorage.getItem("token");
    if (!token) return;

    try {
      setError("");
      const apiBase = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

      const res = await fetch(
        `${apiBase}/api/summary?month=${encodeURIComponent(month)}`,
        {
          headers: { Authorization: `Bearer ${token}` },
        },
      );

      if (res.status === 401) {
        localStorage.removeItem("token");
        router.push("/login");
        return;
      }

      if (!res.ok) {
        const msg = await res.text().catch(() => "");
        throw new Error(msg || `Request failed (${res.status})`);
      }

      const data = (await res.json()) as Summary;
      setSummary(data);
    } catch (err: any) {
      setError(err?.message || "Failed to load summary");
    }
  }

  async function loadIncomes() {
    const token = localStorage.getItem("token");
    if (!token) return;

    try {
      const apiBase = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      const res = await fetch(`${apiBase}/api/incomes`, {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (res.status === 401) {
        localStorage.removeItem("token");
        router.push("/login");
        return;
      }

      if (!res.ok) {
        const msg = await res.text().catch(() => "");
        throw new Error(msg || `Request failed (${res.status})`);
      }

      const data = (await res.json()) as any[];
      const filtered = month
        ? (data || []).filter(
            (x: any) => String(x.received_on).slice(0, 7) === month,
          )
        : data || [];
      setIncomes(filtered);
    } catch (err: any) {
      setError(err?.message || "Failed to load incomes");
    }
  }

  async function loadExpenses() {
    const token = localStorage.getItem("token");
    if (!token) return;

    try {
      const apiBase = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      const res = await fetch(`${apiBase}/api/expenses`, {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (res.status === 401) {
        localStorage.removeItem("token");
        router.push("/login");
        return;
      }

      if (!res.ok) {
        const msg = await res.text().catch(() => "");
        throw new Error(msg || `Request failed (${res.status})`);
      }

      const data = (await res.json()) as any[];
      const filtered = month
        ? (data || []).filter(
            (x: any) => String(x.spent_on).slice(0, 7) === month,
          )
        : data || [];
      setExpenses(filtered);
    } catch (err: any) {
      setError(err?.message || "Failed to load expenses");
    }
  }

  useEffect(() => {
    void loadSummary();
    void loadIncomes();
    void loadExpenses();
  }, [month]);

  function exportIncomesCSV() {
    const rows = (incomes || []).map((x: any) => ({
      id: x.id,
      client_name: x.client_name,
      amount: x.amount,
      currency: x.currency,
      received_on: String(x.received_on).slice(0, 10),
      note: x.note ?? "",
      created_at: x.created_at,
    }));

    const csv = toCSV(rows);
    downloadText(
      `vantro-incomes-${new Date().toISOString().slice(0, 10)}.csv`,
      csv,
    );
  }

  function exportExpensesCSV() {
    const rows = (expenses || []).map((x: any) => ({
      id: x.id,
      vendor_name: x.vendor_name,
      amount: x.amount,
      currency: x.currency,
      spent_on: String(x.spent_on).slice(0, 10),
      note: x.note ?? "",
      created_at: x.created_at,
    }));

    const csv = toCSV(rows);
    downloadText(
      `vantro-expenses-${new Date().toISOString().slice(0, 10)}.csv`,
      csv,
    );
  }

  return (
    <div className="flex min-h-screen">
      <Sidebar />

      <div className="flex flex-col flex-1">
        <Navbar />

        <main className="p-8 space-y-6">
          <h2 className="text-3xl font-bold mb-2">Overview</h2>

          {error && (
            <div className="bg-red-950/30 border border-red-900 text-red-200 rounded-lg px-4 py-3 text-sm">
              {error}
            </div>
          )}

          <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div className="flex items-center gap-3">
              <label className="text-neutral-400 text-sm">Month</label>
              <input
                type="month"
                value={month}
                onChange={(e) => setMonth(e.target.value)}
                className="bg-neutral-900 border border-neutral-800 rounded-lg px-3 py-2"
              />
            </div>

            <div className="flex gap-2">
              <button
                type="button"
                onClick={exportIncomesCSV}
                className="rounded-lg bg-neutral-800 px-3 py-2 hover:bg-neutral-700"
              >
                Export Incomes CSV
              </button>

              <button
                type="button"
                onClick={exportExpensesCSV}
                className="rounded-lg bg-neutral-800 px-3 py-2 hover:bg-neutral-700"
              >
                Export Expenses CSV
              </button>
            </div>
          </div>

          <div className="grid gap-6 grid-cols-1 md:grid-cols-3">
            <Card
              title="TOTAL INCOME"
              value={`${summary.currency} ${summary.total_income}`}
            />
            <Card
              title="TOTAL EXPENSES"
              value={`${summary.currency} ${summary.total_expense}`}
            />
            <Card title="NET" value={`${summary.currency} ${summary.net}`} />
          </div>
        </main>
      </div>
    </div>
  );
}
