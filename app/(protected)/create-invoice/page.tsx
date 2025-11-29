"use client";
import { useState } from "react";

export default function CreateInvoice() {
  const [amount, setAmount] = useState("");

  async function create(e: any) {
    e.preventDefault();

    const token = localStorage.getItem("token");

    await fetch("https://YOUR-BACKEND-URL/api/invoices/create", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ amount }),
    });

    alert("Invoice created!");
  }

  return (
    <div className="p-6 text-white">
      <h1 className="text-2xl font-bold">Create Invoice</h1>

      <form onSubmit={create} className="mt-4 space-y-4 w-60">
        <input
          className="bg-neutral-900 w-full p-2 rounded"
          placeholder="Amount"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
        />

        <button className="bg-white text-black w-full p-2 rounded">
          Create
        </button>
      </form>
    </div>
  );
}
