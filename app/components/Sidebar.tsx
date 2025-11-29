"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const menu = [
  { name: "Dashboard", path: "/dashboard" },
  { name: "Invoices", path: "/invoices" },
  { name: "Create Invoice", path: "/create-invoice" },
  { name: "Products", path: "/products" },
  { name: "Expenses", path: "/expenses" },
];

export default function Sidebar() {
  const path = usePathname();

  return (
    <aside className="w-60 h-screen bg-[#08080c] border-r border-[#222] flex flex-col">
      {/* Brand */}
      <div className="px-6 py-5 border-b border-[#222]">
        <h1 className="text-xl font-extrabold tracking-wide">
          <span className="text-sky-400">VANTRO</span>
        </h1>
        <p className="text-[11px] text-gray-500 mt-1 uppercase tracking-wide">
          Dashboard
        </p>
      </div>

      {/* Menu */}
      <nav className="flex-1 px-3 py-4 space-y-1 text-sm">
        {menu.map((item) => {
          const active = path === item.path;

          return (
            <Link
              key={item.path}
              href={item.path}
              className={`block rounded-lg px-3 py-2 transition-colors ${
                active
                  ? "bg-[#10101a] text-sky-400"
                  : "text-gray-400 hover:text-white hover:bg-[#11111a]"
              }`}
            >
              {item.name}
            </Link>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="px-6 py-4 text-xs text-gray-500 border-t border-[#222]">
        {"\u00A9"} {new Date().getFullYear()} VANTRO
      </div>
    </aside>
  );
}
