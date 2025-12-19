"use client";

import Link from "next/link";
import { logout } from "@/lib/auth";

export default function Navbar() {
  return (
    <div className="w-full bg-neutral-950 border-b border-neutral-800">
      <div className="max-w-4xl mx-auto px-6 py-3 flex items-center justify-between">
        <Link href="/dashboard" className="font-bold text-white">
          VANTRO
        </Link>

        <div className="flex items-center gap-4">
          <Link href="/transactions" className="text-neutral-300 hover:text-white">
            Transactions
          </Link>

          <button
            onClick={logout}
            className="px-3 py-1.5 rounded bg-white text-black hover:opacity-90"
          >
            Logout
          </button>
        </div>
      </div>
    </div>
  );
}
