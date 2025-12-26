"use client";

import { useMemo, useState } from "react";
import { ArrowUpRight, RefreshCw, LogOut } from "lucide-react";
import { Card, CardContent, CardHeader } from "@/app/components/ui/Card";
import { Button } from "@/app/components/ui/Button";
import { Toast } from "@/app/components/ui/Toast";

export default function DashboardPage() {
  const [error, setError] = useState<string | null>("invalid or missing API key");

  const stats = useMemo(
    () => [
      { label: "Income", value: "₹0" },
      { label: "Expense", value: "₹0" },
      { label: "Net", value: "₹0" },
    ],
    []
  );

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

          <Button className="rounded-2xl px-5 py-3 text-sm">
            Add transaction
            <ArrowUpRight className="ml-2 h-4 w-4" />
          </Button>
        </div>

        <section className="mt-8 grid gap-4 md:grid-cols-3">
          {stats.map((s) => (
            <Card key={s.label}>
              <CardHeader>
                <div className="text-xs text-white/50">{s.label}</div>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-semibold tabular-nums tracking-tight">{s.value}</div>
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
              <div className="rounded-xl border border-white/10 bg-white/[0.02] p-4 text-sm text-white/60">
                No transactions yet.
              </div>
            </CardContent>
          </Card>
        </section>
      </main>
    </div>
  );
}
