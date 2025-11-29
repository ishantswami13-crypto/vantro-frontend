"use client";
import { useEffect, useState } from "react";

export default function InvoicesPage() {
  const [invoices, setInvoices] = useState([]);

  useEffect(() => {
    async function load() {
      const token = localStorage.getItem("token");

      const res = await fetch("https://YOUR-BACKEND-URL/api/invoices", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      const data = await res.json();
      setInvoices(data);
    }

    load();
  }, []);

  return (
    <div className="text-white p-6">
      <h1 className="text-3xl font-bold">Invoices</h1>

      <div className="mt-4 space-y-2">
        {invoices.map((inv: any) => (
          <div key={inv.id} className="bg-neutral-900 p-4 rounded">
            <p>Invoice #{inv.id}</p>
            <p>Amount: ₹{inv.amount}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
