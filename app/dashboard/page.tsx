"use client";

import { useEffect, useMemo, useState } from "react";
import { ArrowUpRight, RefreshCw, LogOut } from "lucide-react";
import CountUp from "react-countup";
import { Card, CardContent, CardHeader } from "@/app/components/ui/Card";
import { Button } from "@/app/components/ui/Button";
import { Toast } from "@/app/components/ui/Toast";
import { TransactionRow } from "@/app/components/TransactionRow";
import { EmptyState } from "@/app/components/EmptyState";
import { Modal } from "@/app/components/ui/Modal";
import AddTransactionForm from "@/app/components/AddTransactionForm";
import { BalanceRing } from "@/app/components/BalanceRing";

export default function DashboardPage() {
  const [error, setError] = useState<string | null>("invalid or missing API key");
  const [open, setOpen] = useState(false);
  const [tx, setTx] = useState<any[]>([]);

  const stats = useMemo(
    () => [
      { label: "Income", value: 0 },
      { label: "Expense", value: 0 },
      { label: "Net", value: 0 },
    ],
    []
  );

  useEffect(() => {
    const stored = JSON.parse(localStorage.getItem("tx") || "[]");
    setTx(stored);
  }, []);

  const totals = useMemo(() => {
    const income = tx.filter((t) => t.type === "income").reduce((s, t) => s + Number(t.amount || 0), 0);
    const expense = tx.filter((t) => t.type === "expense").reduce((s, t) => s + Number(t.amount || 0), 0);
    return { income, expense, net: income - expense };
  }, [tx]);

  return (
    <div className="min-h-screen text-white">
      <Toast show={!!error}>
        <div className="flex items-center justify-between gap-3">
          <div className="text-sm text-white/90">
            <span className="font-medium">Heads up:</span>{" "}
            <span className="text-white/70">{error}</span>
          </div>
          <button className="text-xs text-white/70 hover:text-white" onClick={() => setError(null)}>
            Dismiss
          </button>
        </div>
      </Toast>

      <header className="sticky top-0 z-10 border-b border-white/10 bg-black/40 backdrop-blur-xl">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-3">
            <div className="h-9 w-9 rounded-2xl border border-white/10 bg-white/[0.06]" />
            <div className="leading-tight">
              <div className="text-sm font-semibold tracking-tight">VANTRO</div>
              <div className="text-xs text-white/50">Finance OS</div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Button variant="ghost">
              <RefreshCw className="mr-2 h-4 w-4" />
              Refresh
            </Button>
            <Button>
              <LogOut className="mr-2 h-4 w-4" />
              Logout
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-10">
        <div className="flex items-end justify-between gap-6">
          <div>
            <h1 className="text-3xl font-semibold tracking-tight">Dashboard</h1>
            <p className="mt-1 text-sm text-white/55">Your money, clean and controlled.</p>
          </div>

          <Button className="rounded-2xl px-5 py-3 text-sm" onClick={() => setOpen(true)}>
            Add transaction
            <ArrowUpRight className="ml-2 h-4 w-4" />
          </Button>
        </div>

        <section className="mt-8">
          <BalanceRing income={totals.income} expense={totals.expense} />
        </section>

        <section className="mt-8 grid gap-4 md:grid-cols-3">
          {[{ label: "Income", value: totals.income }, { label: "Expense", value: totals.expense }, { label: "Net", value: totals.net }].map((s) => (
            <Card key={s.label}>
              <CardHeader>
                <div className="text-xs text-white/50">{s.label}</div>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-semibold tabular-nums tracking-tight">
                  <CountUp end={s.value} duration={1.2} separator="," prefix="₹" />
                </div>
              </CardContent>
            </Card>
          ))}
        </section>

        <section className="mt-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <div className="text-sm font-semibold tracking-tight">Recent</div>
                <div className="text-xs text-white/50">Latest activity</div>
              </div>
              <button className="text-xs text-white/60 hover:text-white">View all →</button>
            </CardHeader>
            <CardContent>
              {tx.length === 0 ? (
                <EmptyState
                  title="No activity yet"
                  subtitle="Your transactions will appear here once you start using VANTRO."
                  action="Add your first transaction"
                  onAction={() => setOpen(true)}
                />
              ) : (
                <div className="space-y-2">
                  {tx.slice(0, 5).map((t) => (
                    <TransactionRow
                      key={t.id}
                      title={t.title}
                      category={t.category}
                      amount={t.amount}
                      type={t.type}
                    />
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </section>
      </main>

      <Modal open={open} onClose={() => setOpen(false)}>
        <div className="flex items-start justify-between gap-4">
          <div>
            <div className="text-lg font-semibold">Add transaction</div>
            <div className="text-sm text-white/50">Income or expense in 10 seconds.</div>
          </div>
          <button
            onClick={() => setOpen(false)}
            className="rounded-xl border border-white/10 bg-white/5 px-3 py-1 text-sm hover:bg-white/10 transition"
          >
            Close
          </button>
        </div>

        <div className="mt-5">
          <AddTransactionForm
            onDone={() => {
              const stored = JSON.parse(localStorage.getItem("tx") || "[]");
              setTx(stored);
              setOpen(false);
            }}
          />
        </div>
      </Modal>
    </div>
  );
}
