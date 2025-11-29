"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Sidebar from "../components/Sidebar";
import Navbar from "../components/Navbar";

type Invoice = {
  id: number;
  customer?: string;
  amount: number;
  status?: string;
};

export default function InvoicesPage() {
  const router = useRouter();
  const [invoices, setInvoices] = useState<Invoice[]>([]);

  // auth guard (same logic as dashboard)
  useEffect(() => {
    const token =
      typeof window !== "undefined" ? localStorage.getItem("token") : null;

    if (!token) {
      router.push("/login");
      return;
    }

    // later: fetch real invoices with token
    // for now: simple placeholder data
    setInvoices([
      { id: 1, customer: "Test Customer", amount: 0, status: "Draft" },
    ]);
  }, [router]);

  return (
    <div className="flex min-h-screen">
      <Sidebar />

      <div className="flex flex-col flex-1">
        <Navbar />

        <main className="p-8 space-y-6">
          <h2 className="text-3xl font-bold mb-2">Invoices</h2>

          <div className="mt-4 rounded-xl border border-[#222] bg-[#111116]">
            <table className="w-full text-sm">
              <thead className="text-gray-400 border-b border-[#222]">
                <tr>
                  <th className="text-left px-4 py-3">#</th>
                  <th className="text-left px-4 py-3">Customer</th>
                  <th className="text-left px-4 py-3">Amount</th>
                  <th className="text-left px-4 py-3">Status</th>
                </tr>
              </thead>
              <tbody>
                {invoices.map((inv) => (
                  <tr key={inv.id} className="border-t border-[#222]">
                    <td className="px-4 py-3">#{inv.id}</td>
                    <td className="px-4 py-3">{inv.customer ?? "-"}</td>
                    <td className="px-4 py-3">{"\u20B9"}{inv.amount}</td>
                    <td className="px-4 py-3 text-gray-400">
                      {inv.status ?? "Draft"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </main>
      </div>
    </div>
  );
}
