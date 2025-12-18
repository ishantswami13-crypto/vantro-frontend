"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { apiFetch } from "@/lib/api";

type Income = {
  id: string;
  user_id: string;
  client_name: string;
  amount: number;
  currency: string;
  received_on: string; // ISO
  note?: string | null;
  created_at: string; // ISO
};

type Expense = {
  id: string;
  user_id: string;
  vendor_name: string;
  amount: number;
  currency: string;
  spent_on: string; // ISO
  note?: string | null;
  created_at: string; // ISO
};

function toYmd(d: Date) {
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function fmtINR(n: number) {
  try {
    return new Intl.NumberFormat("en-IN", {
      style: "currency",
      currency: "INR",
      maximumFractionDigits: 2,
    }).format(n || 0);
  } catch {
    return `₹${(n || 0).toFixed(2)}`;
  }
}

export default function DashboardPage() {
  const router = useRouter();

  // incomes
  const [incomes, setIncomes] = useState<Income[]>([]);
  const [incomeClient, setIncomeClient] = useState("");
  const [incomeAmount, setIncomeAmount] = useState<number>(5000);
  const [incomeDate, setIncomeDate] = useState<string>(toYmd(new Date()));
  const [incomeNote, setIncomeNote] = useState("");

  // expenses
  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [vendorName, setVendorName] = useState("");
  const [expenseAmount, setExpenseAmount] = useState<number>(1000);
  const [spentOn, setSpentOn] = useState<string>(toYmd(new Date()));
  const [expenseNote, setExpenseNote] = useState("");

  const [status, setStatus] = useState<"Live" | "Loading" | "Error">("Loading");
  const [error, setError] = useState<string>("");

  function handleLogout() {
    localStorage.removeItem("token");
    window.location.href = "/login";
  }

  async function loadAll() {
    setStatus("Loading");
    setError("");

    // if no token, bounce
    const token = localStorage.getItem("token");
    if (!token) {
      router.push("/login");
      return;
    }

    try {
      const [inc, exp] = await Promise.all([
        apiFetch<Income[]>("/api/incomes"),
        apiFetch<Expense[]>("/api/expenses"),
      ]);
      setIncomes(Array.isArray(inc) ? inc : []);
      setExpenses(Array.isArray(exp) ? exp : []);
      setStatus("Live");
    } catch (e: unknown) {
      setStatus("Error");
      setError(e instanceof Error ? e.message : "Failed to load data");
    }
  }

  useEffect(() => {
    void loadAll();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const totalIncome = useMemo(
    () =>
      Array.isArray(incomes)
        ? incomes.reduce((sum, x) => sum + (x.amount || 0), 0)
        : 0,
    [incomes]
  );

  const totalExpense = useMemo(
    () =>
      Array.isArray(expenses)
        ? expenses.reduce((sum, x) => sum + (x.amount || 0), 0)
        : 0,
    [expenses]
  );

  const net = useMemo(() => totalIncome - totalExpense, [totalIncome, totalExpense]);

  async function addIncome() {
    setError("");
    try {
      await apiFetch<{ id: string; message: string }>("/api/incomes", {
        method: "POST",
        body: JSON.stringify({
          client_name: incomeClient || "Client A",
          amount: Number(incomeAmount || 0),
          received_on: incomeDate,
          note: incomeNote || "",
        }),
      });

      setIncomeClient("");
      setIncomeNote("");
      await loadAll();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to add income");
    }
  }

  async function addExpense() {
    setError("");
    try {
      await apiFetch<{ id: string; message: string }>("/api/expenses", {
        method: "POST",
        body: JSON.stringify({
          vendor_name: vendorName || "Vendor A",
          amount: Number(expenseAmount || 0),
          spent_on: spentOn,
          note: expenseNote || "",
        }),
      });

      setVendorName("");
      setExpenseNote("");
      await loadAll();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to add expense");
    }
  }

  return (
    <div className="min-h-screen bg-black text-white">
      <div className="mx-auto max-w-6xl px-6 py-10">
        <div className="flex items-start justify-between gap-4">
          <div>
            <div className="text-sm tracking-widest text-neutral-400">VANTRO</div>
            <h1 className="text-4xl font-bold">Dashboard</h1>
          </div>
          <button
            onClick={handleLogout}
            className="rounded-lg bg-neutral-900 px-4 py-2 text-white hover:bg-neutral-800"
          >
            Logout
          </button>
        </div>

        {error ? (
          <div className="mt-6 rounded-lg border border-red-700/40 bg-red-900/20 p-4 text-red-200">
            {error}
          </div>
        ) : null}

        {/* cards */}
        <div className="mt-8 grid gap-4 md:grid-cols-4">
          <div className="rounded-2xl bg-neutral-900 p-6">
            <div className="text-sm text-neutral-400">Total Income</div>
            <div className="mt-2 text-3xl font-semibold">{fmtINR(totalIncome)}</div>
          </div>

          <div className="rounded-2xl bg-neutral-900 p-6">
            <div className="text-sm text-neutral-400">Total Expense</div>
            <div className="mt-2 text-3xl font-semibold">{fmtINR(totalExpense)}</div>
          </div>

          <div className="rounded-2xl bg-neutral-900 p-6">
            <div className="text-sm text-neutral-400">Net</div>
            <div className="mt-2 text-3xl font-semibold">{fmtINR(net)}</div>
          </div>

          <div className="rounded-2xl bg-neutral-900 p-6">
            <div className="text-sm text-neutral-400">Status</div>
            <div className="mt-2 text-3xl font-semibold">{status}</div>
          </div>
        </div>

        {/* Add income */}
        <div className="mt-8 rounded-2xl bg-neutral-900 p-6">
          <h2 className="text-2xl font-semibold">Add Income</h2>

          <div className="mt-4 grid gap-3 md:grid-cols-4">
            <input
              className="rounded-lg bg-neutral-800 px-3 py-2 outline-none"
              placeholder="Client name (e.g. Client A)"
              value={incomeClient}
              onChange={(e) => setIncomeClient(e.target.value)}
            />
            <input
              className="rounded-lg bg-neutral-800 px-3 py-2 outline-none"
              type="number"
              value={incomeAmount}
              onChange={(e) => setIncomeAmount(Number(e.target.value))}
            />
            <input
              className="rounded-lg bg-neutral-800 px-3 py-2 outline-none"
              type="date"
              value={incomeDate}
              onChange={(e) => setIncomeDate(e.target.value)}
            />
            <button
              onClick={addIncome}
              className="rounded-lg bg-white px-4 py-2 font-medium text-black hover:bg-neutral-200"
            >
              Add
            </button>
          </div>

          <input
            className="mt-3 w-full rounded-lg bg-neutral-800 px-3 py-2 outline-none"
            placeholder="Note (optional)"
            value={incomeNote}
            onChange={(e) => setIncomeNote(e.target.value)}
          />
        </div>

        {/* Add expense */}
        <div className="mt-6 rounded-2xl bg-neutral-900 p-6">
          <h2 className="text-2xl font-semibold">Add Expense</h2>

          <div className="mt-4 grid gap-3 md:grid-cols-4">
            <input
              className="rounded-lg bg-neutral-800 px-3 py-2 outline-none"
              placeholder="Vendor name (e.g. Rent, Zomato, Petrol)"
              value={vendorName}
              onChange={(e) => setVendorName(e.target.value)}
            />
            <input
              className="rounded-lg bg-neutral-800 px-3 py-2 outline-none"
              type="number"
              value={expenseAmount}
              onChange={(e) => setExpenseAmount(Number(e.target.value))}
            />
            <input
              className="rounded-lg bg-neutral-800 px-3 py-2 outline-none"
              type="date"
              value={spentOn}
              onChange={(e) => setSpentOn(e.target.value)}
            />
            <button
              onClick={addExpense}
              className="rounded-lg bg-white px-4 py-2 font-medium text-black hover:bg-neutral-200"
            >
              Add
            </button>
          </div>

          <input
            className="mt-3 w-full rounded-lg bg-neutral-800 px-3 py-2 outline-none"
            placeholder="Note (optional)"
            value={expenseNote}
            onChange={(e) => setExpenseNote(e.target.value)}
          />
        </div>

        {/* lists */}
        <div className="mt-8 grid gap-6 md:grid-cols-2">
          <div className="rounded-2xl bg-neutral-900 p-6">
            <div className="flex items-center justify-between">
              <h3 className="text-xl font-semibold">Your Incomes</h3>
              <button
                onClick={loadAll}
                className="rounded-lg bg-neutral-800 px-3 py-2 hover:bg-neutral-700"
              >
                Refresh
              </button>
            </div>

            <div className="mt-4 overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead className="text-neutral-400">
                  <tr>
                    <th className="py-2">Client</th>
                    <th className="py-2">Amount</th>
                    <th className="py-2">Received</th>
                    <th className="py-2">Note</th>
                  </tr>
                </thead>
                <tbody>
                  {incomes.length === 0 ? (
                    <tr>
                      <td className="py-4 text-neutral-400" colSpan={4}>
                        No incomes yet.
                      </td>
                    </tr>
                  ) : (
                    incomes.map((x) => (
                      <tr key={x.id} className="border-t border-neutral-800">
                        <td className="py-3">{x.client_name}</td>
                        <td className="py-3">{fmtINR(x.amount)}</td>
                        <td className="py-3">{String(x.received_on).slice(0, 10)}</td>
                        <td className="py-3 text-neutral-300">{x.note || ""}</td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>

          <div className="rounded-2xl bg-neutral-900 p-6">
            <div className="flex items-center justify-between">
              <h3 className="text-xl font-semibold">Your Expenses</h3>
              <button
                onClick={loadAll}
                className="rounded-lg bg-neutral-800 px-3 py-2 hover:bg-neutral-700"
              >
                Refresh
              </button>
            </div>

            <div className="mt-4 overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead className="text-neutral-400">
                  <tr>
                    <th className="py-2">Vendor</th>
                    <th className="py-2">Amount</th>
                    <th className="py-2">Spent</th>
                    <th className="py-2">Note</th>
                  </tr>
                </thead>
                <tbody>
                  {expenses.length === 0 ? (
                    <tr>
                      <td className="py-4 text-neutral-400" colSpan={4}>
                        No expenses yet.
                      </td>
                    </tr>
                  ) : (
                    expenses.map((x) => (
                      <tr key={x.id} className="border-t border-neutral-800">
                        <td className="py-3">{x.vendor_name}</td>
                        <td className="py-3">{fmtINR(x.amount)}</td>
                        <td className="py-3">{String(x.spent_on).slice(0, 10)}</td>
                        <td className="py-3 text-neutral-300">{x.note || ""}</td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
